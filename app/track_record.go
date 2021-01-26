package app

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
	"net/http"
)

func (a *App) CreateTrackRecord(trackRecord *model.TrackRecord) (*model.TrackRecord, *model.AppError) {
	rtrackRecord, err := a.Srv().Store.TrackRecord().Save(trackRecord)
	if err != nil {
		var invErr *store.ErrInvalidInput
		var appErr *model.AppError
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("CreateTrackRecord", "app.trackrecord.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("CreateTrackRecord", "app.trackrecord.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return rtrackRecord, nil
}

func (a *App) CreateTrackPointForTrackRecord(trackPoint *model.TrackPoint, trackRecordId string) (*model.TrackPoint, *model.AppError) {

}
