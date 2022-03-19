package ports

import (
	"valet/internal/core/domain"
)

// JobRepository represents the driven actors interface.
type JobRepository interface {
	Create(*domain.Job) error
	Get(id string) (*domain.Job, error)
	Update(*domain.Job) error
	Delete(id string) error
}

// JobService represents the driver actors interface.
type JobService interface {
	Create(name, description string) error
	Get(id string) *domain.Job
	Delete(id string) error
}
