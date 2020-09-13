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

var PERMISSION_INVITE_USER *Permission
var PERMISSION_ADD_USER_TO_TEAM *Permission

var PERMISSION_CREATE_PUBLIC_CHANNEL *Permission
var PERMISSION_CREATE_PRIVATE_CHANNEL *Permission
var PERMISSION_DELETE_PUBLIC_CHANNEL *Permission
var PERMISSION_DELETE_PRIVATE_CHANNEL *Permission

var PERMISSION_CREATE_POST *Permission
var PERMISSION_LIST_USERS_WITHOUT_TEAM *Permission
var PERMISSION_USE_CHANNEL_MENTIONS *Permission
var PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS *Permission
var PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS *Permission
var PERMISSION_VIEW_MEMBERS *Permission
var PERMISSION_EDIT_OTHER_USERS *Permission
var PERMISSION_READ_CHANNEL *Permission

var PERMISSION_MANAGE_PUBLIC_CHANNEL_PROPERTIES *Permission
var PERMISSION_MANAGE_PRIVATE_CHANNEL_PROPERTIES *Permission

var PERMISSION_VIEW_TEAM *Permission

var PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS *Permission
var PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_GROUPS *Permission

// General permission that encompasses all system admin functions
// in the future this could be broken up to allow access to some
// admin functions but not others
var PERMISSION_MANAGE_SYSTEM *Permission

var PERMISSION_EDIT_POST *Permission
var PERMISSION_EDIT_OTHERS_POSTS *Permission
var PERMISSION_DELETE_POST *Permission
var PERMISSION_DELETE_OTHERS_POSTS *Permission

var SysconsoleReadPermissions []*Permission
var SysconsoleWritePermissions []*Permission

