package testlib

import (
	"github.com/masterhung0112/go_server/store/storetest"
	"github.com/masterhung0112/go_server/store/sqlstore"
	"github.com/masterhung0112/go_server/store"
	"os"
	"github.com/masterhung0112/go_server/model"
	"testing"
	"flag"
)

type MainHelper struct {
	Settings         *model.SqlSettings
  Store            store.Store

  SQLSupplier      *sqlstore.SqlSupplier
	status           int
}

type HelperOptions struct {
	EnableStore     bool
	EnableResources bool
}

func NewMainHelperWithOptions(options *HelperOptions) *MainHelper {
  var mainHelper MainHelper
  flag.Parse()

  if options != nil {
		if options.EnableStore && !testing.Short() {
			mainHelper.setupStore()
		}

		if options.EnableResources {
			// mainHelper.setupResources()
		}
	}

  return &mainHelper
}

func (h *MainHelper) setupStore() {
  driverName := os.Getenv("MM_SQLSETTINGS_DRIVERNAME")
	if driverName == "" {
		driverName = model.DATABASE_DRIVER_MYSQL
  }

  h.Settings = storetest.MakeSqlSettings(driverName)

  config := &model.Config{}
  config.SetDefaults()

  h.SQLSupplier = sqlstore.NewSqlSupplier(*h.Settings, nil)
  h.Store = &TestStore{
		h.SQLSupplier,
	}
  // searchlayer.NewSearchLayer(&TestStore{
	// 	h.SQLSupplier,
	// }, h.SearchEngine, config)
}