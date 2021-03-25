package sqlstore

import (
	"database/sql"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
)

type SqlTrackRecordStore struct {
	*SqlStore
}

func newSqlTrackRecordStore(sqlStore *SqlStore) store.TrackRecordStore {
	s := &SqlTrackRecordStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.TrackRecord{}, "TrackRecords").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
	}
	return s
}

func (s *SqlTrackRecordStore) Save(trackRecord *model.TrackRecord) (*model.TrackRecord, error) {
	if len(trackRecord.Id) > 0 {
		return nil, store.NewErrInvalidInput("TrackRecord", "Id", trackRecord.Id)
	}

	trackRecord.PreSave()
	// Check the track point is valid before proceeding.
	if err := trackRecord.IsValidWithoutId(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(trackRecord); err != nil {
		return nil, errors.Wrap(err, "failed to insert TrackRecord")
	}

	return trackRecord, nil
}

func (s *SqlTrackRecordStore) Get(trackRecordId string) (*model.TrackRecord, error) {
	var trackRecord model.TrackRecord

	if err := s.GetReplica().SelectOne(&trackRecord, "SELECT * from TrackRecords WHERE Id = :Id", map[string]interface{}{"Id": trackRecordId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("TrackRecord", trackRecordId)
		}
		return nil, errors.Wrap(err, "failed to get TrackRecord")
	}

	return &trackRecord, nil
}

func (s *SqlTrackRecordStore) update(trackRecord *model.TrackRecord, startUpdate bool, endUpdate bool) (*model.TrackRecord, error) {
	trackRecord.PreUpdate()

	if err := trackRecord.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.TrackRecord{}, trackRecord.Id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get TrackRecord with id=%s", trackRecord.Id)
	}

	if oldResult == nil {
		return nil, store.NewErrInvalidInput("TrackRecord", "id", trackRecord.Id)
	}

	oldTrackRecord := oldResult.(*model.TrackRecord)
	trackRecord.CreateAt = oldTrackRecord.CreateAt
	// User must use the specialized functions to update these fields
	if !startUpdate {
		trackRecord.StartAt = oldTrackRecord.StartAt
	}
	if !endUpdate {
		trackRecord.EndAt = oldTrackRecord.EndAt
	}

	count, err := s.GetMaster().Update(trackRecord)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update TrackRecord with id=%s", trackRecord.Id)
	}
	if count > 1 {
		return nil, errors.Errorf("multiple TrackRecord updated with id=%s", trackRecord.Id)
	}

	// Try to get the new track record after updating from DB
	newUpdatedResult, err := s.GetMaster().Get(model.TrackRecord{}, trackRecord.Id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get TrackRecord after updating with id=%s", trackRecord.Id)
	}

	return newUpdatedResult.(*model.TrackRecord), nil
}

func (s *SqlTrackRecordStore) Update(trackRecord *model.TrackRecord) (*model.TrackRecord, error) {
	return s.update(trackRecord, false, false)
}

func (s *SqlTrackRecordStore) Start(trackRecordId string) (*model.TrackRecord, error) {
	timestamp := model.GetMillis()
	trackRecord, err := s.Get(trackRecordId)
	if err != nil {
		return nil, err
	}

	if trackRecord.StartAt != 0 {
		return nil, errors.Errorf("TrackRecord id=%s has been started", trackRecord.Id)
	}

	if trackRecord.EndAt != 0 {
		return nil, errors.Errorf("TrackRecord id=%s has been ended", trackRecord.Id)
	}

	trackRecord.StartAt = timestamp
	rtrackRecord, err := s.update(trackRecord, true, false)
	if err != nil {
		return nil, err
	}

	return rtrackRecord, nil
}

func (s *SqlTrackRecordStore) End(trackRecordId string) (*model.TrackRecord, error) {
	timestamp := model.GetMillis()
	trackRecord, err := s.Get(trackRecordId)
	if err != nil {
		return nil, err
	}

	if trackRecord.StartAt == 0 {
		return nil, errors.Errorf("TrackRecord id=%s hasn't been started yet", trackRecord.Id)
	}

	if trackRecord.EndAt != 0 {
		return nil, errors.Errorf("TrackRecord id=%s has been ended", trackRecord.Id)
	}

	trackRecord.EndAt = timestamp
	rtrackRecord, err := s.update(trackRecord, false, true)
	if err != nil {
		return nil, err
	}

	return rtrackRecord, nil
}
