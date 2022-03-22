package jobsrv

import (
	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/internal/repository/workerpool/task"
	"valet/pkg/apperrors"
	"valet/pkg/time"
	"valet/pkg/uuidgen"
)

var _ port.JobService = &jobservice{}

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

// Exec executes the job.
func (srv *jobservice) Exec(item domain.JobItem, callback task.TaskFunc) error {
	startedAt := srv.time.Now()
	item.Job.MarkStarted(&startedAt)
	if err := srv.jobRepository.Update(item.Job.ID, item.Job); err != nil {
		return err
	}

	// Perform the actual work.
	resultMetadata, jobErr := callback(item.Job.Metadata)

	select {
	case item.Result <- domain.JobResult{
		JobID:    item.Job.ID,
		Metadata: resultMetadata,
		Error:    jobErr}:
	default:
		// This should never happen as the result queue chan should be unique for this worker.
		panic("failed to write result to the result queue channel")
	}
	close(item.Result)

	if jobErr != nil {
		failedAt := srv.time.Now()
		item.Job.MarkFailed(&failedAt, jobErr.Error())
		if err := srv.jobRepository.Update(item.Job.ID, item.Job); err != nil {
			return err
		}
		return nil
	}

	completedAt := srv.time.Now()
	item.Job.MarkCompleted(&completedAt)
	if err := srv.jobRepository.Update(item.Job.ID, item.Job); err != nil {
		return err
	}
	return nil
}
