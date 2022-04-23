package swagger

import (
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/handler/pipelinehdl"
)

// swagger:route POST /pipelines pipelines createPipelines
// Creates a new pipeline.
// responses:
//   202: postPipelinesResponse
//   400: errorResponse
//   500: errorResponse

// Returns the newly created pipeline metadata.
// swagger:response postPipelinesResponse
type postPipelinesResponse struct {
	// in:body
	Body domain.Pipeline
}

// swagger:parameters postPipelinesRequestParams createPipelines
type postPipelinesRequestParams struct {
	// Name, description and metadata of the new pipeline to be performed.
	// in:body
	Body pipelinehdl.RequestBodyDTO
}

// swagger:parameters getPipelineRequestParams getPipeline
type getPipelineRequestParams struct {
	// The ID of the specified pipeline.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}

// swagger:route GET /pipelines/{id} pipelines getPipeline
// Returns a specified pipeline.
// responses:
//   200: getPipelineResponse
//   404: errorResponse
//   500: errorResponse

// The specified pipeline's metadata.
// swagger:response getPipelineResponse
type getPipelineResponse struct {
	// in:body
	Body domain.Pipeline
}

// swagger:parameters getPipelinesRequestParams getPipelines
type getPipelinesRequestParams struct {
	// The status of the pipelines to be fetched..
	//
	// in:query
	// type:string
	// required:false
	Status string `json:"status"`
}

// swagger:route GET /pipelines pipelines getPipelines
// Returns all pipelines, optionally filters them by status.
// responses:
//   200: getPipelinesResponse
//   404: errorResponse
//   500: errorResponse

// The pipelines metadata.
// swagger:response getPipelinesResponse
type getPipelinesResponse struct {
	// in:body
	Body struct {
		Pipelines []domain.Pipeline `json:"pipelines"`
	}
}

// swagger:parameters getPipelineJobsRequestParams getPipelineJobs
type getPipelineJobsRequestParams struct {
	// The ID of the specified pipeline.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}

// swagger:route GET /pipelines/{id}/jobs pipelines getPipelineJobs
// Returns the jobs of the specified pipeline.
// responses:
//   200: getPipelineJobsResponse
//   404: errorResponse
//   500: errorResponse

// The pipeline jobs metadata.
// swagger:response getPipelineJobsResponse
type getPipelineJobsResponse struct {
	// in:body
	Body struct {
		Jobs []domain.Job `json:"jobs"`
	}
}

// swagger:route PATCH /pipelines/{id} pipelines patchPipeline
// Updates a pipeline's name or description.
// responses:
//   204:
//   404: errorResponse
//   500: errorResponse

// swagger:parameters patchPipelineRequestParams patchPipeline
type patchPipelineRequestParams struct {
	// The ID of the specified pipeline.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}

// swagger:route DELETE /pipelines/{id} pipelines deletePipeline
// Deletes a specified pipeline.
// responses:
//   204:
//   404: errorResponse
//   500: errorResponse

// swagger:parameters deletePipelineRequestParams deletePipeline
type deletePipelineRequestParams struct {
	// The ID of the specified pipeline.
	//
	// in:path
	// type: string
	ID string `json:"id"`
}
