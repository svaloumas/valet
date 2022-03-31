package port

import (
	"context"

	"valet/internal/core/domain"
)

// JobRepository represents a driven actor repository interface.
type JobRepository interface {
	// Create adds new job to the repository.
	Create(j *domain.Job) error

	// Get fetches a job from the repository.
	Get(id string) (*domain.Job, error)

	// Update updates a job to the repository.
	Update(id string, j *domain.Job) error

	// Delete deletes a job from the repository.
	Delete(id string) error
}

// ResultRepository represents a driven actor repository interface.
type ResultRepository interface {
	// Create adds a new job result to the repository.
	Create(result *domain.JobResult) error

	// Get fetches a job result from the repository.
	Get(id string) (*domain.JobResult, error)

	// Delete deletes a job result from the repository.
	Delete(id string) error
}

// JobQueue represents a driven actor queue interface.
type JobQueue interface {
	// Push adds a job to the queue. Returns false if queue is full.
	Push(j *domain.Job) bool

	// Pop removes and returns the head job from the queue.
	Pop() <-chan *domain.Job
}

// WorkerPool represents a driven actor worker pool interface.
type WorkerPool interface {
	// Start starts the worker pool.
	Start()

	// Stop signals the workers to stop working gracefully.
	Stop()

	// Send schedules the job. An error is returned if the job backlog is full.
	Send(jobItem domain.JobItem) error

	// CreateJobItem creates and return a new JobItem instance.
	CreateJobItem(j *domain.Job) domain.JobItem
}

// JobService represents a driver actor service interface.
type JobService interface {
	// Create creates a new job.
	Create(name, taskName, description string, timeout int, metadata interface{}) (*domain.Job, error)

	// Get fetches a job.
	Get(id string) (*domain.Job, error)

	// Update updates a job.
	Update(id, name, description string) error

	// Delete deletes a job.
	Delete(id string) error

	// Exec executes a job.
	Exec(ctx context.Context, item domain.JobItem) error
}

// ResultService represents a driver actor service interface.
type ResultService interface {
	// Create waits until the result is available and
	// creates a new result in the repository.
	Create(futureResult domain.FutureJobResult) error

	// Get fetches a job result.
	Get(id string) (*domain.JobResult, error)

	// Delete deletes a job result.
	Delete(id string) error
}
