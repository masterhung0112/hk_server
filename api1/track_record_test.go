package api1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/masterhung0112/hk_server/model"
)

func TestCreateTrackRecord(t *testing.T) {
  th := Setup(t)
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

	rtrackRecord, resp := th.Client.CreateTrackRecord(trackRecord)
	CheckNoError(t, resp)
	CheckCreatedStatus(t, resp)
	require.NotNil(t, rtrackRecord)
}
