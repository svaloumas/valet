package port

import (
	"context"
	"time"

	"valet/internal/core/domain"
)

// Storage represents a driven actor repository interface.
type Storage interface {
	// CreateJob adds new job to the repository.
	CreateJob(j *domain.Job) error

	// GetJob fetches a job from the repository.
	GetJob(id string) (*domain.Job, error)

	// UpdateJob updates a job to the repository.
	UpdateJob(id string, j *domain.Job) error

	// DeleteJob deletes a job from the repository.
	DeleteJob(id string) error

	// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
	GetDueJobs() ([]*domain.Job, error)

	// CreateJobResult adds a new job result to the repository.
	CreateJobResult(result *domain.JobResult) error

	// GetJobResult fetches a job result from the repository.
	GetJobResult(id string) (*domain.JobResult, error)

	// DeleteJobResult deletes a job result from the repository.
	DeleteJobResult(id string) error

	// CheckHealth checks if the storage is alive.
	CheckHealth() bool
}

// JobQueue represents a driven actor queue interface.
type JobQueue interface {
	// Push adds a job to the queue. Returns false if queue is full.
	Push(j *domain.Job) bool

	// Pop removes and returns the head job from the queue.
	Pop() *domain.Job

	// Close liberates the bound resources of the job queue.
	Close()
}

// JobService represents a driver actor service interface.
type JobService interface {
	// Create creates a new job.
	Create(name, taskName, description, runAt string, timeout int, taskParams map[string]interface{}) (*domain.Job, error)

	// Get fetches a job.
	Get(id string) (*domain.Job, error)

	// Update updates a job.
	Update(id, name, description string) error

	// Delete deletes a job.
	Delete(id string) error
}

// ResultService represents a driver actor service interface.
type ResultService interface {
	// Get fetches a job result.
	Get(id string) (*domain.JobResult, error)

	// Delete deletes a job result.
	Delete(id string) error
}

// WorkService represents a driver actor service interface.
type WorkService interface {
	// Start starts the worker pool.
	Start()

	// Stop signals the workers to stop working gracefully.
	Stop()

	// Send sends a work to the worker pool.
	Send(w domain.Work)

	// CreateWork creates and return a new Work instance.
	CreateWork(j *domain.Job) domain.Work

	// Exec executes a work.
	Exec(ctx context.Context, w domain.Work) error
}

type Consumer interface {
	// Consume listens to the job queue for messages, consumes them and
	// schedules the job items for execution.
	Consume(ctx context.Context, duration time.Duration)
}

// Scheduler represents a domain event listener.
type Scheduler interface {
	// Schedule polls the repository in given interval and schedules due jobs for execution.
	Schedule(ctx context.Context, duration time.Duration)
}
