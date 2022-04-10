package factory

import (
	"valet/internal/config"
	"valet/internal/core/port"
	"valet/internal/repository/storage/memorydb"
	"valet/internal/repository/storage/mysql"
)

func StorageFactory(cfg config.Repository) port.Storage {
	if cfg.Option == "memory" {
		return memorydb.NewMemoryDB()
	}

	options := &mysql.MySQLOptions{
		ConnectionMaxLifetime: cfg.MySQL.ConnectionMaxLifetime,
		MaxIdleConnections:    cfg.MySQL.MaxIdleConnections,
		MaxOpenConnections:    cfg.MySQL.MaxOpenConnections,
	}
	return mysql.New(cfg.MySQL.DSN, cfg.MySQL.CaPemFile, options)
}
