package app

import (
	"net/http"

	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/store"
	"github.com/pkg/errors"
)

func (a *App) CreateTrackPoint(trackPoint *model.TrackPoint) (*model.TrackPoint, *model.AppError) {
	rtrackPoint, err := a.Srv().Store.TrackPoint().Save(trackPoint)
	if err != nil {
		var invErr *store.ErrInvalidInput
		var appErr *model.AppError
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("CreateTrackPoint", "app.trackpoint.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("CreateTrackPoint", "app.trackpoint.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return rtrackPoint, nil
}
