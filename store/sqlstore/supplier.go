package sqlstore

import (
	dbsql "database/sql"
	"encoding/json"
	"fmt"
	"github.com/masterhung0112/hk_server/einterfaces"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/mattermost/gorp"
	"github.com/pkg/errors"
)

type SqlSupplierStores struct {
	team            store.TeamStore
	user            store.UserStore
	bot             store.BotStore
	post            store.PostStore
	thread          store.ThreadStore
  system          store.SystemStore
  status          store.StatusStore
	role            store.RoleStore
	scheme          store.SchemeStore
	channel         store.ChannelStore
	session         store.SessionStore
	userAccessToken store.UserAccessTokenStore
	token           store.TokenStore
	preference      store.PreferenceStore
	group           store.GroupStore
}

type SqlSupplier struct {
	// rrCounter and srCounter should be kept first.
	// See https://github.com/mattermost/mattermost-server/v5/pull/7281
	rrCounter int64
	srCounter int64

	master         *gorp.DbMap
	replicas       []*gorp.DbMap
	searchReplicas []*gorp.DbMap
	stores         SqlSupplierStores
	settings       *model.SqlSettings
	lockedToMaster bool
	license        *model.License
}

const (
	DB_PING_ATTEMPTS     = 18
	DB_PING_TIMEOUT_SECS = 10
)

const (
	EXIT_GENERIC_FAILURE = 1
	EXIT_CREATE_TABLE    = 100
	EXIT_DB_OPEN         = 101
	EXIT_PING            = 102
	EXIT_NO_DRIVER       = 103
	EXIT_TABLE_EXISTS    = 104
)

type mattermConverter struct{}

type JSONSerializable interface {
	ToJson() string
}

func (me mattermConverter) ToDb(val interface{}) (interface{}, error) {

	switch t := val.(type) {
	case model.StringMap:
		return model.MapToJson(t), nil
	case map[string]string:
		return model.MapToJson(model.StringMap(t)), nil
	case model.StringArray:
		return model.ArrayToJson(t), nil
	case model.StringInterface:
		return model.StringInterfaceToJson(t), nil
	case map[string]interface{}:
		return model.StringInterfaceToJson(model.StringInterface(t)), nil
	case JSONSerializable:
		return t.ToJson(), nil
	case *opengraph.OpenGraph:
		return json.Marshal(t)
	}

	return val, nil
}

func (me mattermConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
	case *model.StringMap:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_map"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *map[string]string:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_map"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *model.StringArray:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_array"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *model.StringInterface:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_interface"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *map[string]interface{}:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_interface"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	}

	return gorp.CustomScanner{}, false
}

func (ss *SqlSupplier) DriverName() string {
	return *ss.settings.DriverName
}

func (ss *SqlSupplier) GetMaster() *gorp.DbMap {
	return ss.master
}

func (ss *SqlSupplier) GetSearchReplica() *gorp.DbMap {
	//TODO: Open
	// ss.licenseMutex.RLock()
	license := ss.license
	// ss.licenseMutex.RUnlock()
	if license == nil {
		return ss.GetMaster()
	}

	if len(ss.settings.DataSourceSearchReplicas) == 0 {
		return ss.GetReplica()
	}

	rrNum := atomic.AddInt64(&ss.srCounter, 1) % int64(len(ss.searchReplicas))
	return ss.searchReplicas[rrNum]
}

func (ss *SqlSupplier) GetReplica() *gorp.DbMap {
	if len(ss.settings.DataSourceReplicas) == 0 || ss.lockedToMaster || ss.license == nil {
		return ss.GetMaster()
	}

	rrNum := atomic.AddInt64(&ss.rrCounter, 1) % int64(len(ss.replicas))
	return ss.replicas[rrNum]
}

func (ss *SqlSupplier) User() store.UserStore {
	return ss.stores.user
}

func (ss *SqlSupplier) Bot() store.BotStore {
	return ss.stores.bot
}

func (ss *SqlSupplier) Post() store.PostStore {
	return ss.stores.post
}

func (ss *SqlSupplier) Thread() store.ThreadStore {
	return ss.stores.thread
}

func setupConnection(con_type string, dataSource string, settings *model.SqlSettings) *gorp.DbMap {
	db, err := dbsql.Open(*settings.DriverName, dataSource)
	if err != nil {
		mlog.Critical("Failed to open SQL connection to err.", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_DB_OPEN)
	}

	for i := 0; i < DB_PING_ATTEMPTS; i++ {
		//TODO: Add ping attemp
	}

	connectionTimeout := time.Duration(*settings.QueryTimeout) * time.Second

	var dbmap *gorp.DbMap

	if *settings.DriverName == model.DATABASE_DRIVER_SQLITE {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.SqliteDialect{}, QueryTimeout: connectionTimeout}
	} else if *settings.DriverName == model.DATABASE_DRIVER_MYSQL {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}, QueryTimeout: connectionTimeout}
	} else if *settings.DriverName == model.DATABASE_DRIVER_POSTGRES {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.PostgresDialect{}, QueryTimeout: connectionTimeout}
	} else {
		mlog.Critical("Failed to create dialect specific driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_NO_DRIVER)
	}

	return dbmap
}

