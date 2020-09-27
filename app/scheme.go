package app

import (
	"errors"
	"net/http"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
)

func (a *App) CreateScheme(scheme *model.Scheme) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	// Clear any user-provided values for trusted properties.
	scheme.DefaultTeamAdminRole = ""
	scheme.DefaultTeamUserRole = ""
	scheme.DefaultTeamGuestRole = ""
	scheme.DefaultChannelAdminRole = ""
	scheme.DefaultChannelUserRole = ""
	scheme.DefaultChannelGuestRole = ""
	scheme.CreateAt = 0
	scheme.UpdateAt = 0
	scheme.DeleteAt = 0

	scheme, err := a.Srv().Store.Scheme().Save(scheme)
	if err != nil {
		var invErr *store.ErrInvalidInput
		var appErr *model.AppError
		switch {
		case errors.As(err, &appErr):
			return nil, appErr
		case errors.As(err, &invErr):
			return nil, model.NewAppError("CreateScheme", "app.scheme.save.invalid_scheme.app_error", nil, err.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("CreateScheme", "app.scheme.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return scheme, nil
}

func (a *App) UpdateScheme(scheme *model.Scheme) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	scheme, err := a.Srv().Store.Scheme().Save(scheme)
	if err != nil {
		var invErr *store.ErrInvalidInput
		var appErr *model.AppError
		switch {
		case errors.As(err, &appErr):
			return nil, appErr
		case errors.As(err, &invErr):
			return nil, model.NewAppError("UpdateScheme", "app.scheme.save.invalid_scheme.app_error", nil, err.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("UpdateScheme", "app.scheme.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return scheme, nil
}

func (a *App) GetScheme(id string) (*model.Scheme, *model.AppError) {
	if appErr := a.IsPhase2MigrationCompleted(); appErr != nil {
		return nil, appErr
	}

	scheme, err := a.Srv().Store.Scheme().Get(id)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetScheme", "app.scheme.get.app_error", nil, err.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetScheme", "app.scheme.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return scheme, nil
}

func (a *App) GetSchemeByName(name string) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	scheme, err := a.Srv().Store.Scheme().GetByName(name)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetSchemeByName", "app.scheme.get.app_error", nil, err.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetSchemeByName", "app.scheme.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return scheme, nil
}

func (s *Server) IsPhase2MigrationCompleted() *model.AppError {
	if s.phase2PermissionsMigrationComplete {
		return nil
	}

	if _, err := s.Store.System().GetByName(model.MIGRATION_KEY_ADVANCED_PERMISSIONS_PHASE_2); err != nil {
		return model.NewAppError("App.IsPhase2MigrationCompleted", "app.schemes.is_phase_2_migration_completed.not_completed.app_error", nil, err.Error(), http.StatusNotImplemented)
	}

	s.phase2PermissionsMigrationComplete = true

	return nil
}

func (a *App) IsPhase2MigrationCompleted() *model.AppError {
	return a.Srv().IsPhase2MigrationCompleted()
}
