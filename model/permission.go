package model

const (
	PermissionScopeSystem  = "system_scope"
	PermissionScopeTeam    = "team_scope"
	PermissionScopeChannel = "channel_scope"
)

type Permission struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
}

var AllPermissions []*Permission

var PERMISSION_LIST_USERS_WITHOUT_TEAM *Permission

func initializePermissions() {
  PERMISSION_LIST_USERS_WITHOUT_TEAM = &Permission{
		"list_users_without_team",
		"authentication.permissions.list_users_without_team.name",
		"authentication.permissions.list_users_without_team.description",
		PermissionScopeSystem,
  }

  SystemScopedPermissionsMinusSysconsole := []*Permission{
    PERMISSION_LIST_USERS_WITHOUT_TEAM,
  }

  AllPermissions = []*Permission{}
  AllPermissions = append(AllPermissions, SystemScopedPermissionsMinusSysconsole...)
}

func init() {
  initializePermissions()
}