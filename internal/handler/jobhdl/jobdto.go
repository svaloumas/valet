package jobhdl

import (
	"valet/internal/core/domain"
	"valet/internal/core/port"
)

// BodyDTO is the data transfer object used for a job creation or update.
type BodyDTO struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Metadata    port.Metadata `json:"metadata"`
}

func NewBodyDTO(metadata port.Metadata) *BodyDTO {
	return &BodyDTO{
		Metadata: metadata,
	}
}

// ResponseDTO is the response data transfer object used for a job creation or update.
type ResponseDTO *domain.Job

// BuildResponseDTO creates a new ResponseDTO.
func BuildResponseDTO(resource *domain.Job) ResponseDTO {
	return ResponseDTO(resource)
}
