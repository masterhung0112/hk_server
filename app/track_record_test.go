package app

import (
	"testing"

	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/require"
)

func TestCreateTrackRecord(t *testing.T) {
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
  require.Greater(t, rTrackRecord.CreateAt, int64(0))
  require.Equal(t, rTrackRecord.StartAt, int64(0))
  require.Equal(t, rTrackRecord.EndAt, int64(0))
  require.Equal(t, rTrackRecord.WeightedAverage, float64(0.0))
}
