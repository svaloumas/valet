package jobsrv

import (
	"valet/internal/core/domain"
	"valet/internal/core/ports"
	"valet/pkg/time"
	"valet/pkg/uuidgen"
)

type service struct {
	jobRepository ports.JobRepository
	uuidGen       uuidgen.UUIDGenerator
	time          time.Time
}

// New creates a new job service.
func New(jobRepository ports.JobRepository,
	uuidGen uuidgen.UUIDGenerator,
	time time.Time) *service {
	return &service{
		jobRepository: jobRepository,
		uuidGen:       uuidGen,
		time:          time,
	}
}

// Create creates a new job.
func (srv *service) Create(name, description string) (*domain.Job, error) {
	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	createdAt := srv.time.Now()
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
