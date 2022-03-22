package jobhdl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"

	"valet/internal/core/port"
	"valet/internal/repository"
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
	body := NewBodyDTO()
	c.BindJSON(&body)

	j, err := hdl.jobService.Create(body.Name, body.Description, body.Metadata)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, BuildResponseDTO(j))
}

// Get fetches a job.
func (hdl *JobHTTPHandler) Get(c *gin.Context) {
	j, err := hdl.jobService.Get(c.Param("id"))
	if err != nil && xerrors.Is(err, &repository.NotFoundErr{}) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, BuildResponseDTO(j))
}

// Update updates a job.
func (hdl *JobHTTPHandler) Update(c *gin.Context) {
	body := BodyDTO{}
	c.BindJSON(&body)

	err := hdl.jobService.Update(c.Param("id"), body.Name, body.Description)
	if err != nil && xerrors.Is(err, &repository.NotFoundErr{}) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Delete deletes a job.
func (hdl *JobHTTPHandler) Delete(c *gin.Context) {
	err := hdl.jobService.Delete(c.Param("id"))
	if err != nil {
		if xerrors.Is(err, &repository.NotFoundErr{}) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
