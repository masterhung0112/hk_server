package app

import (
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/utils"
	"reflect"
)

// This function migrates the default built in roles from code/config to the database.
func (a *App) DoAdvancedPermissionsMigration() {
	// If the migration is already marked as completed, don't do it again.
	if _, err := a.Srv().Store.System().GetByName(model.ADVANCED_PERMISSIONS_MIGRATION_KEY); err == nil {
		return
	}

	mlog.Info("Migrating roles to database.")
	roles := model.MakeDefaultRoles()
	roles = utils.SetRolePermissionsFromConfig(roles, a.Config(), a.Srv().License() != nil)

	allSucceeded := true

	for _, role := range roles {
		_, err := a.Srv().Store.Role().Save(role)
		if err == nil {
			continue
		}

		// If this failed for reasons other than the role already existing, don't mark the migration as done.
		fetchedRole, err := a.Srv().Store.Role().GetByName(role.Name)
		if err != nil {
			mlog.Critical("Failed to migrate role to database.", mlog.Err(err))
			allSucceeded = false
			continue
		}

		// If the role already existed, check it is the same and update if not.
		if !reflect.DeepEqual(fetchedRole.Permissions, role.Permissions) ||
			fetchedRole.DisplayName != role.DisplayName ||
			fetchedRole.Description != role.Description ||
			fetchedRole.SchemeManaged != role.SchemeManaged {
			role.Id = fetchedRole.Id
			if _, err = a.Srv().Store.Role().Save(role); err != nil {
				// Role is not the same, but failed to update.
				mlog.Critical("Failed to migrate role to database.", mlog.Err(err))
				allSucceeded = false
			}
		}
	}

	if !allSucceeded {
		return
	}

	config := a.Config()
	if *config.ServiceSettings.DEPRECATED_DO_NOT_USE_AllowEditPost == model.ALLOW_EDIT_POST_ALWAYS {
		*config.ServiceSettings.PostEditTimeLimit = -1
		if err := a.SaveConfig(config, true); err != nil {
			mlog.Error("Failed to update config in Advanced Permissions Phase 1 Migration.", mlog.Err(err))
		}
	}

	system := model.System{
		Name:  model.ADVANCED_PERMISSIONS_MIGRATION_KEY,
		Value: "true",
	}

	if err := a.Srv().Store.System().Save(&system); err != nil {
		mlog.Critical("Failed to mark advanced permissions migration as completed.", mlog.Err(err))
	}
}

func (a *App) DoAppMigrations() {
	a.DoAdvancedPermissionsMigration()
	// a.DoEmojisPermissionsMigration()
	a.DoGuestRolesCreationMigration()
	a.DoSystemConsoleRolesCreationMigration()
	// This migration always must be the last, because can be based on previous
	// migrations. For example, it needs the guest roles migration.
	err := a.DoPermissionsMigrations()
	if err != nil {
		mlog.Critical("(app.App).DoPermissionsMigrations failed", mlog.Err(err))
	}
}
