package model

import (
	"strings"
)

const (
  SYSTEM_GUEST_ROLE_ID             = "system_guest"
	SYSTEM_USER_ROLE_ID              = "system_user"
	SYSTEM_ADMIN_ROLE_ID             = "system_admin"
	SYSTEM_POST_ALL_ROLE_ID          = "system_post_all"
	SYSTEM_POST_ALL_PUBLIC_ROLE_ID   = "system_post_all_public"
  SYSTEM_USER_ACCESS_TOKEN_ROLE_ID = "system_user_access_token"

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

func (r *Role) IsValidWithoutId() bool {
	if !IsValidRoleName(r.Name) {
		return false
	}

	if len(r.DisplayName) == 0 || len(r.DisplayName) > ROLE_DISPLAY_NAME_MAX_LENGTH {
		return false
	}

	if len(r.Description) > ROLE_DESCRIPTION_MAX_LENGTH {
		return false
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
			return false
		}
	}

	return true
}