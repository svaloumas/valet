package resulthdl

import (
	"github.com/svaloumas/valet/internal/core/domain"
)

// ResponseBodyDTO is the response data transfer object used for a job result retrieval.
type ResponseBodyDTO *domain.JobResult

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *domain.JobResult) ResponseBodyDTO {
	return ResponseBodyDTO(resource)
}
