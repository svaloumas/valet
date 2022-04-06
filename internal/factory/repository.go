package factory

import (
	"valet/internal/config"
	"valet/internal/core/port"
	"valet/internal/repository/storage"
)

func StorageFactory(cfg config.Repository) port.Storage {
	if cfg.Option == "memory" {
		return storage.NewMemoryDB()
	}
	return nil
}
