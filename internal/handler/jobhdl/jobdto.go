package jobhdl

import (
	"valet/internal/core/domain"
)

// RequestBodyDTO is the data transfer object used for a job creation or update.
type RequestBodyDTO struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TaskType    string      `json:"task_type"`
	Timeout     int         `json:"timeout"`
	Metadata    interface{} `json:"metadata"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *RequestBodyDTO {
	return &RequestBodyDTO{}
}

// ResponseBodyDTO is the response data transfer object used for a job creation or update.
type ResponseBodyDTO *domain.Job

// BuildResponseDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *domain.Job) ResponseBodyDTO {
	return ResponseBodyDTO(resource)
}
