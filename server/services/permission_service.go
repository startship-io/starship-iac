package service

import (
	"github.com/casbin/casbin/v2"
	"github.com/deckarep/golang-set"
	"github.com/starship-cloud/starship-iac/server/events/models"
	"github.com/starship-cloud/starship-iac/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateRole(role *models.Role, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddPolicy(role.RoleName, role.Id, role.Permission)
}

func AddRoleForUser(roleForUser *models.RoleForUser, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddRoleForUser(roleForUser.UserId, roleForUser.RoleName)
}

func DeleteRoleForUser(roleForUser *models.RoleForUser, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.DeleteRoleForUser(roleForUser.UserId, roleForUser.RoleName)
}

func GetRoleForUser(userId string, enforcer *casbin.Enforcer) ([]string, error) {
	return enforcer.GetRolesForUser(userId)
}

func AddProjectPermissionForUser(permission *models.ProjectPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddPolicy(permission.Id, permission.ProjectId, permission.Permission)
}

func DeleteProjectPermissionForUser(permission *models.ProjectPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemovePolicy(permission.Id, permission.ProjectId, permission.Permission)
}

func GetProjectIdsForUser(userId string, enforcer *casbin.Enforcer) mapset.Set {
	filter := &bson.M{"v2": bson.M{"$in": []string{utils.ReadOnly, utils.Config}}}
	enforcer.LoadFilteredPolicy(filter)
	projectPermissions := enforcer.GetFilteredPolicy(0, userId)
	enforcer.LoadPolicy()
	projectIds := mapset.NewSet()
	for _, projectPermission := range projectPermissions {
		projectIds.Add(projectPermission[1])
	}
	return projectIds
}

//user id and group id
func GetUserIdsForProject(projectId string, enforcer *casbin.Enforcer) mapset.Set {
	projectPermissions := enforcer.GetFilteredPolicy(1, projectId)
	userIds := mapset.NewSet()
	for _, projectPermission := range projectPermissions {
		userIds.Add(projectPermission[0])
	}
	return userIds
}

func AddProjectPermissionForGroup(permission *models.ProjectPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddGroupingPolicy(permission.Id, permission.ProjectId, permission.Permission)
}

func DeleteProjectPermissionForGroup(permission *models.ProjectPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemoveGroupingPolicy(permission.Id, permission.ProjectId, permission.Permission)
}

func GetAllProjectPermissionsForGroup(groupId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredGroupingPolicy(0, groupId)
}

func AddEnvironmentPermissionForUser(permission *models.EnvironmentPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddPolicy(permission.Id, permission.EnvironmentId, permission.Permission)
}

func DeleteEnvironmentPermissionForUser(permission *models.EnvironmentPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemovePolicy(permission.Id, permission.EnvironmentId, permission.Permission)
}

func GetAllEnvironmentPermissionsForUser(userId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredPolicy(0, userId)
}

func AddEnvironmentPermissionForGroup(permission *models.EnvironmentPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddPolicy(permission.Id, permission.EnvironmentId, permission.Permission)
}

func DeleteEnvironmentPermissionForGroup(permission *models.EnvironmentPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemovePolicy(permission.Id, permission.EnvironmentId, permission.Permission)
}

func GetAllEnvironmentPermissionsForGroup(groupId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredPolicy(0, groupId)
}

func AddConfigurationPermissionForUser(permission *models.ConfigurationPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddPolicy(permission.Id, permission.ConfigurationId, permission.Permission)
}

func DeleteConfigurationPermissionForUser(permission *models.ConfigurationPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemovePolicy(permission.Id, permission.ConfigurationId, permission.Permission)
}

func GetAllConfigurationPermissionsForUser(userId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredPolicy(0, userId)
}

func AddConfigurationPermissionForGroup(permission *models.ConfigurationPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddGroupingPolicy(permission.Id, permission.ConfigurationId, permission.Permission)
}

func DeleteConfigurationPermissionForGroup(permission *models.ConfigurationPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemoveGroupingPolicy(permission.Id, permission.ConfigurationId, permission.Permission)
}

func GetAllConfigurationPermissionsForGroup(groupId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredGroupingPolicy(0, groupId)
}

func AddSecretPermissionForUser(permission *models.SecretPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddPolicy(permission.Id, permission.SecretId, permission.Permission)
}

func DeleteSecretPermissionForUser(permission *models.SecretPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemovePolicy(permission.Id, permission.SecretId, permission.Permission)
}

func GetAllSecretPermissionsForUser(userId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredPolicy(0, userId)
}

func AddSecretPermissionForGroup(permission *models.SecretPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.AddGroupingPolicy(permission.Id, permission.SecretId, permission.Permission)
}

func DeleteSecretPermissionForGroup(permission *models.SecretPermission, enforcer *casbin.Enforcer) (bool, error) {
	return enforcer.RemoveGroupingPolicy(permission.Id, permission.SecretId, permission.Permission)
}

func GetAllSecretPermissionsForGroup(groupId string, enforcer *casbin.Enforcer) [][]string {
	return enforcer.GetFilteredGroupingPolicy(0, groupId)
}
