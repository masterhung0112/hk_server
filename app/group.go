package app

import (
	"github.com/masterhung0112/hk_server/model"
	"net/http"
)

// UserIsInAdminRoleGroup returns true at least one of the user's groups are configured to set the members as
// admins in the given syncable.
func (a *App) UserIsInAdminRoleGroup(userID, syncableID string, syncableType model.GroupSyncableType) (bool, *model.AppError) {
	groupIDs, err := a.Srv().Store.Group().AdminRoleGroupsForSyncableMember(userID, syncableID, syncableType)
	if err != nil {
		return false, model.NewAppError("UserIsInAdminRoleGroup", "app.select_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if len(groupIDs) == 0 {
		return false, nil
	}

	return true, nil
}
