package swagger

import (
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/handler/jobhdl"
)

// swagger:route POST /jobs jobs createJobs
// Creates a new job.
// responses:
//   202: postJobsResponse
//   400: errorResponse
//   500: errorResponse

// Returns the newly created job metadata.
// swagger:response postJobsResponse
type postJobsResponse struct {
	// in:body
	Body domain.Job
}

// swagger:parameters postJobsRequestParams createJobs
type postJobsRequestParams struct {
	// Name, description and metadata of the new job to be performed.
	// in:body
	Body jobhdl.RequestBodyDTO
}

// swagger:route GET /jobs/{id} jobs getJob
// Returns a specified job.
// responses:
//   200: getJobResponse
//   404: errorResponse
//   500: errorResponse

// The specified job's metadata.
// swagger:response getJobResponse
type getJobResponse struct {
	// in:body
	Body domain.Job
}

// swagger:parameters getJobsRequestParams getJobs
type getJobsRequestParams struct {
	// The status of the jobs to be fetched..
	//
	// in:query
	// type:string
	// required:false
	Status string `json:"status"`
}

// swagger:route GET /jobs jobs getJobs
// Returns all jobs, optionally filters them by status.
// responses:
//   200: getJobsResponse
//   404: errorResponse
//   500: errorResponse

// The jobs metadata.
// swagger:response getJobsResponse
type getJobsResponse struct {
	// in:body
	Body struct {
		Jobs []domain.Job `json:"jobs"`
	}
}

// swagger:parameters getJobRequestParams getJob
type getJobRequestParams struct {
	// The ID of the specified job.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}

// swagger:route PATCH /jobs/{id} jobs patchJob
// Updates a job's name or description.
// responses:
//   204:
//   404: errorResponse
//   500: errorResponse

// swagger:parameters patchJobRequestParams patchJob
type patchJobRequestParams struct {
	// The ID of the specified job.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}

// swagger:route DELETE /jobs/{id} jobs deleteJob
// Deletes a specified job.
// responses:
//   204:
//   404: errorResponse
//   500: errorResponse

// swagger:parameters deleteJobRequestParams deleteJob
type deleteJobRequestParams struct {
	// The ID of the specified job.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}
