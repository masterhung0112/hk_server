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
    // Id: "",
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

func TestCreateTrackPointForRecord(t *testing.T) {
  th := Setup(t)
  defer th.TearDown()

  trackRecord := th.CreateTrackRecord()

  trackPoint := &model.TrackPoint{

  }
  rTrackPoint, resp := th.Client.CreateTrackPointForRecord(trackRecord.Id, trackPoint)
	CheckBadRequestStatus(t, resp)
  require.Nil(t, rTrackPoint)

  // Start the record
  trackRecordStarted, resp := th.Client.StartTrackRecord(trackRecord.Id)
  CheckOKStatus(t, resp)
  require.NotNil(t, trackRecordStarted)

  // expect we can create track point for record after starting the record
  rTrackPoint, resp = th.Client.CreateTrackPointForRecord(trackRecord.Id, trackPoint)
	CheckCreatedStatus(t, resp)
  require.NotNil(t, rTrackPoint)

  // End the track record
  trackRecordEnd, resp := th.Client.EndTrackRecord(trackRecord.Id)
  CheckOKStatus(t, resp)
  require.NotNil(t, trackRecordEnd)
}
