package jobsrv

import (
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/ports"
	uuidgen "valet/pkg"
)

type service struct {
	jobRepository ports.JobRepository
	uuidGen       uuidgen.UUIDGen
}

// New creates a new job service.
func New(jobRepository ports.JobRepository) *service {
	return &service{jobRepository: jobRepository}
}

// Create creates a new job.
func (srv *service) Create(name, description string) (*domain.Job, error) {
	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	createdAt := time.Now()
	j := &domain.Job{
		ID:          uuid,
		Name:        name,
		Description: description,
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	if err := j.Validate(); err != nil {
		return nil, err
	}
	if err := srv.jobRepository.Create(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get fetches a job.
func (srv *service) Get(id string) (*domain.Job, error) {
	return srv.jobRepository.Get(id)
}

// Delete deletes a job.
func (srv *service) Delete(id string) error {
	return srv.jobRepository.Delete(id)
}
