package factory

import (
	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/repository/storage/memorydb"
	"github.com/svaloumas/valet/internal/repository/storage/redis"
	"github.com/svaloumas/valet/internal/repository/storage/relational"
	"github.com/svaloumas/valet/internal/repository/storage/relational/mysql"
	"github.com/svaloumas/valet/internal/repository/storage/relational/postgres"
)

func StorageFactory(cfg config.Storage) port.Storage {
	if cfg.Option == "memory" {
		return memorydb.New()
	}
	if cfg.Option == "redis" {
		return redis.New(
			cfg.Redis.URL, cfg.Redis.PoolSize, cfg.Redis.MinIdleConns, cfg.Redis.KeyPrefix)
	}

	// Init common options for relational databases.
	options := new(relational.DBOptions)
	if cfg.Option == "postgres" {
		options.ConnectionMaxLifetime = cfg.Postgres.ConnectionMaxLifetime
		options.MaxOpenConnections = cfg.Postgres.MaxOpenConnections
		options.MaxIdleConnections = cfg.Postgres.MaxIdleConnections
		return postgres.New(cfg.Postgres.DSN, options)
	}
	options.ConnectionMaxLifetime = cfg.MySQL.ConnectionMaxLifetime
	options.MaxOpenConnections = cfg.MySQL.MaxOpenConnections
	options.MaxIdleConnections = cfg.MySQL.MaxIdleConnections
	return mysql.New(cfg.MySQL.DSN, cfg.MySQL.CaPemFile, options)
}
