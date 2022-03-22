package resultsrv

import (
	"valet/internal/core/domain"
	"valet/internal/core/port"
)

var _ port.ResultService = &resultservice{}

type resultservice struct {
	resultRepository port.ResultRepository
}

// New creates a new job result service.
func New(resultRepository port.ResultRepository) *resultservice {
	return &resultservice{
		resultRepository: resultRepository,
	}
}

// Create waits until the result is available and
// creates a new result in the repository.
func (srv *resultservice) Create(futureResult domain.FutureJobResult) error {
	result := futureResult.Wait()
	return srv.resultRepository.Create(&result)
}

// Get fetches a job result.
func (srv *resultservice) Get(id string) (*domain.JobResult, error) {
	return srv.resultRepository.Get(id)
}

// Delete deletes a job result.
func (srv *resultservice) Delete(id string) error {
	return srv.resultRepository.Delete(id)
}
