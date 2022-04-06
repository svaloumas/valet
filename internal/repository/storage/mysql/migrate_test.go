package mysql

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func TestCreateDBIfNotExists(t *testing.T) {

	mySQL := getTestMigrateMySQL()

	if mySQL.checkIfDBExists() {
		panic(fmt.Errorf("db: %s to be created already exists", mySQL.Config.DBName))
	}

	mySQL.CreateDBIfNotExists()
	defer mySQL.dropDB()

	if !mySQL.checkIfDBExists() {
		t.Errorf("db: %s could not be created", mySQL.Config.DBName)
	}
}

func TestMigrate(t *testing.T) {

	migrationsPath := "./migrations/*.sql"

	mySQL := getTestMigrateMySQL()

	mySQL.CreateDBIfNotExists()

	var err error
	mySQL.DB, err = sql.Open("mysql", mySQL.Config.FormatDSN())
	if err != nil {
		panic(err)
	}
	defer func() {
		mySQL.DB.Close()
		mySQL.dropDB()
	}()

	err = mySQL.Migrate(migrationsPath)
	if err != nil {
		t.Errorf("could not migrate db: %s", err)
	}

	files, err := filepath.Glob(migrationsPath)
	if err != nil {
		t.Errorf("could not fetch migration files: %s", err)
	}

	sql := "SELECT count(*) AS count FROM migrations;"

	rows, err := mySQL.DB.Query(sql)
	if err != nil {
		t.Errorf("could not get count of migrations in db: %s", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			t.Errorf("could not read count of migrations: %s", err)
		}
	}

	if len(files) != count {
		t.Errorf("expected %d in migrations table, got %d", len(files), count)
	}
}

func TestBrokenMigrationFailure(t *testing.T) {

	migrationsPath := "./testdata/migrations/*.sql"

	mySQL := getTestMigrateMySQL()
	mySQL.Config.DBName = "test_broken_migrations_db"

	mySQL.CreateDBIfNotExists()

	var err error
	mySQL.DB, err = sql.Open("mysql", mySQL.Config.FormatDSN())
	if err != nil {
		panic(err)
	}
	defer func() {
		mySQL.DB.Close()
		mySQL.dropDB()
	}()

	expectedErr := "You have an error in your SQL syntax"

	err = mySQL.Migrate(migrationsPath)
	if err == nil {
		t.Errorf("an error containing %q should have occured, got nil instead", expectedErr)
	}
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("an error containg %q should have occured, got %s instead", expectedErr, err.Error())
	}
}

func getTestMigrateMySQL() *MySQL {
	dsn := os.Getenv("MYSQL_DSN")
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic(err)
	}

	mySQL := new(MySQL)
	mySQL.Config = config
	mySQL.Config.DBName = "test_migrate_db"
	mySQL.Log = &logrus.Logger{Out: ioutil.Discard}

	return mySQL
}
