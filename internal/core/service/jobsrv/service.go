package jobsrv

import (
	"time"
	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"
)

var _ port.JobService = &jobservice{}

type jobservice struct {
	storage  port.Storage
	jobQueue port.JobQueue
	taskrepo *taskrepo.TaskRepository
	uuidGen  uuidgen.UUIDGenerator
	time     rtime.Time
}

// New creates a new job service.
func New(
	storage port.Storage,
	jobQueue port.JobQueue,
	taskrepo *taskrepo.TaskRepository,
	uuidGen uuidgen.UUIDGenerator,
	time rtime.Time) *jobservice {
	return &jobservice{
		storage:  storage,
		jobQueue: jobQueue,
		taskrepo: taskrepo,
		uuidGen:  uuidGen,
		time:     time,
	}
}

// Create creates a new job.
func (srv *jobservice) Create(
	name, taskName, description, runAt string,
	timeout int, taskParams interface{}) (*domain.Job, error) {
	var runAtTime time.Time

	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	if runAt != "" {
		runAtTime, err = time.Parse(time.RFC3339Nano, runAt)
		if err != nil {
			return nil, &apperrors.ParseTimeErr{Message: err.Error()}
		}
	}
	createdAt := srv.time.Now()
	j := domain.NewJob(uuid, name, taskName, description, timeout, &runAtTime, &createdAt, taskParams)

	if err := j.Validate(srv.taskrepo); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}
	if runAtTime.IsZero() {
		if ok := srv.jobQueue.Push(j); !ok {
			return nil, &apperrors.FullQueueErr{}
		}
	}
	if err := srv.storage.CreateJob(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get fetches a job.
func (srv *jobservice) Get(id string) (*domain.Job, error) {
	return srv.storage.GetJob(id)
}

// Update updates a job.
func (srv *jobservice) Update(id, name, description string) error {
	j, err := srv.storage.GetJob(id)
	if err != nil {
		return err
	}
	j.Name = name
	j.Description = description
	return srv.storage.UpdateJob(id, j)
}

// Delete deletes a job.
func (srv *jobservice) Delete(id string) error {
	return srv.storage.DeleteJob(id)
}
