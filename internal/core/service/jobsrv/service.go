package jobsrv

import (
	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"
)

var _ port.JobService = &jobservice{}

type jobservice struct {
	jobRepository port.JobRepository
	jobQueue      port.JobQueue
	taskrepo      *taskrepo.TaskRepository
	uuidGen       uuidgen.UUIDGenerator
	time          rtime.Time
}

// New creates a new job service.
func New(
	jobRepository port.JobRepository,
	jobQueue port.JobQueue,
	taskrepo *taskrepo.TaskRepository,
	uuidGen uuidgen.UUIDGenerator,
	time rtime.Time) *jobservice {
	return &jobservice{
		jobRepository: jobRepository,
		jobQueue:      jobQueue,
		taskrepo:      taskrepo,
		uuidGen:       uuidGen,
		time:          time,
	}
}

// Create creates a new job.
func (srv *jobservice) Create(
	name, taskName, description string,
	timeout int, taskParams interface{}) (*domain.Job, error) {

	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	createdAt := srv.time.Now()
	j := domain.NewJob(uuid, name, taskName, description, timeout, &createdAt, taskParams)

	if err := j.Validate(srv.taskrepo); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}
	if ok := srv.jobQueue.Push(j); !ok {
		return nil, &apperrors.FullQueueErr{}
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
