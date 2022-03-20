package ports

import (
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

// JobQueue represents a driven actor queue interface.
type JobQueue interface {
	// Push adds a job to the queue. Returns false if queue is full.
	Push(j *domain.Job) bool
	// Pop removes and returns the head job from the queue.
	Pop() *domain.Job
}

// WorkerPool represents a driven actor worker pool interface.
type WorkerPool interface {
	// Start starts a new worker pool, waiting for incoming jobs deliver to the workers.
	Start()
	// Stop signals the workers to stop working gracefully.
	Stop()
	// Send schedules the job. An error is returned if the job backlog is full.
	Send(j *domain.Job) error
}

type Task interface {
	// Run performs a given task.
	Run(j *domain.Job) ([]byte, error)
}

type Transmitter interface {
	// Transmit transmits the job from the job queue to the worker pool.
	Transmit()
}

// JobService represents a driver actor service interface.
type JobService interface {
	// Create creates a new job.
	Create(name, description string) (*domain.Job, error)
	// Get fetches a job.
	Get(id string) (*domain.Job, error)
	// Update updates a job.
	Update(id, name, description string) error
	// Delete deletes a job.
	Delete(id string) error
}
