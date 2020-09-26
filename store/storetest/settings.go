package storetest

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/masterhung0112/hk_server/model"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path"
)

const (
	defaultMysqlDSN      = "hkuser:mostest@tcp(localhost:7306)/hungknow_test?charset=utf8mb4,utf8\u0026readTimeout=30s\u0026writeTimeout=30s"
	defaultPostgresqlDSN = "postgres://hkuser:mostest@localhost:7432/hungknow_test?sslmode=disable&connect_timeout=10"
	defaultMysqlRootPWD  = "mostest"
)

func getEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	} else {
		return defaultValue
	}
}

func log(message string) {
	verbose := false
	if verboseFlag := flag.Lookup("test.v"); verboseFlag != nil {
		verbose = verboseFlag.Value.String() != ""
	}
	if verboseFlag := flag.Lookup("v"); verboseFlag != nil {
		verbose = verboseFlag.Value.String() != ""
	}

	if verbose {
		fmt.Println(message)
	}
}

func mySQLRootDSN(dsn string) string {
	rootPwd := getEnv("TEST_DATABASE_MYSQL_ROOT_PASSWD", defaultMysqlRootPWD)
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic("failed to parse dsn " + dsn + ": " + err.Error())
	}

	cfg.User = "root"
	cfg.Passwd = rootPwd
	cfg.DBName = "mysql"

	return cfg.FormatDSN()
}

func postgreSQLRootDSN(dsn string) string {
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		panic("failed to parse dsn " + dsn + ": " + err.Error())
	}

	// // Assume the unittesting database has the same password.
	// password := ""
	// if dsnUrl.User != nil {
	// 	password, _ = dsnUrl.User.Password()
	// }

	// dsnUrl.User = url.UserPassword("", password)
	dsnUrl.Path = "postgres"

	return dsnUrl.String()
}

// execAsRoot executes the given sql as root against the testing database
func execAsRoot(settings *model.SqlSettings, sqlCommand string) error {
	var dsn string
	var driver = *settings.DriverName

	switch driver {
	case model.DATABASE_DRIVER_MYSQL:
		dsn = mySQLRootDSN(*settings.DataSource)
	case model.DATABASE_DRIVER_POSTGRES:
		dsn = postgreSQLRootDSN(*settings.DataSource)
	default:
		return fmt.Errorf("unsupported driver %s", driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s database as root", driver)
	}
	defer db.Close()
	if _, err = db.Exec(sqlCommand); err != nil {
		return errors.Wrapf(err, "failed to execute `%s` against %s database as root", sqlCommand, driver)
	}

	return nil
}

func mySQLDSNDatabase(dsn string) string {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic("failed to parse dsn " + dsn + ": " + err.Error())
	}

	return cfg.DBName
}

func postgreSQLDSNDatabase(dsn string) string {
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		panic("failed to parse dsn " + dsn + ": " + err.Error())
	}

	return path.Base(dsnUrl.Path)
}

func CleanupSqlSettings(settings *model.SqlSettings) {
	var driver = *settings.DriverName
	var dbName string

	switch driver {
	case model.DATABASE_DRIVER_MYSQL:
		dbName = mySQLDSNDatabase(*settings.DataSource)
	case model.DATABASE_DRIVER_POSTGRES:
		dbName = postgreSQLDSNDatabase(*settings.DataSource)
	default:
		panic("unsupported driver " + driver)
	}

	if err := execAsRoot(settings, "DROP DATABASE "+dbName); err != nil {
		panic("failed to drop temporary database " + dbName + ": " + err.Error())
	}

	log("Dropped temporary database " + dbName)
}

func databaseSettings(driver, dataSource string) *model.SqlSettings {
	settings := &model.SqlSettings{
		DriverName:         &driver,
		DataSource:         &dataSource,
		DataSourceReplicas: []string{},
		QueryTimeout:       new(int),
	}

	*settings.QueryTimeout = 60

	return settings
}

// MySQLSettings returns the database settings to connect to the MySQL unit testing database.
// The database name is generated randomly and must be created before use.
func MySQLSettings() *model.SqlSettings {
	dsn := getEnv("TEST_DATABASE_MYSQL_DSN", defaultMysqlDSN)
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic("failed to parse dsn " + dsn + ": " + err.Error())
	}

	cfg.DBName = "hk_test_db" + model.NewId()

	return databaseSettings("mysql", cfg.FormatDSN())
}

// PostgresSQLSettings returns the database settings to connect to the PostgreSQL unittesting database.
// The database name is generated randomly and must be created before use.
func PostgreSQLSettings() *model.SqlSettings {
	dsn := getEnv("TEST_DATABASE_POSTGRESQL_DSN", defaultPostgresqlDSN)
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		panic("failed to parse dsn " + dsn + ": " + err.Error())
	}

	// Generate a random database name
	dsnUrl.Path = "hk_test_db" + model.NewId()

	return databaseSettings("postgres", dsnUrl.String())
}

// MakeSqlSettings created a random named database on Database and returns the corresponding sql settings
func MakeSqlSettings(driver string) *model.SqlSettings {
	var settings *model.SqlSettings
	var dbName string

	switch driver {
	case model.DATABASE_DRIVER_MYSQL:
		settings = MySQLSettings()
		dbName = mySQLDSNDatabase(*settings.DataSource)
	case model.DATABASE_DRIVER_POSTGRES:
		settings = PostgreSQLSettings()
		dbName = postgreSQLDSNDatabase(*settings.DataSource)
	default:
		panic("unsupported driver " + driver)
	}

	// Create new database on DB
	if err := execAsRoot(settings, "CREATE DATABASE "+dbName); err != nil {
		panic("failed to create temporary database " + dbName + ": " + err.Error())
	}

	// Grand permission
	switch driver {
	case model.DATABASE_DRIVER_MYSQL:
		if err := execAsRoot(settings, "GRANT ALL PRIVILEGES ON "+dbName+".* TO 'hkuser'"); err != nil {
			panic("failed to grant hkuser permission to " + dbName + ":" + err.Error())
		}
	case model.DATABASE_DRIVER_POSTGRES:
		if err := execAsRoot(settings, "GRANT ALL PRIVILEGES ON DATABASE \""+dbName+"\" TO hkuser"); err != nil {
			panic("failed to grant hkuser permission to " + dbName + ":" + err.Error())
		}
	default:
		panic("unsupported driver " + driver)
	}

	log("Created temporary " + driver + " database " + dbName)

	return settings
}
