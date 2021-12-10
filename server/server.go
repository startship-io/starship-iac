package server

import (
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"github.com/kataras/iris/v12"
	"github.com/starship-cloud/starship-iac/server/controller"
	"github.com/starship-cloud/starship-iac/server/core/db"
	"github.com/starship-cloud/starship-iac/server/events"
	"github.com/starship-cloud/starship-iac/server/logging"
	"github.com/starship-cloud/starship-iac/utils"
	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Port                 int
	Logger               logging.SimpleLogging
	App                  *iris.Application
	StatusController     *controllers.StatusController
	UsersController      *controllers.UsersController
	AdminController      *controllers.AdminController
	AuthController       *controllers.AuthController
	PermissionController *controllers.PermissionController

	SSLCertFile       string
	SSLKeyFile        string
	SSLPort           int
	SkipAuthToken     bool
	Drainer           *events.Drainer
	WebAuthentication bool
	WebUsername       string
	WebPassword       string
}

type Config struct {
	AllowForkPRsFlag        string
	StarshipURLFlag         string
	StarshipVersion         string
	DefaultTFVersionFlag    string
	RepoConfigJSONFlag      string
	SilenceForkPRErrorsFlag string
}

func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	logger, err := logging.NewStructuredLoggerFromLevel(userConfig.ToLogLevel())

	if err != nil {
		return nil, err
	}

	drainer := &events.Drainer{}
	db, err := db.NewDB(&db.DBConfig{
		MongoDBConnectionUri: userConfig.MongoDBConnectionUri,
		MongoDBName:          userConfig.MongoDBName,
		MongoDBUserName:      userConfig.MongoDBUserName,
		MongoDBPassword:      userConfig.MongoDBPassword,
		MaxConnection:        userConfig.MaxConnection,
		RootCmdLogPath:       userConfig.RootCmdLogPath,
		RootSecret:           userConfig.RootSecret,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize db.")
	}

	userController := &controllers.UsersController{
		Logger:  logger,
		Drainer: drainer,
		DB:      db,
	}

	permissionController, err := initPermissionSystem(logger, drainer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize permission system.")
	}

	app := iris.New()

	return &Server{
		Port:                 userConfig.Port,
		Logger:               logger,
		SSLKeyFile:           userConfig.SSLKeyFile,
		SSLCertFile:          userConfig.SSLCertFile,
		SkipAuthToken:        userConfig.SkipAuthToken,
		Drainer:              drainer,
		UsersController:      userController,
		PermissionController: permissionController,
		App:                  app,
	}, nil
}

func initPermissionSystem(logger logging.SimpleLogging, drainer *events.Drainer) (*controllers.PermissionController, error) {
	dbConfig := db.DBConfig{
		MongoDBConnectionUri: utils.MongoDBConnectionUri,
		MongoDBName:          utils.MongoAuthDBName,
		MongoDBUserName:      utils.MongoDBUserName,
		MongoDBPassword:      utils.MongoDBPassword,
		MaxConnection:        utils.MaxConnection,
		RootCmdLogPath:       utils.RootCmdLogPath,
		RootSecret:           utils.RootSecret,
	}
	clientOptions := options.Client().ApplyURI(dbConfig.MongoDBConnectionUri)
	clientOptions.SetMaxPoolSize(uint64(dbConfig.MaxConnection))
	credential := options.Credential{
		Username: dbConfig.MongoDBUserName,
		Password: dbConfig.MongoDBPassword,
	}

	clientOptions.SetAuth(credential)

	adapter, err := mongodbadapter.NewAdapterWithClientOption(clientOptions, utils.MongoAuthDBName)
	if err != nil {
		return nil, err
	}

	rbacModel := model.NewModel()
	rbacModel.AddDef("r", "r", "sub, obj, act")
	rbacModel.AddDef("p", "p", "sub, obj, act")
	rbacModel.AddDef("e", "e", "some(where (p.eft == allow))")
	rbacModel.AddDef("m", "m", "m = g(r.sub, p.sub) && ( r.obj == p.obj || p.obj==\"*\" ) && ( r.act == p.act || p.act==\"*\" )")

	enforcer, err := casbin.NewEnforcer(rbacModel, adapter)
	if err != nil {
		return nil, err
	}
	enforcer.EnableAutoSave(true)
	return &controllers.PermissionController{
		Logger:   logger,
		Drainer:  drainer,
		Enforcer: enforcer,
	}, nil
}

func (s *Server) ControllersInitialize() {
	apiVer := "/api/v1"
	s.App.Get(apiVer+"/status", s.StatusController.Status)

	s.App.Get(apiVer+"/users/{userId:string}", s.UsersController.Get)
	s.App.Post(apiVer+"/users/create", s.UsersController.Create)
	s.App.Post(apiVer+"/users/delete", s.UsersController.Delete)
	s.App.Get(apiVer+"/users/search", s.UsersController.Search)

	s.App.Get(apiVer+"/admin/users", s.AdminController.Users)
	s.App.Post(apiVer+"/login", s.AuthController.Login)
}

func (s *Server) Start() error {
	defer s.Logger.Flush()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	s.ControllersInitialize()

	go func() {
		s.Logger.Info("Starship-IaC started - listening on port %v", s.Port)

		var err error
		if !s.SkipAuthToken {
			s.App.UseGlobal(checkToken)
		} else {
			s.Logger.Warn("auth token was skipped *** dangious and only can be used in testing/developing phase")
		}

		if s.SSLCertFile != "" && s.SSLKeyFile != "" {
			port := fmt.Sprint(":", s.SSLPort)
			err = s.App.Run(iris.TLS(port, s.SSLCertFile, s.SSLKeyFile))
		} else {
			port := fmt.Sprint(":", s.Port)
			err = s.App.Run(iris.Addr(port))
		}

		if err != nil {
			fmt.Println(err.Error())
		}

	}()
	<-stop

	s.Logger.Warn("Received interrupt. Waiting for in-progress operations to complete")
	s.waitForDrain()
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint: vet

	if err := s.App.Shutdown(ctx); err != nil {
		return cli.NewExitError(fmt.Sprintf("while shutting down: %s", err), 1)
	}
	return nil
}

func (s *Server) waitForDrain() {
	drainComplete := make(chan bool, 1)
	go func() {
		s.Drainer.ShutdownBlocking()
		drainComplete <- true
	}()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-drainComplete:
			s.Logger.Info("All in-progress operations complete, shutting down")
			return
		case <-ticker.C:
			s.Logger.Info("Waiting for in-progress operations to complete, current in-progress ops: %d", s.Drainer.GetStatus().InProgressOps)
		}
	}
}
