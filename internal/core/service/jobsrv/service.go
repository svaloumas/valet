package jobsrv

import (
	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/internal/core/service"
	"valet/pkg/time"
	"valet/pkg/uuidgen"
)

type jobservice struct {
	jobRepository port.JobRepository
	jobQueue      port.JobQueue
	uuidGen       uuidgen.UUIDGenerator
	time          time.Time
}

// New creates a new job service.
func New(jobRepository port.JobRepository,
	jobQueue port.JobQueue,
	uuidGen uuidgen.UUIDGenerator,
	time time.Time) *jobservice {
	return &jobservice{
		jobRepository: jobRepository,
		jobQueue:      jobQueue,
		uuidGen:       uuidGen,
		time:          time,
	}
}

// Create creates a new job.
func (srv *jobservice) Create(name, description string, metadata interface{}) (*domain.Job, error) {
	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	createdAt := srv.time.Now()
	j := &domain.Job{
		ID:          uuid,
		Name:        name,
		Description: description,
		Metadata:    metadata,
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	if err := j.Validate(); err != nil {
		return nil, err
	}
	if ok := srv.jobQueue.Push(j); !ok {
		return nil, &service.FullQueueErr{}
	}
	if err := srv.jobRepository.Create(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get fetches a job.
func (srv *jobservice) Get(id string) (*domain.Job, error) {
	return srv.jobRepository.Get(id)
}

// Update updates a job.
func (srv *jobservice) Update(id, name, description string) error {
	j, err := srv.jobRepository.Get(id)
	if err != nil {
		return err
	}
	j.Name = name
	j.Description = description
	return srv.jobRepository.Update(id, j)
}

// Delete deletes a job.
func (srv *jobservice) Delete(id string) error {
	return srv.jobRepository.Delete(id)
}