func (ss *SqlSupplier) initConnection() {
	// Setup connection object for master
	ss.master = setupConnection("master", *ss.settings.DataSource, ss.settings)

	// Setup connection object to replicas
	if len(ss.settings.DataSourceReplicas) > 0 {
		ss.replicas = make([]*gorp.DbMap, len(ss.settings.DataSourceReplicas))
		for i, replica := range ss.settings.DataSourceReplicas {
			ss.replicas[i] = setupConnection(fmt.Sprintf("replica-%v", i), replica, ss.settings)
		}
	}

	if len(ss.settings.DataSourceSearchReplicas) > 0 {
		ss.searchReplicas = make([]*gorp.DbMap, len(ss.settings.DataSourceSearchReplicas))
		for i, replica := range ss.settings.DataSourceSearchReplicas {
			ss.searchReplicas[i] = setupConnection(fmt.Sprintf("search-replica-%v", i), replica, ss.settings)
		}
	}
}

func (ss *SqlSupplier) Close() {
	ss.master.Db.Close()
	for _, replica := range ss.replicas {
		replica.Db.Close()
	}
}

func (ss *SqlSupplier) DropAllTables() {
	ss.master.TruncateTables()
}

func (ss *SqlSupplier) MarkSystemRanUnitTests() {
	// props, err := ss.System().Get()
	// if err != nil {
	// 	return
	// }

	//TODO: Open this
	// unitTests := props[model.SYSTEM_RAN_UNIT_TESTS]
	// if len(unitTests) == 0 {
	// 	systemTests := &model.System{Name: model.SYSTEM_RAN_UNIT_TESTS, Value: "1"}
	// 	ss.System().Save(systemTests)
	// }
}

func (ss *SqlSupplier) getQueryBuilder() sq.StatementBuilderType {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}
	return builder
}

func (ss *SqlSupplier) GetAllConns() []*gorp.DbMap {
	all := make([]*gorp.DbMap, len(ss.replicas)+1)
	copy(all, ss.replicas)
	all[len(ss.replicas)] = ss.master
	return all
}

func NewSqlSupplier(settings model.SqlSettings, metrics einterfaces.MetricsInterface) *SqlSupplier {
	supplier := &SqlSupplier{
		rrCounter: 0,
		srCounter: 0,
		settings:  &settings,
	}

	supplier.initConnection()

	// Create tables if necessary
	supplier.stores.team = newSqlTeamStore(supplier)
	supplier.stores.user = newSqlUserStore(supplier)
	supplier.stores.bot = newSqlBotStore(supplier, metrics)
	supplier.stores.post = newSqlPostStore(supplier, metrics)
	supplier.stores.thread = newSqlThreadStore(supplier)
  supplier.stores.system = newSqlSystemStore(supplier)
  supplier.stores.status = newSqlStatusStore(supplier)
	supplier.stores.role = newSqlRoleStore(supplier)
	supplier.stores.scheme = newSqlSchemeStore(supplier)
	supplier.stores.channel = newSqlChannelStore(supplier)
	supplier.stores.session = newSqlSessionStore(supplier)
	supplier.stores.userAccessToken = newSqlUserAccessTokenStore(supplier)
	supplier.stores.token = newSqlTokenStore(supplier)
	supplier.stores.preference = newSqlPreferenceStore(supplier)
	supplier.stores.group = newSqlGroupStore(supplier)

	err := supplier.GetMaster().CreateTablesIfNotExists()
	if err != nil {
		mlog.Critical("Error creating database tables.", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_TABLE)
	}

	return supplier
}

// Check if the error is belong to MySQL or Postgres DB
func IsUniqueConstraintError(err error, indexName []string) bool {
	unique := false
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		unique = true
	}

	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
		unique = true
	}

	field := false
	for _, contain := range indexName {
		if strings.Contains(err.Error(), contain) {
			field = true
			break
		}
	}

	return unique && field
}

func (ss *SqlSupplier) System() store.SystemStore {
	return ss.stores.system
}

func (ss *SqlSupplier) Status() store.StatusStore {
	return ss.stores.status
}

func (ss *SqlSupplier) Team() store.TeamStore {
	return ss.stores.team
}

func (ss *SqlSupplier) Role() store.RoleStore {
	return ss.stores.role
}

func (ss *SqlSupplier) Scheme() store.SchemeStore {
	return ss.stores.scheme
}

func (ss *SqlSupplier) Channel() store.ChannelStore {
	return ss.stores.channel
}

func (ss *SqlSupplier) Session() store.SessionStore {
	return ss.stores.session
}

func (ss *SqlSupplier) UserAccessToken() store.UserAccessTokenStore {
	return ss.stores.userAccessToken
}

func (ss *SqlSupplier) Token() store.TokenStore {
	return ss.stores.token
}

func (ss *SqlSupplier) Preference() store.PreferenceStore {
	return ss.stores.preference
}

func (ss *SqlSupplier) Group() store.GroupStore {
	return ss.stores.group
}

func (ss *SqlSupplier) LockToMaster() {
	ss.lockedToMaster = true
}

func (ss *SqlSupplier) UnlockFromMaster() {
	ss.lockedToMaster = false
}
