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
