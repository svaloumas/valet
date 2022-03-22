package resulthdl

import (
	"valet/internal/core/domain"
)

// ResponseDTO is the response data transfer object used for a job result retrieval.
type ResponseDTO *domain.JobResult

// BuildResponseDTO creates a new ResponseDTO.
func BuildResponseDTO(resource *domain.JobResult) ResponseDTO {
	return ResponseDTO(resource)
}
