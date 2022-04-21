package jobhdl

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/handler"
	"github.com/svaloumas/valet/pkg/apperrors"
)

// JobHTTPHandler is an HTTP handler that exposes job endpoints.
type JobHTTPHandler struct {
	handler.HTTPHandler
	jobService port.JobService
	jobQueue   port.JobQueue
}

// NewJobHTTPHandler creates and returns a new JobHTTPHandler.
func NewJobHTTPHandler(jobService port.JobService, jobQueue port.JobQueue) *JobHTTPHandler {
	return &JobHTTPHandler{
		jobService: jobService,
		jobQueue:   jobQueue,
	}
}

// Create creates a new job.
func (hdl *JobHTTPHandler) Create(c *gin.Context) {
	body := NewRequestBodyDTO()
	c.BindJSON(&body)

	j, err := hdl.jobService.Create(
		body.Name, body.TaskName, body.Description, body.RunAt, body.Timeout, body.TaskParams)
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
	if !j.IsScheduled() {
		if err := hdl.jobQueue.Push(j); err != nil {
			switch err.(type) {
			case *apperrors.FullQueueErr:
				hdl.HandleError(c, http.StatusServiceUnavailable, err)
				return
			default:
				hdl.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		}
	}
	c.JSON(http.StatusAccepted, BuildResponseBodyDTO(j))
}

// Get fetches a job.
func (hdl *JobHTTPHandler) Get(c *gin.Context) {
	j, err := hdl.jobService.Get(c.Param("id"))
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

// GetJobs fetches all jobs, optionally filters them by status.
func (hdl *JobHTTPHandler) GetJobs(c *gin.Context) {
	var status string
	value, ok := c.GetQuery("status")
	if ok {
		status = value
	}
	jobs, err := hdl.jobService.GetJobs(status)
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
		"jobs": jobs,
	}
	c.JSON(http.StatusOK, res)
}

// Update updates a job.
func (hdl *JobHTTPHandler) Update(c *gin.Context) {
	body := RequestBodyDTO{}
	c.BindJSON(&body)

	err := hdl.jobService.Update(c.Param("id"), body.Name, body.Description)
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

// Delete deletes a job.
func (hdl *JobHTTPHandler) Delete(c *gin.Context) {
	err := hdl.jobService.Delete(c.Param("id"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		case *apperrors.CannotDeletePipelineJobErr:
			hdl.HandleError(c, http.StatusConflict, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
