package sqlstore

import (
	"testing"

	"github.com/masterhung0112/hk_server/v5/store/storetest"
	"github.com/stretchr/testify/suite"
)

func TestSchemeStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &storetest.SchemeStoreTestSuite{}, func(t *testing.T, testSuite storetest.StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}