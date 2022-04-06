package mysql

import (
	"os"
	"testing"
)

const (
	migrationsPath = "./migrations/*.sql"
)

// var (
// 	mysqlTest *MySQL
// 	db        *sql.DB
// )

// func TestMain(m *testing.M) {
// 	mysqlDSN := os.Getenv("MYSQL_DSN")
// 	mysqlTest = New(mysqlDSN, "", migrationsPath, &MySQLOptions{}, nil)

// 	var err error
// 	db, err = sql.Open("mysql", mysqlTest.Config.FormatDSN())
// 	if err != nil {
// 		panic(err)
// 	}

// 	defer func() {
// 		mysqlTest.DB.Close()
// 		db.Close()
// 		err := mysqlTest.dropDB()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()
// 	m.Run()
// }

func TestCheckHealth(t *testing.T) {
	mysqlDSN := os.Getenv("MYSQL_DSN")
	mysqlTest := New(mysqlDSN, "", migrationsPath, &MySQLOptions{}, nil)
	result := mysqlTest.CheckHealth()
	if result != true {
		t.Fatalf("expected true got %#v instead", result)
	}
}
