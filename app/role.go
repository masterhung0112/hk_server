package app

import (
	"github.com/masterhung0112/hk_server/model"
	"net/http"
)

// mergeChannelHigherScopedPermissions updates the permissions based on the role type, whether the permission is
// moderated, and the value of the permission on the higher-scoped scheme.
func (s *Server) mergeChannelHigherScopedPermissions(roles []*model.Role) *model.AppError {
	var higherScopeNamesToQuery []string

	for _, role := range roles {
		if role.SchemeManaged {
			higherScopeNamesToQuery = append(higherScopeNamesToQuery, role.Name)
		}
	}

	if len(higherScopeNamesToQuery) == 0 {
		return nil
	}

	higherScopedPermissionsMap, err := s.Store.Role().ChannelHigherScopedPermissions(higherScopeNamesToQuery)
	if err != nil {
		return model.NewAppError("mergeChannelHigherScopedPermissions", "app.role.get_by_names.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, role := range roles {
		if role.SchemeManaged {
			if higherScopedPermissions, ok := higherScopedPermissionsMap[role.Name]; ok {
				role.MergeChannelHigherScopedPermissions(higherScopedPermissions)
			}
		}
	}

	return nil
}

// mergeChannelHigherScopedPermissions updates the permissions based on the role type, whether the permission is
// moderated, and the value of the permission on the higher-scoped scheme.
func (a *App) mergeChannelHigherScopedPermissions(roles []*model.Role) *model.AppError {
	return a.Srv().mergeChannelHigherScopedPermissions(roles)
}

func (a *App) GetRolesByNames(names []string) ([]*model.Role, *model.AppError) {
	roles, nErr := a.Srv().Store.Role().GetByNames(names)
	if nErr != nil {
		return nil, model.NewAppError("GetRolesByNames", "app.role.get_by_names.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	}

	err := a.mergeChannelHigherScopedPermissions(roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (a *App) CheckRolesExist(roleNames []string) *model.AppError {
	roles, err := a.GetRolesByNames(roleNames)
	if err != nil {
		return err
	}

	for _, name := range roleNames {
		nameFound := false
		for _, role := range roles {
			if name == role.Name {
				nameFound = true
				break
			}
		}
		if !nameFound {
			return model.NewAppError("CheckRolesExist", "app.role.check_roles_exist.role_not_found", nil, "role="+name, http.StatusBadRequest)
		}
	}

	return nil
}

func (a *App) GetAllRoles() ([]*model.Role, *model.AppError) {
	roles, err := a.Srv().Store.Role().GetAll()
	if err != nil {
		return nil, model.NewAppError("GetAllRoles", "app.role.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return roles, nil
}
