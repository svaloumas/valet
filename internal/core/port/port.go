package port

import (
	"context"
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/internal/core/service/worksrv/work"
)

// Storage represents a driven actor repository interface.
type Storage interface {
	// CreateJob adds a new job to the repository.
	CreateJob(j *domain.Job) error

	// GetJob fetches a job from the repository.
	GetJob(id string) (*domain.Job, error)

	// GetJobs fetches all jobs from the repository, optionally filters the jobs by status.
	GetJobs(status domain.JobStatus) ([]*domain.Job, error)

	// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
	GetDueJobs() ([]*domain.Job, error)

	// GetJobsByPipelineID fetches the jobs of the specified pipeline.
	GetJobsByPipelineID(pipelineID string) ([]*domain.Job, error)

	// UpdateJob updates a job to the repository.
	UpdateJob(id string, j *domain.Job) error

	// DeleteJob deletes a job from the repository.
	DeleteJob(id string) error

	// CreateJobResult adds a new job result to the repository.
	CreateJobResult(result *domain.JobResult) error

	// GetJobResult fetches a job result from the repository.
	GetJobResult(jobID string) (*domain.JobResult, error)

	// UpdateJobResult updates a job result to the repository.
	UpdateJobResult(jobID string, result *domain.JobResult) error

	// DeleteJobResult deletes a job result from the repository.
	DeleteJobResult(jobID string) error

	// CreatePipeline adds a new pipeline and of its jobs to the repository.
	CreatePipeline(p *domain.Pipeline) error

	// GetPipeline fetches a pipeline from the repository.
	GetPipeline(id string) (*domain.Pipeline, error)

	// GetPipelines fetches all pipelines from the repository, optionally filters the pipelines by status.
	GetPipelines(status domain.JobStatus) ([]*domain.Pipeline, error)

	// UpdatePipeline updates a pipeline to the repository.
	UpdatePipeline(id string, p *domain.Pipeline) error

	// DeletePipeline deletes a pipeline and all its jobs from the repository.
	DeletePipeline(id string) error

	// CheckHealth checks if the storage is alive.
	CheckHealth() bool

	// Close terminates any store connections gracefully.
	Close() error
}

// JobQueue represents a driven actor queue interface.
type JobQueue interface {
	// Push adds a job to the queue.
	Push(j *domain.Job) error

	// Pop removes and returns the head job from the queue.
	Pop() *domain.Job

	// CheckHealth checks if the job queue is alive.
	CheckHealth() bool

	// Close liberates the bound resources of the job queue.
	Close()
}

// JobService represents a driver actor service interface.
type JobService interface {
	// Create creates a new job.
	Create(name, taskName, description, runAt string, timeout int, taskParams map[string]interface{}) (*domain.Job, error)

	// Get fetches a job.
	Get(id string) (*domain.Job, error)

	// GetJobs fetches all jobs, optionally filters the jobs by status.
	GetJobs(status string) ([]*domain.Job, error)

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

// PipelineService represents a driver actor service interface.
type PipelineService interface {
	// Create creates a new pipeline.
	Create(name, description, runAt string, jobs []*domain.Job) (*domain.Pipeline, error)

	// Get fetches a pipeline.
	Get(id string) (*domain.Pipeline, error)

	// GetPipelines fetches all pipelines, optionally filters the pipelines by status.
	GetPipelines(status string) ([]*domain.Pipeline, error)

	// GetPipelineJobs fetches the jobs of a specified pipeline.
	GetPipelineJobs(id string) ([]*domain.Job, error)

	// Update updates a pipeline.
	Update(id, name, description string) error

	// Delete deletes a pipeline.
	Delete(id string) error
}

// WorkService represents a driver actor service interface.
type WorkService interface {
	// Start starts the worker pool.
	Start()

	// Stop signals the workers to stop working gracefully.
	Stop()

	// Dispatch dispatches a work to the worker pool.
	Dispatch(w work.Work)

	// CreateWork creates and return a new Work instance.
	CreateWork(j *domain.Job) work.Work

	// Exec executes a work.
	Exec(ctx context.Context, w work.Work) error
}

// TaskService represents a driver actor service interface.
type TaskService interface {
	// Register registers a new task in the task repository.
	Register(name string, taskFunc taskrepo.TaskFunc)

	// GetTaskRepository returns the task repository.
	GetTaskRepository() *taskrepo.TaskRepository
}

// Scheduler represents a domain event listener.
type Scheduler interface {
	// Schedule polls the repository in given interval and schedules due jobs for execution.
	Schedule(ctx context.Context, duration time.Duration)
	// Dispatch listens to the job queue for messages, consumes them and
	// dispatches the jobs for execution.
	Dispatch(ctx context.Context, duration time.Duration)
}

// Server represents a driver actor service interface.
type Server interface {
	// Serve start the server.
	Serve()

	// GracefullyStop gracefully stops the server.
	GracefullyStop()
}
