package model

import (
	"fmt"
	"strings"
)

const (
	SYSTEM_GUEST_ROLE_ID             = "system_guest"
	SYSTEM_USER_ROLE_ID              = "system_user"
	SYSTEM_ADMIN_ROLE_ID             = "system_admin"
	SYSTEM_POST_ALL_ROLE_ID          = "system_post_all"
	SYSTEM_POST_ALL_PUBLIC_ROLE_ID   = "system_post_all_public"
	SYSTEM_USER_ACCESS_TOKEN_ROLE_ID = "system_user_access_token"
	SYSTEM_USER_MANAGER_ROLE_ID      = "system_user_manager"
	SYSTEM_READ_ONLY_ADMIN_ROLE_ID   = "system_read_only_admin"
	SYSTEM_MANAGER_ROLE_ID           = "system_manager"

	ROLE_NAME_MAX_LENGTH         = 64
	ROLE_DISPLAY_NAME_MAX_LENGTH = 128
	ROLE_DESCRIPTION_MAX_LENGTH  = 1024

	CHANNEL_GUEST_ROLE_ID = "channel_guest"
	CHANNEL_USER_ROLE_ID  = "channel_user"
	CHANNEL_ADMIN_ROLE_ID = "channel_admin"

	TEAM_GUEST_ROLE_ID           = "team_guest"
	TEAM_USER_ROLE_ID            = "team_user"
	TEAM_ADMIN_ROLE_ID           = "team_admin"
	TEAM_POST_ALL_ROLE_ID        = "team_post_all"
	TEAM_POST_ALL_PUBLIC_ROLE_ID = "team_post_all_public"
)

// SysconsoleAncillaryPermissions maps the non-sysconsole permissions required by each sysconsole view.
var SysconsoleAncillaryPermissions map[string][]*Permission

var BuiltInSchemeManagedRoleIDs []string

var NewSystemRoleIDs []string

type RoleType string
type RoleScope string

type Role struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	CreateAt      int64    `json:"create_at"`
	UpdateAt      int64    `json:"update_at"`
	DeleteAt      int64    `json:"delete_at"`
	Permissions   []string `json:"permissions"`
	SchemeManaged bool     `json:"scheme_managed"`
	BuiltIn       bool     `json:"built_in"`
}

type RolePatch struct {
	Permissions *[]string `json:"permissions"`
}

type RolePermissions struct {
	RoleID      string
	Permissions []string
}

func IsValidRoleName(roleName string) bool {
	if len(roleName) <= 0 || len(roleName) > ROLE_NAME_MAX_LENGTH {
		return false
	}

	if strings.TrimLeft(roleName, "abcdefghijklmnopqrstuvwxyz0123456789_") != "" {
		return false
	}

	return true
}

func CleanRoleNames(roleNames []string) ([]string, bool) {
	var cleanedRoleNames []string
	for _, roleName := range roleNames {
		if strings.TrimSpace(roleName) == "" {
			continue
		}

		if !IsValidRoleName(roleName) {
			return roleNames, false
		}

		cleanedRoleNames = append(cleanedRoleNames, roleName)
	}

	return cleanedRoleNames, true
}

// MergeChannelHigherScopedPermissions is meant to be invoked on a channel scheme's role and merges the higher-scoped
// channel role's permissions.
func (r *Role) MergeChannelHigherScopedPermissions(higherScopedPermissions *RolePermissions) {
	mergedPermissions := []string{}

	higherScopedPermissionsMap := AsStringBoolMap(higherScopedPermissions.Permissions)
	rolePermissionsMap := AsStringBoolMap(r.Permissions)

	for _, cp := range AllPermissions {
		if cp.Scope != PermissionScopeChannel {
			continue
		}

		_, presentOnHigherScope := higherScopedPermissionsMap[cp.Id]

		// For the channel admin role always look to the higher scope to determine if the role has ther permission.
		// The channel admin is a special case because they're not part of the UI to be "channel moderated", only
		// channel members and channel guests are.
		if higherScopedPermissions.RoleID == CHANNEL_ADMIN_ROLE_ID && presentOnHigherScope {
			mergedPermissions = append(mergedPermissions, cp.Id)
			continue
		}

		_, permissionIsModerated := ChannelModeratedPermissionsMap[cp.Id]
		if permissionIsModerated {
			_, presentOnRole := rolePermissionsMap[cp.Id]
			if presentOnRole && presentOnHigherScope {
				mergedPermissions = append(mergedPermissions, cp.Id)
			}
		} else {
			if presentOnHigherScope {
				mergedPermissions = append(mergedPermissions, cp.Id)
			}
		}
	}

	r.Permissions = mergedPermissions
}

