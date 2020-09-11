package testlib

import (
	"flag"
	"github.com/masterhung0112/go_server/mlog"
	"log"
	"os"
	"testing"

	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
	"github.com/masterhung0112/go_server/store/sqlstore"
	"github.com/masterhung0112/go_server/store/storetest"
)

// Keep
// - SqlSetting
// - Store
//
type MainHelper struct {
	Settings *model.SqlSettings
	Store    store.Store

	SQLSupplier *sqlstore.SqlSupplier
	status      int
}

type HelperOptions struct {
	EnableStore     bool
	EnableResources bool
}

func NewMainHelperWithOptions(options *HelperOptions) *MainHelper {
	var mainHelper MainHelper
	flag.Parse()

	// Setup a global logger to catch tests logging outside of app context
	// The global logger will be stomped by apps initializing but that's fine for testing.
	// Ideally this won't happen.
	mlog.InitGlobalLogger(mlog.NewLogger(&mlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "error",
		EnableFile:    false,
	}))

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
		// Use MySQL my default for database
		driverName = model.DATABASE_DRIVER_MYSQL
	}

	// Setup the SQL setting
	h.Settings = storetest.MakeSqlSettings(driverName)

	// Get the default config
	config := &model.Config{}
	config.SetDefaults()

	// Create SQL Store
	h.SQLSupplier = sqlstore.NewSqlSupplier(*h.Settings)
	h.Store = &TestStore{
		h.SQLSupplier,
	}
	// searchlayer.NewSearchLayer(&TestStore{
	// 	h.SQLSupplier,
	// }, h.SearchEngine, config)
}

func (h *MainHelper) Main(m *testing.M) {
	h.status = m.Run()
}

func (h *MainHelper) Close() error {
	if h.SQLSupplier != nil {
		h.SQLSupplier.Close()
	}
	if h.Settings != nil {
		storetest.CleanupSqlSettings(h.Settings)
	}
	//TODO: Open
	// if h.testResourcePath != "" {
	// 	os.RemoveAll(h.testResourcePath)
	// }

	if r := recover(); r != nil {
		log.Fatalln(r)
	}

	os.Exit(h.status)

	return nil
}

func (h *MainHelper) GetStore() store.Store {
	if h.Store == nil {
		panic("MainHelper not initialized with store.")
	}

	return h.Store
}

func (h *MainHelper) GetSQLSupplier() *sqlstore.SqlSupplier {
	if h.SQLSupplier == nil {
		panic("MainHelper not initialized with sql supplier.")
	}

	return h.SQLSupplier
}
