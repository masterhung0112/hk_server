package app

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
	"net/http"
)

type permissionTransformation struct {
	On     func(*model.Role, map[string]map[string]bool) bool
	Add    []string
	Remove []string
}

type permissionsMap []permissionTransformation

func applyPermissionsMap(role *model.Role, roleMap map[string]map[string]bool, migrationMap permissionsMap) []string {
	var result []string

	roleName := role.Name
	for _, transformation := range migrationMap {
		if transformation.On(role, roleMap) {
			for _, permission := range transformation.Add {
				roleMap[roleName][permission] = true
			}
			for _, permission := range transformation.Remove {
				roleMap[roleName][permission] = false
			}
		}
	}

	for key, active := range roleMap[roleName] {
		if active {
			result = append(result, key)
		}
	}
	return result
}

func (a *App) doPermissionsMigration(key string, migrationMap permissionsMap) *model.AppError {
	if _, err := a.Srv().Store.System().GetByName(key); err == nil {
		return nil
	}

	roles, err := a.GetAllRoles()
	if err != nil {
		return err
	}

	roleMap := make(map[string]map[string]bool)
	for _, role := range roles {
		roleMap[role.Name] = make(map[string]bool)
		for _, permission := range role.Permissions {
			roleMap[role.Name][permission] = true
		}
	}

	for _, role := range roles {
		role.Permissions = applyPermissionsMap(role, roleMap, migrationMap)
		if _, err := a.Srv().Store.Role().Save(role); err != nil {
			var invErr *store.ErrInvalidInput
			switch {
			case errors.As(err, &invErr):
				return model.NewAppError("doPermissionsMigration", "app.role.save.invalid_role.app_error", nil, invErr.Error(), http.StatusBadRequest)
			default:
				return model.NewAppError("doPermissionsMigration", "app.role.save.insert.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		}
	}

	if err := a.Srv().Store.System().Save(&model.System{Name: key, Value: "true"}); err != nil {
		return model.NewAppError("doPermissionsMigration", "app.system.save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

// DoPermissionsMigrations execute all the permissions migrations need by the current version.
func (a *App) DoPermissionsMigrations() error {
	PermissionsMigrations := []struct {
		Key       string
		Migration func() (permissionsMap, error)
	}{
		//TODO: Open
		// {Key: model.MIGRATION_KEY_EMOJI_PERMISSIONS_SPLIT, Migration: a.getEmojisPermissionsSplitMigration},
		// {Key: model.MIGRATION_KEY_WEBHOOK_PERMISSIONS_SPLIT, Migration: a.getWebhooksPermissionsSplitMigration},
		// {Key: model.MIGRATION_KEY_LIST_JOIN_PUBLIC_PRIVATE_TEAMS, Migration: a.getListJoinPublicPrivateTeamsPermissionsMigration},
		// {Key: model.MIGRATION_KEY_REMOVE_PERMANENT_DELETE_USER, Migration: a.removePermanentDeleteUserMigration},
		// {Key: model.MIGRATION_KEY_ADD_BOT_PERMISSIONS, Migration: a.getAddBotPermissionsMigration},
		// {Key: model.MIGRATION_KEY_APPLY_CHANNEL_MANAGE_DELETE_TO_CHANNEL_USER, Migration: a.applyChannelManageDeleteToChannelUser},
		// {Key: model.MIGRATION_KEY_REMOVE_CHANNEL_MANAGE_DELETE_FROM_TEAM_USER, Migration: a.removeChannelManageDeleteFromTeamUser},
		// {Key: model.MIGRATION_KEY_VIEW_MEMBERS_NEW_PERMISSION, Migration: a.getViewMembersPermissionMigration},
		// {Key: model.MIGRATION_KEY_ADD_MANAGE_GUESTS_PERMISSIONS, Migration: a.getAddManageGuestsPermissionsMigration},
		// {Key: model.MIGRATION_KEY_CHANNEL_MODERATIONS_PERMISSIONS, Migration: a.channelModerationPermissionsMigration},
		// {Key: model.MIGRATION_KEY_ADD_USE_GROUP_MENTIONS_PERMISSION, Migration: a.getAddUseGroupMentionsPermissionMigration},
		// {Key: model.MIGRATION_KEY_ADD_SYSTEM_CONSOLE_PERMISSIONS, Migration: a.getAddSystemConsolePermissionsMigration},
		// {Key: model.MIGRATION_KEY_ADD_CONVERT_CHANNEL_PERMISSIONS, Migration: a.getAddConvertChannelPermissionsMigration},
	}

	for _, migration := range PermissionsMigrations {
		migMap, err := migration.Migration()
		if err != nil {
			return err
		}
		if err := a.doPermissionsMigration(migration.Key, migMap); err != nil {
			return err
		}
	}
	return nil
}
