package pipelinehdl

import (
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/handler/jobhdl"
)

// RequestBodyDTO is the data transfer object used for a job creation or update.
type RequestBodyDTO struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	RunAt       string                   `json:"run_at"`
	Jobs        []*jobhdl.RequestBodyDTO `json:"jobs"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *RequestBodyDTO {
	return &RequestBodyDTO{}
}

// ResponseBodyDTO is the response data transfer object used for a pipeline creation or update.
type ResponseBodyDTO *domain.Pipeline

// BuildResponseDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *domain.Pipeline) ResponseBodyDTO {
	return ResponseBodyDTO(resource)
}
