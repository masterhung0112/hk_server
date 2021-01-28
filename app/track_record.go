package app

import (
	"net/http"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
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

func (a *App) GetTrackRecord(trackRecordId string) (*model.TrackRecord, *model.AppError) {
	trackRecord, err := a.Srv().Store.TrackRecord().Get(trackRecordId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetTrackRecord", "app.trackrecord.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetTrackRecord", "app.trackrecord.get.find.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return trackRecord, nil
}

func (a *App) CreateTrackPointForTrackRecord(trackPoint *model.TrackPoint, trackRecordId string) (*model.TrackPoint, *model.AppError) {
	// Find if track Record exist
	trackRecord, err := a.GetTrackRecord(trackRecordId)
	if err != nil {
		return nil, err
	}

	if trackRecord.EndAt != 0 {
		return nil, model.NewAppError("CreateTrackPointForTrackRecord", "app.trackrecord.create.recordclosed.app_error", nil, "trackrecord closed", http.StatusBadRequest)
	}

	// Create trackpoint for trackRecord
	trackPoint.TargetId = trackRecordId

	// Write trackpoint to DB
	rTrackPoint, err := a.CreateTrackPoint(trackPoint)
	if err != nil {
		return nil, err
	}

	//TODO: Update the metric for track record

	return rTrackPoint, nil
}

func (a *App) StartTrackRecord(trackRecordId string) (*model.TrackRecord, *model.AppError) {
	trackRecord, err := a.Srv().Store.TrackRecord().Start(trackRecordId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("StartTrackRecord", "app.trackrecord.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("StartTrackRecord", "app.trackrecord.start.set.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return trackRecord, nil
}

func (a *App) EndTrackRecord(trackRecordId string) (*model.TrackRecord, *model.AppError) {
	trackRecord, err := a.Srv().Store.TrackRecord().End(trackRecordId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("EndTrackRecord", "app.trackrecord.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("EndTrackRecord", "app.trackrecord.end.set.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return trackRecord, nil
}
