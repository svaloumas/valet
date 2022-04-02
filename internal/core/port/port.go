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

	// Close liberates the bound resources of the job queue.
	Close()
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
	Exec(ctx context.Context, w domain.Work) error
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

// ConsumerService represents a domain event listener.
type ConsumerService interface {
	// Consume listens to the job queue for messages, consumes them and
	// schedules the job items for execution.
	Consume()

	// Stop terminates the job item service.
	Stop()
}
