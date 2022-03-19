package jobhdl

import (
	"valet/internal/core/domain"
)

// CreateBody is the DTO for a job creation.
type CreateBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateResponse is the response DTO for a job creation.
type CreateResponse *domain.Job

// BuildCreateResponse creates a new CreateResponse.
func BuildCreateResponse(resource *domain.Job) CreateResponse {
	return CreateResponse(resource)
}
