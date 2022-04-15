package resultsrv

import (
	"valet/internal/core/domain"
	"valet/internal/core/port"
)

var _ port.ResultService = &resultservice{}

type resultservice struct {
	storage port.Storage
}

// New creates a new job result service.
func New(storage port.Storage) *resultservice {
	return &resultservice{
		storage: storage,
	}
}

// Get fetches a job result.
func (srv *resultservice) Get(id string) (*domain.JobResult, error) {
	return srv.storage.GetJobResult(id)
}

// Delete deletes a job result.
func (srv *resultservice) Delete(id string) error {
	_, err := srv.storage.GetJobResult(id)
	if err != nil {
		return err
	}
	return srv.storage.DeleteJobResult(id)
}
