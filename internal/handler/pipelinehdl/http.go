package pipelinehdl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/handler"
	"github.com/svaloumas/valet/pkg/apperrors"
)

// PipelineHTTPHandler is an HTTP handler that exposes pipeline endpoints.
type PipelineHTTPHandler struct {
	handler.HTTPHandler
	pipelineService port.PipelineService
	jobQueue        port.JobQueue
}

// NewPipelineHTTPHandler creates and returns a new PipelineHTTPHandler.
func NewPipelineHTTPHandler(
	pipelineService port.PipelineService,
	jobQueue port.JobQueue) *PipelineHTTPHandler {

	return &PipelineHTTPHandler{
		pipelineService: pipelineService,
		jobQueue:        jobQueue,
	}
}

// Create creates a new pipeline and all of its jobs.
func (hdl *PipelineHTTPHandler) Create(c *gin.Context) {
	body := NewRequestBodyDTO()
	c.BindJSON(&body)

	jobs := make([]*domain.Job, 0)
	for _, jobDTO := range body.Jobs {
		j := &domain.Job{
			Name:               jobDTO.Name,
			Description:        jobDTO.Description,
			TaskName:           jobDTO.TaskName,
			Timeout:            jobDTO.Timeout,
			TaskParams:         jobDTO.TaskParams,
			UsePreviousResults: jobDTO.UsePreviousResults,
		}
		jobs = append(jobs, j)
	}

	p, err := hdl.pipelineService.Create(body.Name, body.Description, body.RunAt, jobs)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		case *apperrors.ParseTimeErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}

	if !p.IsScheduled() {
		// Push it as one job into the queue.
		p.MergeJobsInOne()

		// Push only the first job of the pipeline.
		if err := hdl.jobQueue.Push(p.Jobs[0]); err != nil {
			switch err.(type) {
			case *apperrors.FullQueueErr:
				hdl.HandleError(c, http.StatusServiceUnavailable, err)
				return
			default:
				hdl.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		}
		// Do not include next job in the response body.
		p.UnmergeJobs()
	}
	c.JSON(http.StatusAccepted, BuildResponseBodyDTO(p))
}

// Get fetches a pipeline.
func (hdl *PipelineHTTPHandler) Get(c *gin.Context) {
	j, err := hdl.pipelineService.Get(c.Param("id"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(j))
}

// GetPipelines fetches all pipelines, optionally filters them by status.
func (hdl *PipelineHTTPHandler) GetPipelines(c *gin.Context) {
	var status string
	value, ok := c.GetQuery("status")
	if ok {
		status = value
	}
	pipelines, err := hdl.pipelineService.GetPipelines(status)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	res := map[string]interface{}{
		"pipelines": pipelines,
	}
	c.JSON(http.StatusOK, res)
}

// GetPipelineJobs fetches the jobs of a specified pipeline.
func (hdl *PipelineHTTPHandler) GetPipelineJobs(c *gin.Context) {
	jobs, err := hdl.pipelineService.GetPipelineJobs(c.Param("id"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	res := map[string]interface{}{
		"jobs": jobs,
	}
	c.JSON(http.StatusOK, res)
}

// Update updates a pipeline.
func (hdl *PipelineHTTPHandler) Update(c *gin.Context) {
	body := RequestBodyDTO{}
	c.BindJSON(&body)

	err := hdl.pipelineService.Update(c.Param("id"), body.Name, body.Description)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Delete deletes a pipelines and all its jobs.
func (hdl *PipelineHTTPHandler) Delete(c *gin.Context) {
	err := hdl.pipelineService.Delete(c.Param("id"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
