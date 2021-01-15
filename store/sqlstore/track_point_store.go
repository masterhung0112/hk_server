package sqlstore

import (
	"database/sql"
	"fmt"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
)

type SqlTrackPointStore struct {
	*SqlStore
}

type TrackPoint struct {
	Id         string
	TargetId   uint64
	TargetType string
	Lat        float64
	Lng        float64
	CreateAt   int64
	DeviceId   string
	DeviceType string
}

func NewTrackPointFromModel(trackPointModel *model.TrackPoint) *TrackPoint {
	return &TrackPoint{
		Id:         trackPointModel.Id,
		TargetId:   trackPointModel.TargetId,
		TargetType: trackPointModel.TargetType,
		Lat:        trackPointModel.Point.Lat,
		Lng:        trackPointModel.Point.Lng,
		CreateAt:   trackPointModel.CreateAt,
		DeviceId:   trackPointModel.DeviceId,
		DeviceType: trackPointModel.DeviceType,
	}
}

func (trackPoint TrackPoint) ToModel() *model.TrackPoint {
	return &model.TrackPoint{
		Id:         trackPoint.Id,
		TargetId:   trackPoint.TargetId,
		TargetType: trackPoint.TargetType,
		Point: model.GeoPoint{
			Lat: trackPoint.Lat,
			Lng: trackPoint.Lng,
		},
		CreateAt:   trackPoint.CreateAt,
		DeviceId:   trackPoint.DeviceId,
		DeviceType: trackPoint.DeviceType,
	}
}

func newSqlTrackPointStore(sqlStore *SqlStore) store.TrackPointStore {
	s := &SqlTrackPointStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(TrackPoint{}, "TrackPoints").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
	}
	return s
}

func (s *SqlTrackPointStore) Save(trackPoint *model.TrackPoint) (*model.TrackPoint, error) {
	if len(trackPoint.Id) > 0 {
		return nil, store.NewErrInvalidInput("TrackPoint", "Id", trackPoint.Id)
	}

	trackPoint.PreSave()
	// Check the track point is valid before proceeding.
	if err := trackPoint.IsValidWithoutId(); err != nil {
		return nil, err
	}

	dbTrackPoint := NewTrackPointFromModel(trackPoint)
	if rowsChanged, err := s.GetMaster().Update(dbTrackPoint); err != nil {
		return nil, errors.Wrap(err, "failed to update TrackPoint")
	} else if rowsChanged != 1 {
		return nil, fmt.Errorf("invalid number of updated rows, expected 1 but got %d", rowsChanged)
	}

	return dbTrackPoint.ToModel(), nil
}

func (s *SqlTrackPointStore) Get(trackPointId string) (*model.TrackPoint, error) {
	var dbTrackPoint TrackPoint

	if err := s.GetReplica().SelectOne(&dbTrackPoint, "SELECT * from TrackPoints WHERE Id = :Id", map[string]interface{}{"Id": trackPointId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("TrackPoint", trackPointId)
		}
		return nil, errors.Wrap(err, "failed to get TrackPoint")
	}

	return dbTrackPoint.ToModel(), nil
}
