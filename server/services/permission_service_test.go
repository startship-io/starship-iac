package service

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"github.com/starship-cloud/starship-iac/server/core/db"
	"github.com/starship-cloud/starship-iac/server/events/models"
	"github.com/starship-cloud/starship-iac/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"testing"
)

func NewEnforcer() *casbin.Enforcer {
	uri := utils.MongoDBConnectionUri
	if !strings.HasPrefix(uri, "mongodb+srv://") && !strings.HasPrefix(uri, "mongodb://") {
		uri = fmt.Sprint("mongodb://" + uri)
	}

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

	a, err := mongodbadapter.NewAdapterWithClientOption(clientOptions, utils.MongoAuthDBName)
	if err != nil {
		panic(err)
	}

	m := model.NewModel()
	m.LoadModelFromText(`
			[request_definition]
			r = sub, obj, act
			
			[policy_definition]
			p = sub, obj, act
			
			[role_definition]
			g = _, _
			
			[policy_effect]
			e = some(where (p.eft == allow))
			
			[matchers]
			m = g(r.sub, p.sub) && ( r.obj == p.obj || p.obj=="*" ) && ( r.act == p.act || p.act=="*" )
			`)

	e, err := casbin.NewEnforcer(m, a)
	e.EnableAutoSave(true)
	if err != nil {
		panic(err)
	}
	return e
}

func Test_CreateRole(t *testing.T) {
	e := NewEnforcer()

	admin_role := &models.Role{
		RoleName:   "admin",
		Id:         "*",
		Permission: "*",
	}
	res, err := CreateRole(admin_role, e)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(res)
	}

	projectCreater := &models.Role{
		RoleName:   "projectCreater",
		Id:         "*",
		Permission: "project_create",
	}
	res, err = CreateRole(projectCreater, e)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(res)
	}

	security := &models.Role{
		RoleName:   "security",
		Id:         "*",
		Permission: "secret",
	}
	res, err = CreateRole(security, e)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(res)
	}
}

func Test_RoleForUser(t *testing.T) {
	e := NewEnforcer()

	userId := "zs"
	adminRole := &models.RoleForUser{UserId: userId, RoleName: utils.Admin}
	res, _ := AddRoleForUser(adminRole, e)
	fmt.Println(res)

	res, _ = DeleteRoleForUser(adminRole, e)
	fmt.Println(res)

	projCreateRole := &models.RoleForUser{UserId: userId, RoleName: utils.ProjectCreater}
	res, _ = AddRoleForUser(projCreateRole, e)
	fmt.Println(res)

	//add admin role again
	res, _ = AddRoleForUser(adminRole, e)
	fmt.Println(res)

	role, _ := GetRoleForUser(userId, e)
	fmt.Println(role)
}

func Test_ProjectPermissionsForUser(t *testing.T) {
	e := NewEnforcer()
	userId := "zs"
	projectId := "proj1"
	permission1 := &models.ProjectPermission{
		UserId:     userId,
		ProjectId:  projectId,
		Permission: utils.ReadOnly,
	}
	res, _ := AddProjectPermissionForUser(permission1, e)
	fmt.Println(res)

	permission2 := &models.ProjectPermission{
		UserId:     userId,
		ProjectId:  projectId,
		Permission: utils.Config,
	}
	res, _ = AddProjectPermissionForUser(permission2, e)
	fmt.Println(res)

	res, _ = DeleteProjectPermissionForUser(permission2, e)
	fmt.Println(res)

}

func Test_GetAllProjectPermissionsForUser(t *testing.T) {
	e := NewEnforcer()

	userId := "zs"
	res := GetAllProjectPermissionsForUser(userId, e)
	fmt.Println(res)

}

func Test_GetUsersByProjectId(t *testing.T) {
	e := NewEnforcer()

	projectId := "proj1"
	res := GetUsersByProjectId(projectId, e)
	fmt.Println(res)

}
