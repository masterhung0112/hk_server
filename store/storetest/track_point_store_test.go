package storetest

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TrackPointStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

func TestTrackPointStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &TrackPointStoreTestSuite{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func (s *TrackPointStoreTestSuite) TestSave() {
	tp1 := &model.TrackPoint{
		TargetId:   "111",
		TargetType: "work_bike",
		Point: model.GeoPoint{
			Lat: 123.456,
			Lng: 456.678,
		},
		DeviceId:   "",
		DeviceType: "mobile",
	}

	etp1, err := s.Store().TrackPoint().Save(tp1)
	if s.Nil(err) && s.NotNil(etp1) {
		s.Len(etp1.Id, 26)
		s.Equal(etp1.TargetId, tp1.TargetId)
		s.Equal(etp1.TargetType, tp1.TargetType)
		s.Equal(etp1.DeviceId, tp1.DeviceId)
		s.Equal(etp1.DeviceType, tp1.DeviceType)
		s.NotEqual(etp1.CreateAt, 0)
	}
}

// Test Get individual track point
func (s *TrackPointStoreTestSuite) TestGet() {
}
