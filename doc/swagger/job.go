package swagger

import (
	"valet/internal/core/domain"
	"valet/internal/handler/jobhdl"
)

// swagger:route POST /jobs jobs createJobs
// Creates a new job.
// responses:
//   200: jobResponse
//   500: errorResponse

// Returns the newly created job metadata.
// swagger:response jobResponse
type jobResponse struct {
	// in:body
	Body domain.Job
}

// swagger:parameters postJobsRequestParams createJobs
type postJobsRequestParams struct {
	// Name, description and metadata of the new job to be performed.
	// in:body
	Body jobhdl.RequestBodyDTO
}
