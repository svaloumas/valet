package factory

import (
	"valet/internal/config"
	"valet/internal/core/port"
	"valet/internal/repository/storage/memorydb"
	"valet/internal/repository/storage/mysql"
	"valet/pkg/log"
)

const (
	mysqlMigrationsPath = "./internal/repository/storage/mysql/migrations/*.sql"
)

func StorageFactory(cfg config.Repository, loggingFormat string) port.Storage {
	if cfg.Option == "memory" {
		return memorydb.NewMemoryDB()
	}

	options := &mysql.MySQLOptions{
		ConnectionMaxLifetime: cfg.MySQL.ConnectionMaxLifetime,
		MaxIdleConnections:    cfg.MySQL.MaxIdleConnections,
		MaxOpenConnections:    cfg.MySQL.MaxOpenConnections,
	}

	logger := log.NewLogger("mysql", loggingFormat)
	mySQL := mysql.New(
		cfg.MySQL.DSN, cfg.MySQL.CaPemFile, mysqlMigrationsPath, options, logger)
	return mySQL
}
