package swagger

import (
	"valet/internal/core/domain"
)

// swagger:route GET /jobs/{id}/results jobs getJobResults
// Returns a specified job's results.
// responses:
//   200: getJobResultsResponse
//   404: errorResponse
//   500: errorResponse

// Returns a specified job results.
// swagger:response getJobResultsResponse
type getJobResultsResponse struct {
	// in:body
	Body domain.JobResult
}

// swagger:parameters getJobResultsRequestParams getJobResults
type getJobResultsRequestParams struct {
	// The ID of the specified job.
	//
	// in:path
	ID string `json:"id"`
}

// swagger:route DELETE /jobs/{id}/results jobs deleteJobResults
// Deletes a job's results.
// responses:
//   204:
//   404: errorResponse
//   500: errorResponse

// swagger:parameters deleteJobResultsRequestsParams deleteJobResults
type deleteJobResultsRequestsParams struct {
	// The ID of the specified job.
	//
	// in:path
	ID string `json:"id"`
}
