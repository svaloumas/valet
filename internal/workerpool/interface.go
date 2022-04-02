package wp

import (
	"valet/internal/core/domain"
)

// WorkerPool represents a worker pool interface.
type WorkerPool interface {
	// Start starts the worker pool.
	Start()

	// Stop signals the workers to stop working gracefully.
	Stop()

	// Send schedules the job. An error is returned if the job backlog is full.
	Send(w domain.Work) error

	// CreateWork creates and return a new Work instance.
	CreateWork(j *domain.Job) domain.Work
}
