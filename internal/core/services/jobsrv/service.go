package jobsrv

import (
	"time"
	"valet/internal/core/domain"
	"valet/internal/core/ports"
)

type service struct {
	jobRepository ports.JobRepository
	// Implement this.
	uuidGen string
}

// New creates a new job service.
func New() *service {
	return &service{}
}

// Create creates a new job.
func (srv *service) Create(name, description string) error {
	createdAt := time.Now()
	job := &domain.Job{
		ID:          srv.uuidGen,
		Name:        name,
		Description: description,
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	if err := job.Validate(); err != nil {
		return err
	}
	return srv.jobRepository.Create(job)
}

// Get fetches a job.
func (srv *service) Get(id string) (*domain.Job, error) {
	return srv.jobRepository.Get(id)
}

// Delete deletes a job.
func (srv *service) Delete(id string) error {
	return srv.jobRepository.Delete(id)
}
