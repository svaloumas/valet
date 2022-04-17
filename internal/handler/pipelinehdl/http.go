package pipelinehdl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/pkg/apperrors"
)

// PipelineHTTPHandler is an HTTP handler that exposes pipeline endpoints.
type PipelineHTTPHandler struct {
	pipelineService port.PipelineService
	jobService      port.JobService
	jobQueue        port.JobQueue
}

// NewPipelineHTTPHandler creates and returns a new PipelineHTTPHandler.
func NewPipelineHTTPHandler(
	pipelineService port.PipelineService,
	jobServicce port.JobService,
	jobQueue port.JobQueue) *PipelineHTTPHandler {

	return &PipelineHTTPHandler{
		pipelineService: pipelineService,
		jobService:      jobServicce,
	}
}

// Create creates a new pipeline.
func (hdl *PipelineHTTPHandler) Create(c *gin.Context) {
	body := NewRequestBodyDTO()
	c.Bind(&body)

	jobs := make([]*domain.Job, 0)
	for _, jobDTO := range body.Jobs {
		j, err := hdl.jobService.Create(
			jobDTO.Name, jobDTO.TaskName, jobDTO.Description, body.RunAt, jobDTO.Timeout, jobDTO.TaskParams)
		if err != nil {
			switch err.(type) {
			case *apperrors.ResourceValidationErr:
				hdl.handleError(c, http.StatusBadRequest, err)
				return
			case *apperrors.ParseTimeErr:
				hdl.handleError(c, http.StatusBadRequest, err)
				return
			default:
				hdl.handleError(c, http.StatusInternalServerError, err)
				return
			}
		}
		jobs = append(jobs, j)
	}

	// TODO: Inside the service create make the linked list and the pipeline params to the job.
	// Mark the jobs as IsPiped as well.
	p, err := hdl.pipelineService.Create(body.Name, body.Description, body.RunAt, jobs)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			hdl.handleError(c, http.StatusBadRequest, err)
			return
		default:
			hdl.handleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	// Push only the first job the pipeline.
	if err := hdl.jobQueue.Push(p.Jobs[0]); err != nil {
		switch err.(type) {
		case *apperrors.FullQueueErr:
			hdl.handleError(c, http.StatusServiceUnavailable, err)
			return
		default:
			hdl.handleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(p))
}

func (hdl *PipelineHTTPHandler) handleError(c *gin.Context, code int, err error) {
	c.Error(err)
	c.AbortWithStatusJSON(code, gin.H{"error": true, "code": code, "message": err.Error()})
}
