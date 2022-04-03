package jobhdl

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"valet/internal/core/port"
	"valet/pkg/apperrors"
)

// JobHTTPHandler is an HTTP handler that exposes job endpoints.
type JobHTTPHandler struct {
	jobService port.JobService
}

// NewJobHTTPHandler creates and returns a new JobHTTPHandler.
func NewJobHTTPHandler(jobService port.JobService) *JobHTTPHandler {
	return &JobHTTPHandler{
		jobService: jobService,
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
			hdl.handleError(c, http.StatusBadRequest, err)
			return
		case *apperrors.FullQueueErr:
			hdl.handleError(c, http.StatusServiceUnavailable, err)
			return
		default:
			hdl.handleError(c, http.StatusInternalServerError, err)
			return
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
			hdl.handleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.handleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(j))
}

// Update updates a job.
func (hdl *JobHTTPHandler) Update(c *gin.Context) {
	body := RequestBodyDTO{}
	c.BindJSON(&body)

	err := hdl.jobService.Update(c.Param("id"), body.Name, body.Description)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.handleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.handleError(c, http.StatusInternalServerError, err)
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
			hdl.handleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.handleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (hdl *JobHTTPHandler) handleError(c *gin.Context, code int, err error) {
	c.Error(err)
	c.AbortWithStatusJSON(code, gin.H{"error": true, "code": code, "message": err.Error()})
}
