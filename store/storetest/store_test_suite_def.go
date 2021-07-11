package storetest

import (
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/stretchr/testify/suite"
)


type PostStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

type RoleStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

type TeamStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

type TrackPointStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

type UserStoreTS struct {
	suite.Suite
	StoreTestSuite
}


type UserStoreGetAllProfilesTS struct {
	suite.Suite
	StoreTestSuite

	u1 *model.User
	u2 *model.User
	u3 *model.User
	u4 *model.User
	u5 *model.User
	u6 *model.User
	u7 *model.User
}

type UserStoreGetProfilesTS struct {
	suite.Suite
	StoreTestSuite

	u1     *model.User
	u2     *model.User
	u3     *model.User
	u4     *model.User
	u5     *model.User
	u6     *model.User
	u7     *model.User
	teamId string
}

type UserStoreGetProfilesByIdsTS struct {
	suite.Suite
	StoreTestSuite

	u1     *model.User
	u2     *model.User
	u3     *model.User
	u4     *model.User
	u5     *model.User
	u6     *model.User
	u7     *model.User
	teamId string
}

type ChannelStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}
