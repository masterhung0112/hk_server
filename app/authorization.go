package app

import (
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/model"
)

func (a *App) HasPermissionTo(askingUserId string, permission *model.Permission) bool {
	user, err := a.GetUser(askingUserId)
	if err != nil {
		return false
	}

	roles := user.GetRoles()

	return a.RolesGrantPermission(roles, permission.Id)
}

func (a *App) HasPermissionToTeam(askingUserId string, teamId string, permission *model.Permission) bool {
	if teamId == "" || askingUserId == "" {
		return false
	}
	teamMember, _ := a.GetTeamMember(teamId, askingUserId)
	if teamMember != nil && teamMember.DeleteAt == 0 {
		if a.RolesGrantPermission(teamMember.GetRoles(), permission.Id) {
			return true
		}
	}
	return a.HasPermissionTo(askingUserId, permission)
}

func (a *App) RolesGrantPermission(roleNames []string, permissionId string) bool {
	roles, err := a.GetRolesByNames(roleNames)
	if err != nil {
		// This should only happen if something is very broken. We can't realistically
		// recover the situation, so deny permission and log an error.
		mlog.Error("Failed to get roles from database with role names: "+strings.Join(roleNames, ",")+" ", mlog.Err(err))
		return false
	}

	for _, role := range roles {
		if role.DeleteAt != 0 {
			continue
		}

		permissions := role.Permissions
		for _, permission := range permissions {
			if permission == permissionId {
				return true
			}
		}
	}

	return false
}