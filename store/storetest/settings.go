package storetest

import (
	"flag"
	"github.com/pkg/errors"
	"database/sql"
	"fmt"
	"os"
	"path"
	"net/url"
  "github.com/masterhung0112/go_server/model"
  "github.com/go-sql-driver/mysql"
)

const (
	defaultMysqlDSN      = "mmuser:mostest@tcp(localhost:3306)/mattermost_test?charset=utf8mb4,utf8\u0026readTimeout=30s\u0026writeTimeout=30s"
	defaultPostgresqlDSN = "postgres://mmuser:mostest@localhost:5432/mattermost_test?sslmode=disable&connect_timeout=10"
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