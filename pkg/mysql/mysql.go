package mysql

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// MySQLOptions represents config parameters for the MySQL client.
type MySQLOptions struct {
	ConnectionMaxLifetime int
	MaxIdleConnections    int
	MaxOpenConnections    int
}

// MySQL represents a MySQL DB client.
type MySQL struct {
	DB        *sql.DB
	Config    *mysql.Config
	Log       *logrus.Logger
	CaPemFile string
}

// New initializes and returns a MySQL client.
func New(
	dsn,
	caPemFile,
	migrationsPath string,
	options *MySQLOptions,
	logger *logrus.Logger) *MySQL {

	mySQL := new(MySQL)

	if logger == nil {
		mySQL.Log = &logrus.Logger{Out: ioutil.Discard}
	} else {
		mySQL.Log = logger
	}

	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic(err)
	}
	mySQL.Config = config

	if caPemFile != "" {
		rootCertPool := x509.NewCertPool()
		pem, _ := ioutil.ReadFile(caPemFile)
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			panic(fmt.Errorf("failed to append PEM: %s", caPemFile))
		}
		mysql.RegisterTLSConfig("custom", &tls.Config{RootCAs: rootCertPool})
		mySQL.Config.TLSConfig = "custom"
		mySQL.CaPemFile = caPemFile
	}

	mySQL.DB, err = sql.Open("mysql", config.FormatDSN())
	if err != nil {
		panic(err)
	}
	mySQL.DB.SetConnMaxLifetime(time.Duration(options.ConnectionMaxLifetime) * time.Millisecond)
	mySQL.DB.SetMaxIdleConns(options.MaxIdleConnections)
	mySQL.DB.SetMaxOpenConns(options.MaxOpenConnections)

	err = mySQL.CreateDBIfNotExists()
	if err != nil {
		panic(err)
	}

	err = mySQL.Migrate(migrationsPath)
	if err != nil {
		panic(err)
	}

	return mySQL
}

// CheckHealth returns the status of MySQL.
func (mySQL *MySQL) CheckHealth() bool {

	err := mySQL.DB.Ping()
	return err == nil
}

// Close terminates any store connections gracefully.
func (mySQL *MySQL) Close() error {
	return mySQL.DB.Close()
}
