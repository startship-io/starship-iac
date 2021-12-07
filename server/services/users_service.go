package users_service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/starship-cloud/starship-iac/server/core/db"
	"github.com/starship-cloud/starship-iac/server/events/models"
	"github.com/starship-cloud/starship-iac/utils"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

const (
	DB_NAME       = "starship"
	DB_COLLECTION = "users"
)

func GetUser(userId string, db *db.MongoDB) (*models.UserEntity, error) {
	collection := db.DBClient.Database(DB_NAME).Collection(DB_COLLECTION)

	filter := bson.M{"userid": userId}

	userEntity := &models.UserEntity{}
	err := db.GetOne(collection, filter, &userEntity)

	if err != nil {
		return nil, fmt.Errorf("get user with user id %s failed due to DB operation", userId)
	} else if userEntity.Userid != "" {
		return userEntity, nil
	} else {
		//not found
		return nil, nil
	}
}

func CreateUser(user *models.UserEntity, db *db.MongoDB) (*models.UserEntity, error) {
	collection := db.DBClient.Database(DB_NAME).Collection(DB_COLLECTION)
	userEntity := &models.UserEntity{}

	filter := bson.M{"username": user.Username}
	db.GetOne(collection, filter, userEntity)

	if userEntity.Userid != "" {
		return nil, fmt.Errorf("the user %s with userId %s has been exist.", user.Username, userEntity.Userid)
	} else {
		userId := utils.GenUserId()
		t := time.Now().Unix()

		newUser := &models.UserEntity{
			Userid:    userId,
			Username:  user.Username,
			Email:     user.Email,
			Password:  user.Password,
			UserLocal: true,
			CreateAt:  t,
			UpdateAt:  t,
		}

		_, err := db.Insert(collection, newUser)
		if err != nil {
			return nil, fmt.Errorf("save user %s failed due to DB operation", user.Username)
		} else {
			return newUser, nil
		}
	}
}

func DeleteUser(user *models.UserEntity, db *db.MongoDB) (*models.UserEntity, error) {
	if len(strings.TrimSpace(user.Userid)) == 0 {
		return nil, errors.New("userid is required.")
	}

	collection := db.DBClient.Database(DB_NAME).Collection(DB_COLLECTION)

	userEntity := &models.UserEntity{}
	filter := bson.M{"userid": user.Userid}
	err := db.GetOne(collection, filter, userEntity)

	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	} else if userEntity.Userid != "" {
		//found, delete
		_, err := db.Delete(collection, filter)
		return nil, err
	} else {
		return nil, fmt.Errorf("the user with user id: %s has been deleted.", user.Userid)
	}

}

func UpdateUser(user *models.UserEntity, db *db.MongoDB) (*models.UserEntity, error) {
	if len(strings.TrimSpace(user.Userid)) == 0 ||
		len(strings.TrimSpace(user.Username)) == 0 ||
		len(strings.TrimSpace(user.Email)) == 0 {
		return nil, errors.New("userid/username/email are required.")
	}

	collection := db.DBClient.Database(DB_NAME).Collection(DB_COLLECTION)
	userEntity := &models.UserEntity{}
	filter := bson.M{"userid": user.Userid}

	db.GetOne(collection, filter, &userEntity)

	if userEntity.Userid != "" {
		//found
		newUser := &models.UserEntity{
			Userid:   userEntity.Userid,
			Username: user.Username,
			Email:    user.Email,
			Password: userEntity.Password, //can't be updated
			CreateAt: time.Now().Unix(),
		}

		_, err := db.UpdateOrSave(collection, newUser, bson.M{})
		if err != nil {
			return nil, fmt.Errorf("update user %s failed due to DB operation", user.Username)
		} else {
			return newUser, nil
		}
	} else {
		return nil, fmt.Errorf("the user %s with user id %s not exist.", userEntity.Username, user.Userid)
	}
}

func SearchUsers(userName string, db *db.MongoDB, pageinOpt *models.PaginOption) ([]models.UserEntity, error) {
	collection := db.DBClient.Database(DB_NAME).Collection(DB_COLLECTION)
	var users []models.UserEntity
	filter := bson.M{
		"username": bson.M{
			"$regex":   userName,
			"$options": "i",
		},
	}

	db.GetList(collection, filter, &users, *pageinOpt)

	if len(users) == 0 {
		return nil, fmt.Errorf("get user %s failed due to DB operation", userName)
	} else {
		return users, nil
	}
}
