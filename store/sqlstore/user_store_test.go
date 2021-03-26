package sqlstore

import (
	"testing"

	"github.com/masterhung0112/hk_server/v5/store/storetest"
	"github.com/stretchr/testify/suite"
)

func TestUserStoreTS(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &storetest.UserStoreTS{}, func(t *testing.T, testSuite storetest.StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func TestUserStoreGetAllProfilesTS(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &storetest.UserStoreGetAllProfilesTS{}, func(t *testing.T, testSuite storetest.StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}