package ports

import (
	"valet/internal/core/domain"
)

// JobRepository represents the driven actors interface.
type JobRepository interface {
	// Create creates a new job to the repository.
	Create(*domain.Job) error
	// Get fetches a job from the repository.
	Get(id string) (*domain.Job, error)
	// Update updates a job to the repository.
	Update(*domain.Job) error
	// Delete deletes a job from the repository.
	Delete(id string) error
}

// JobService represents the driver actors interface.
type JobService interface {
	// Create creates a new job.
	Create(name, description string) error
	// Get fetches a job.
	Get(id string) *domain.Job
	// Delete deletes a job.
	Delete(id string) error
}