func initializePermissions() {
	PERMISSION_INVITE_USER = &Permission{
		"invite_user",
		"authentication.permissions.team_invite_user.name",
		"authentication.permissions.team_invite_user.description",
		PermissionScopeTeam,
	}
	PERMISSION_ADD_USER_TO_TEAM = &Permission{
		"add_user_to_team",
		"authentication.permissions.add_user_to_team.name",
		"authentication.permissions.add_user_to_team.description",
		PermissionScopeTeam,
	}
	PERMISSION_CREATE_POST = &Permission{
		"create_post",
		"authentication.permissions.create_post.name",
		"authentication.permissions.create_post.description",
		PermissionScopeChannel,
	}
	PERMISSION_CREATE_PUBLIC_CHANNEL = &Permission{
		"create_public_channel",
		"authentication.permissions.create_public_channel.name",
		"authentication.permissions.create_public_channel.description",
		PermissionScopeTeam,
	}
	PERMISSION_CREATE_PRIVATE_CHANNEL = &Permission{
		"create_private_channel",
		"authentication.permissions.create_private_channel.name",
		"authentication.permissions.create_private_channel.description",
		PermissionScopeTeam,
	}
	PERMISSION_DELETE_PUBLIC_CHANNEL = &Permission{
		"delete_public_channel",
		"authentication.permissions.delete_public_channel.name",
		"authentication.permissions.delete_public_channel.description",
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
	PERMISSION_EDIT_OTHER_USERS = &Permission{
		"edit_other_users",
		"authentication.permissions.edit_other_users.name",
		"authentication.permissions.edit_other_users.description",
		PermissionScopeSystem,
	}
	PERMISSION_READ_CHANNEL = &Permission{
		"read_channel",
		"authentication.permissions.read_channel.name",
		"authentication.permissions.read_channel.description",
		PermissionScopeChannel,
	}
	PERMISSION_MANAGE_SYSTEM = &Permission{
		"manage_system",
		"authentication.permissions.manage_system.name",
		"authentication.permissions.manage_system.description",
		PermissionScopeSystem,
	}
	PERMISSION_VIEW_TEAM = &Permission{
		"view_team",
		"authentication.permissions.view_team.name",
		"authentication.permissions.view_team.description",
		PermissionScopeTeam,
	}

	PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS = &Permission{
		"sysconsole_read_user_management_groups",
		"authentication.permissions.use_group_mentions.name",
		"authentication.permissions.use_group_mentions.description",
		PermissionScopeSystem,
	}
	PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_GROUPS = &Permission{
		"sysconsole_write_user_management_groups",
		"authentication.permissions.use_group_mentions.name",
		"authentication.permissions.use_group_mentions.description",
		PermissionScopeSystem,
	}
	PERMISSION_MANAGE_PUBLIC_CHANNEL_PROPERTIES = &Permission{
		"manage_public_channel_properties",
		"authentication.permissions.manage_public_channel_properties.name",
		"authentication.permissions.manage_public_channel_properties.description",
		PermissionScopeChannel,
	}
	PERMISSION_MANAGE_PRIVATE_CHANNEL_PROPERTIES = &Permission{
		"manage_private_channel_properties",
		"authentication.permissions.manage_private_channel_properties.name",
		"authentication.permissions.manage_private_channel_properties.description",
		PermissionScopeChannel,
	}
	PERMISSION_DELETE_PUBLIC_CHANNEL = &Permission{
		"delete_public_channel",
		"authentication.permissions.delete_public_channel.name",
		"authentication.permissions.delete_public_channel.description",
		PermissionScopeChannel,
	}
	PERMISSION_DELETE_PRIVATE_CHANNEL = &Permission{
		"delete_private_channel",
		"authentication.permissions.delete_private_channel.name",
		"authentication.permissions.delete_private_channel.description",
		PermissionScopeChannel,
	}
	PERMISSION_EDIT_POST = &Permission{
		"edit_post",
		"authentication.permissions.edit_post.name",
		"authentication.permissions.edit_post.description",
		PermissionScopeChannel,
	}
	PERMISSION_EDIT_OTHERS_POSTS = &Permission{
		"edit_others_posts",
		"authentication.permissions.edit_others_posts.name",
		"authentication.permissions.edit_others_posts.description",
		PermissionScopeChannel,
	}
	PERMISSION_DELETE_POST = &Permission{
		"delete_post",
		"authentication.permissions.delete_post.name",
		"authentication.permissions.delete_post.description",
		PermissionScopeChannel,
	}
	PERMISSION_DELETE_OTHERS_POSTS = &Permission{
		"delete_others_posts",
		"authentication.permissions.delete_others_posts.name",
		"authentication.permissions.delete_others_posts.description",
		PermissionScopeChannel,
	}

	SystemScopedPermissionsMinusSysconsole := []*Permission{
		PERMISSION_MANAGE_SYSTEM,
		PERMISSION_LIST_USERS_WITHOUT_TEAM,
		PERMISSION_EDIT_OTHER_USERS,
	}

	TeamScopedPermissions := []*Permission{
		PERMISSION_INVITE_USER,
		PERMISSION_ADD_USER_TO_TEAM,
		PERMISSION_CREATE_PUBLIC_CHANNEL,
		PERMISSION_CREATE_PRIVATE_CHANNEL,
		PERMISSION_VIEW_MEMBERS,
		PERMISSION_VIEW_TEAM,
	}

	ChannelScopedPermissions := []*Permission{
		PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS,
		PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS,
		PERMISSION_MANAGE_PUBLIC_CHANNEL_PROPERTIES,
		PERMISSION_MANAGE_PRIVATE_CHANNEL_PROPERTIES,
		PERMISSION_DELETE_PUBLIC_CHANNEL,
		PERMISSION_DELETE_PRIVATE_CHANNEL,
		PERMISSION_READ_CHANNEL,
		PERMISSION_CREATE_POST,
		PERMISSION_EDIT_POST,
		PERMISSION_EDIT_OTHERS_POSTS,
		PERMISSION_DELETE_POST,
		PERMISSION_DELETE_OTHERS_POSTS,
		PERMISSION_USE_CHANNEL_MENTIONS,
	}

	SysconsoleReadPermissions = []*Permission{
		PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS,
	}
	SysconsoleWritePermissions = []*Permission{
		PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_GROUPS,
	}

	AllPermissions = []*Permission{}
	AllPermissions = append(AllPermissions, SystemScopedPermissionsMinusSysconsole...)
	AllPermissions = append(AllPermissions, TeamScopedPermissions...)
	AllPermissions = append(AllPermissions, ChannelScopedPermissions...)
	AllPermissions = append(AllPermissions, SysconsoleReadPermissions...)
	AllPermissions = append(AllPermissions, SysconsoleWritePermissions...)

	ChannelModeratedPermissions = []string{
		PERMISSION_CREATE_POST.Id,
		"create_reactions",
		"manage_members",
		PERMISSION_USE_CHANNEL_MENTIONS.Id,
	}

	ChannelModeratedPermissionsMap = map[string]string{
		PERMISSION_CREATE_POST.Id: ChannelModeratedPermissions[0],
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
