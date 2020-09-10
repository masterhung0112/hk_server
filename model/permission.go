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

var ChannelModeratedPermissions []string
var ChannelModeratedPermissionsMap map[string]string

var PERMISSION_CREATE_POST *Permission
var PERMISSION_LIST_USERS_WITHOUT_TEAM *Permission
var PERMISSION_USE_CHANNEL_MENTIONS *Permission
var PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS *Permission
var PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS *Permission
var PERMISSION_VIEW_MEMBERS *Permission

func initializePermissions() {
  PERMISSION_CREATE_POST = &Permission{
		"create_post",
		"authentication.permissions.create_post.name",
		"authentication.permissions.create_post.description",
		PermissionScopeChannel,
  }
  PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS = &Permission{
		"manage_public_channel_members",
		"authentication.permissions.manage_public_channel_members.name",
		"authentication.permissions.manage_public_channel_members.description",
		PermissionScopeChannel,
	}
  PERMISSION_LIST_USERS_WITHOUT_TEAM = &Permission{
		"list_users_without_team",
		"authentication.permissions.list_users_without_team.name",
		"authentication.permissions.list_users_without_team.description",
		PermissionScopeSystem,
  }
  PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS = &Permission{
		"manage_private_channel_members",
		"authentication.permissions.manage_private_channel_members.name",
		"authentication.permissions.manage_private_channel_members.description",
		PermissionScopeChannel,
  }
  PERMISSION_USE_CHANNEL_MENTIONS = &Permission{
		"use_channel_mentions",
		"authentication.permissions.use_channel_mentions.name",
		"authentication.permissions.use_channel_mentions.description",
		PermissionScopeChannel,
  }
  PERMISSION_VIEW_MEMBERS = &Permission{
		"view_members",
		"authentication.permisssions.view_members.name",
		"authentication.permisssions.view_members.description",
		PermissionScopeTeam,
	}

  SystemScopedPermissionsMinusSysconsole := []*Permission{
    PERMISSION_LIST_USERS_WITHOUT_TEAM,
  }

  TeamScopedPermissions := []*Permission{
		PERMISSION_VIEW_MEMBERS,
  }

  AllPermissions = []*Permission{}
  AllPermissions = append(AllPermissions, SystemScopedPermissionsMinusSysconsole...)
	AllPermissions = append(AllPermissions, TeamScopedPermissions...)

  ChannelModeratedPermissions = []string{
		PERMISSION_CREATE_POST.Id,
		"create_reactions",
		"manage_members",
		PERMISSION_USE_CHANNEL_MENTIONS.Id,
	}

  ChannelModeratedPermissionsMap = map[string]string{
		PERMISSION_CREATE_POST.Id:                    ChannelModeratedPermissions[0],
		// PERMISSION_ADD_REACTION.Id:                   ChannelModeratedPermissions[1],
		// PERMISSION_REMOVE_REACTION.Id:                ChannelModeratedPermissions[1],
		PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id:  ChannelModeratedPermissions[2],
		PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS.Id: ChannelModeratedPermissions[2],
		PERMISSION_USE_CHANNEL_MENTIONS.Id:           ChannelModeratedPermissions[3],
	}
}

func init() {
  initializePermissions()
}