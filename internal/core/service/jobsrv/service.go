package jobsrv

import (
	"context"
	"fmt"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/domain/task"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"
)

var _ port.JobService = &jobservice{}
var defaultJobTimeout time.Duration = 84600

type jobservice struct {
	jobRepository port.JobRepository
	jobQueue      port.JobQueue
	taskrepo      *task.TaskRepository
	uuidGen       uuidgen.UUIDGenerator
	time          rtime.Time
}

// New creates a new job service.
func New(jobRepository port.JobRepository,
	jobQueue port.JobQueue,
	taskrepo *task.TaskRepository,
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
	timeout int, metadata interface{}) (*domain.Job, error) {

	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	createdAt := srv.time.Now()
	j := domain.NewJob(uuid, name, taskName, description, timeout, &createdAt, metadata)

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

// Exec executes the job.
func (srv *jobservice) Exec(ctx context.Context, item domain.JobItem) error {
	// Should be already validated.
	taskFunc, _ := srv.taskrepo.GetTaskFunc(item.Job.TaskName)
	startedAt := srv.time.Now()
	item.Job.MarkStarted(&startedAt)
	if err := srv.jobRepository.Update(item.Job.ID, item.Job); err != nil {
		return err
	}
	timeout := defaultJobTimeout
	if item.Job.Timeout > 0 && item.Job.Timeout <= 84600 {
		timeout = time.Duration(item.Job.Timeout)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout*item.TimeoutUnit)
	defer cancel()

	jobResultChan := make(chan domain.JobResult, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				result := domain.JobResult{
					JobID:    item.Job.ID,
					Metadata: nil,
					Error:    fmt.Errorf("%v", p).Error(),
				}
				jobResultChan <- result
			}
		}()
		var errMsg string

		// Perform the actual work.
		resultMetadata, jobErr := taskFunc(item.Job.Metadata)
		if jobErr != nil {
			errMsg = jobErr.Error()
		}

		result := domain.JobResult{
			JobID:    item.Job.ID,
			Metadata: resultMetadata,
			Error:    errMsg,
		}
		jobResultChan <- result
		close(jobResultChan)
	}()

	var jobResult domain.JobResult
	select {
	case <-ctx.Done():
		failedAt := srv.time.Now()
		item.Job.MarkFailed(&failedAt, ctx.Err().Error())

		jobResult = domain.JobResult{
			JobID:    item.Job.ID,
			Metadata: nil,
			Error:    ctx.Err().Error(),
		}
	case jobResult = <-jobResultChan:
		if jobResult.Error != "" {
			failedAt := srv.time.Now()
			item.Job.MarkFailed(&failedAt, jobResult.Error)
		} else {
			completedAt := srv.time.Now()
			item.Job.MarkCompleted(&completedAt)
		}
	}
	if err := srv.jobRepository.Update(item.Job.ID, item.Job); err != nil {
		return err
	}
	item.Result <- jobResult
	close(item.Result)

	return nil
}
