package jobhdl

import (
	"valet/internal/core/domain"
)

// BodyDTO is the data transfer object used for a job creation or update.
type BodyDTO struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Metadata    interface{} `json:"metadata"`
}

// NewBodyDTO initializes and returns a new BodyDTO instance.
func NewBodyDTO() *BodyDTO {
	return &BodyDTO{}
}

// ResponseDTO is the response data transfer object used for a job creation or update.
type ResponseDTO *domain.Job

// BuildResponseDTO creates a new ResponseDTO.
func BuildResponseDTO(resource *domain.Job) ResponseDTO {
	return ResponseDTO(resource)
}
