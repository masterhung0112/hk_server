package sqlstore

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
)

type SqlTrackPointStore struct {
	*SqlStore
}

type TrackPoint struct {
	Id         string
	TargetId   string
	Lat        float64
	Lng        float64
	CreateAt   int64
	DeviceId   string
}

func NewTrackPointFromModel(trackPointModel *model.TrackPoint) *TrackPoint {
	return &TrackPoint{
		Id:         trackPointModel.Id,
		TargetId:   trackPointModel.TargetId,
		Lat:        trackPointModel.Point.Lat,
		Lng:        trackPointModel.Point.Lng,
		CreateAt:   trackPointModel.CreateAt,
		DeviceId:   trackPointModel.DeviceId,
	}
}

func (trackPoint TrackPoint) ToModel() *model.TrackPoint {
	return &model.TrackPoint{
		Id:         trackPoint.Id,
		TargetId:   trackPoint.TargetId,
		Point: model.GeoPoint{
			Lat: trackPoint.Lat,
			Lng: trackPoint.Lng,
		},
		CreateAt:   trackPoint.CreateAt,
		DeviceId:   trackPoint.DeviceId,
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
	if err := s.GetMaster().Insert(dbTrackPoint); err != nil {
		return nil, errors.Wrap(err, "failed to insert TrackPoint")
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

func (s *SqlTrackPointStore) GetByTargetId(targetId string) ([]*model.TrackPoint, error) {
	query, args, _ := s.getQueryBuilder().
		Select("*").
		Where(sq.Eq{"TargetId": targetId}).
		ToSql()
	var result []*model.TrackPoint
	if _, err := s.GetReplica().Select(&result, query, args...); err != nil {
		return nil, errors.Wrap(err, "failed to fetch track points for target id")
	}
	return result, nil
}