func (r *Role) IsValidWithoutId() (bool, error) {
	if !IsValidRoleName(r.Name) {
		return false, fmt.Errorf("Invalid role name %s", r.Name)
	}

	if len(r.DisplayName) == 0 || len(r.DisplayName) > ROLE_DISPLAY_NAME_MAX_LENGTH {
		return false, fmt.Errorf("Invalid length of display name: %d", len(r.DisplayName))
	}

	if len(r.Description) > ROLE_DESCRIPTION_MAX_LENGTH {
		return false, fmt.Errorf("Invalid length of description: %d", len(r.Description))
	}

	for _, permission := range r.Permissions {
		permissionValidated := false
		for _, p := range append(AllPermissions) {
			if permission == p.Id {
				permissionValidated = true
				break
			}
		}

		if !permissionValidated {
			return false, fmt.Errorf("permission invalidated: %s", permission)
		}
	}

	return true, nil
}

func MakeDefaultRoles() map[string]*Role {
	roles := make(map[string]*Role)

	roles[CHANNEL_GUEST_ROLE_ID] = &Role{
		Name:        "channel_guest",
		DisplayName: "authentication.roles.channel_guest.name",
		Description: "authentication.roles.channel_guest.description",
		Permissions: []string{
			PERMISSION_READ_CHANNEL.Id,
			// PERMISSION_ADD_REACTION.Id,
			// PERMISSION_REMOVE_REACTION.Id,
			// PERMISSION_UPLOAD_FILE.Id,
			PERMISSION_EDIT_POST.Id,
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
			// PERMISSION_USE_SLASH_COMMANDS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[CHANNEL_USER_ROLE_ID] = &Role{
		Name:        "channel_user",
		DisplayName: "authentication.roles.channel_user.name",
		Description: "authentication.roles.channel_user.description",
		Permissions: []string{
			PERMISSION_READ_CHANNEL.Id,
			// PERMISSION_ADD_REACTION.Id,
			// PERMISSION_REMOVE_REACTION.Id,
			PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id,
			// PERMISSION_UPLOAD_FILE.Id,
			// PERMISSION_GET_PUBLIC_LINK.Id,
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
			// PERMISSION_USE_SLASH_COMMANDS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[CHANNEL_ADMIN_ROLE_ID] = &Role{
		Name:        "channel_admin",
		DisplayName: "authentication.roles.channel_admin.name",
		Description: "authentication.roles.channel_admin.description",
		Permissions: []string{
			// PERMISSION_MANAGE_CHANNEL_ROLES.Id,
			// PERMISSION_USE_GROUP_MENTIONS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[TEAM_GUEST_ROLE_ID] = &Role{
		Name:        "team_guest",
		DisplayName: "authentication.roles.team_guest.name",
		Description: "authentication.roles.team_guest.description",
		Permissions: []string{
			PERMISSION_VIEW_TEAM.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[TEAM_USER_ROLE_ID] = &Role{
		Name:        "team_user",
		DisplayName: "authentication.roles.team_user.name",
		Description: "authentication.roles.team_user.description",
		Permissions: []string{
			// PERMISSION_LIST_TEAM_CHANNELS.Id,
			// PERMISSION_JOIN_PUBLIC_CHANNELS.Id,
			// PERMISSION_READ_PUBLIC_CHANNEL.Id,
			PERMISSION_VIEW_TEAM.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[TEAM_POST_ALL_ROLE_ID] = &Role{
		Name:        "team_post_all",
		DisplayName: "authentication.roles.team_post_all.name",
		Description: "authentication.roles.team_post_all.description",
		Permissions: []string{
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[TEAM_POST_ALL_PUBLIC_ROLE_ID] = &Role{
		Name:        "team_post_all_public",
		DisplayName: "authentication.roles.team_post_all_public.name",
		Description: "authentication.roles.team_post_all_public.description",
		Permissions: []string{
			// PERMISSION_CREATE_POST_PUBLIC.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[TEAM_ADMIN_ROLE_ID] = &Role{
		Name:        "team_admin",
		DisplayName: "authentication.roles.team_admin.name",
		Description: "authentication.roles.team_admin.description",
		Permissions: []string{
			// PERMISSION_REMOVE_USER_FROM_TEAM.Id,
			// PERMISSION_MANAGE_TEAM.Id,
			// PERMISSION_IMPORT_TEAM.Id,
			// PERMISSION_MANAGE_TEAM_ROLES.Id,
			// PERMISSION_MANAGE_CHANNEL_ROLES.Id,
			// PERMISSION_MANAGE_OTHERS_INCOMING_WEBHOOKS.Id,
			// PERMISSION_MANAGE_OTHERS_OUTGOING_WEBHOOKS.Id,
			// PERMISSION_MANAGE_SLASH_COMMANDS.Id,
			// PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS.Id,
			// PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id,
			// PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id,
			// PERMISSION_CONVERT_PUBLIC_CHANNEL_TO_PRIVATE.Id,
			// PERMISSION_CONVERT_PRIVATE_CHANNEL_TO_PUBLIC.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_USER_ROLE_ID] = &Role{
		Name:        "system_user",
		DisplayName: "authentication.roles.global_user.name",
		Description: "authentication.roles.global_user.description",
		Permissions: []string{
			// PERMISSION_LIST_PUBLIC_TEAMS.Id,
			// PERMISSION_JOIN_PUBLIC_TEAMS.Id,
			// PERMISSION_CREATE_DIRECT_CHANNEL.Id,
			// PERMISSION_CREATE_GROUP_CHANNEL.Id,
			PERMISSION_VIEW_MEMBERS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_GUEST_ROLE_ID] = &Role{
		Name:        "system_guest",
		DisplayName: "authentication.roles.global_guest.name",
		Description: "authentication.roles.global_guest.description",
		Permissions: []string{
			// PERMISSION_CREATE_DIRECT_CHANNEL.Id,
			// PERMISSION_CREATE_GROUP_CHANNEL.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_USER_ROLE_ID] = &Role{
		Name:        "system_user",
		DisplayName: "authentication.roles.global_user.name",
		Description: "authentication.roles.global_user.description",
		Permissions: []string{
			// PERMISSION_LIST_PUBLIC_TEAMS.Id,
			// PERMISSION_JOIN_PUBLIC_TEAMS.Id,
			// PERMISSION_CREATE_DIRECT_CHANNEL.Id,
			// PERMISSION_CREATE_GROUP_CHANNEL.Id,
			PERMISSION_VIEW_MEMBERS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_POST_ALL_ROLE_ID] = &Role{
		Name:        "system_post_all",
		DisplayName: "authentication.roles.system_post_all.name",
		Description: "authentication.roles.system_post_all.description",
		Permissions: []string{
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_USER_ACCESS_TOKEN_ROLE_ID] = &Role{
		Name:        "system_user_access_token",
		DisplayName: "authentication.roles.system_user_access_token.name",
		Description: "authentication.roles.system_user_access_token.description",
		Permissions: []string{
			// PERMISSION_CREATE_USER_ACCESS_TOKEN.Id,
			// PERMISSION_READ_USER_ACCESS_TOKEN.Id,
			// PERMISSION_REVOKE_USER_ACCESS_TOKEN.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_USER_MANAGER_ROLE_ID] = &Role{
		Name:        "system_user_manager",
		DisplayName: "authentication.roles.system_user_manager.name",
		Description: "authentication.roles.system_user_manager.description",
		Permissions: []string{
			PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_TEAMS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_CHANNELS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_PERMISSIONS.Id,
			PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_GROUPS.Id,
			// PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_TEAMS.Id,
			// PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_CHANNELS.Id,
			// PERMISSION_SYSCONSOLE_READ_AUTHENTICATION.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_READ_ONLY_ADMIN_ROLE_ID] = &Role{
		Name:        "system_read_only_admin",
		DisplayName: "authentication.roles.system_read_only_admin.name",
		Description: "authentication.roles.system_read_only_admin.description",
		Permissions: []string{
			// PERMISSION_SYSCONSOLE_READ_ABOUT.Id,
			// PERMISSION_SYSCONSOLE_READ_REPORTING.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_USERS.Id,
			PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_TEAMS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_CHANNELS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_PERMISSIONS.Id,
			// PERMISSION_SYSCONSOLE_READ_ENVIRONMENT.Id,
			// PERMISSION_SYSCONSOLE_READ_SITE.Id,
			// PERMISSION_SYSCONSOLE_READ_AUTHENTICATION.Id,
			// PERMISSION_SYSCONSOLE_READ_PLUGINS.Id,
			// PERMISSION_SYSCONSOLE_READ_INTEGRATIONS.Id,
			// PERMISSION_SYSCONSOLE_READ_EXPERIMENTAL.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_MANAGER_ROLE_ID] = &Role{
		Name:        "system_manager",
		DisplayName: "authentication.roles.system_manager.name",
		Description: "authentication.roles.system_manager.description",
		Permissions: []string{
			// PERMISSION_SYSCONSOLE_READ_ABOUT.Id,
			// PERMISSION_SYSCONSOLE_READ_REPORTING.Id,
			PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_TEAMS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_CHANNELS.Id,
			// PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_PERMISSIONS.Id,
			PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_GROUPS.Id,
			// PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_TEAMS.Id,
			// PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_CHANNELS.Id,
			// PERMISSION_SYSCONSOLE_WRITE_USERMANAGEMENT_PERMISSIONS.Id,
			// PERMISSION_SYSCONSOLE_READ_ENVIRONMENT.Id,
			// PERMISSION_SYSCONSOLE_WRITE_ENVIRONMENT.Id,
			// PERMISSION_SYSCONSOLE_READ_SITE.Id,
			// PERMISSION_SYSCONSOLE_WRITE_SITE.Id,
			// PERMISSION_SYSCONSOLE_READ_AUTHENTICATION.Id,
			// PERMISSION_SYSCONSOLE_READ_PLUGINS.Id,
			// PERMISSION_SYSCONSOLE_READ_INTEGRATIONS.Id,
			// PERMISSION_SYSCONSOLE_WRITE_INTEGRATIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	// Add the ancillary permissions to each new system role
	for _, role := range []*Role{roles[SYSTEM_USER_MANAGER_ROLE_ID], roles[SYSTEM_READ_ONLY_ADMIN_ROLE_ID], roles[SYSTEM_MANAGER_ROLE_ID]} {
		for _, rolePermission := range role.Permissions {
			if ancillaryPermissions, ok := SysconsoleAncillaryPermissions[rolePermission]; ok {
				for _, ancillaryPermission := range ancillaryPermissions {
					role.Permissions = append(role.Permissions, ancillaryPermission.Id)
				}
			}
		}
	}

	allPermissionIDs := []string{}
	for _, permission := range AllPermissions {
		allPermissionIDs = append(allPermissionIDs, permission.Id)
	}

	roles[SYSTEM_ADMIN_ROLE_ID] = &Role{
		Name:        "system_admin",
		DisplayName: "authentication.roles.global_admin.name",
		Description: "authentication.roles.global_admin.description",
		// System admins can do anything channel and team admins can do
		// plus everything members of teams and channels can do to all teams
		// and channels on the system
		Permissions:   allPermissionIDs,
		SchemeManaged: true,
		BuiltIn:       true,
	}

	return roles
}
