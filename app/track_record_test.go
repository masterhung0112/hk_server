package app

import (
	"testing"

	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/require"
)

func TestCreateAndGetTrackRecord(t *testing.T) {
  th := Setup(t).InitBasic()
  defer th.TearDown()

  trackRecord := &model.TrackRecord{
    Id: "",
    OwnerId: th.BasicUser.Id,
    Categories: []string{},
    CreateAt: 0,
    StartAt: 0,
    EndAt: 0,
    WeightedAverage: 0.0,
    WeightedAverageLastId: "",
  }

  rTrackRecord, err := th.App.CreateTrackRecord(trackRecord)
  require.Nil(t, err, "Should create a new track record")
  require.NotEmpty(t, rTrackRecord.Id)
  require.Greater(t, rTrackRecord.CreateAt, int64(0))
  require.Equal(t, rTrackRecord.StartAt, int64(0))
  require.Equal(t, rTrackRecord.EndAt, int64(0))
  require.Equal(t, rTrackRecord.WeightedAverage, float64(0.0))

  trackRecordFetched, err := th.App.GetTrackRecord(rTrackRecord.Id)
  require.Nil(t, err, "Should get a track record")
  require.Equal(t, rTrackRecord, trackRecordFetched)

  _, err = th.App.EndTrackRecord(rTrackRecord.Id)
  require.NotNil(t, err, "Should not able end the track record")

  trackRecordStarted, err := th.App.StartTrackRecord(rTrackRecord.Id)
  require.Nil(t, err, "Should start the track record")
  require.Greater(t, trackRecordStarted.StartAt, int64(0))
  require.Equal(t, trackRecordStarted.EndAt, int64(0))

  _, err = th.App.StartTrackRecord(rTrackRecord.Id)
  require.NotNil(t, err, "Should start the track record")

  trackRecordEnded, err := th.App.EndTrackRecord(rTrackRecord.Id)
  require.Nil(t, err, "Should not able end the track record")
  require.Greater(t, trackRecordEnded.StartAt, int64(0))
  require.Greater(t, trackRecordEnded.EndAt, int64(0))
}
