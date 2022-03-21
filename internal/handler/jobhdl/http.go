package jobhdl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"

	"valet/internal/core/port"
	"valet/internal/repository"
)

type HTTPHandler struct {
	jobService   port.JobService
	taskMetadata port.Metadata
}

func NewHTTPHandler(jobService port.JobService, taskMetadata port.Metadata) *HTTPHandler {
	return &HTTPHandler{
		jobService:   jobService,
		taskMetadata: taskMetadata,
	}
}

func (hdl *HTTPHandler) Create(c *gin.Context) {
	body := NewBodyDTO(hdl.taskMetadata)
	c.BindJSON(&body)

	j, err := hdl.jobService.Create(body.Name, body.Description, body.Metadata)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, BuildResponseDTO(j))
}

func (hdl *HTTPHandler) Get(c *gin.Context) {
	j, err := hdl.jobService.Get(c.Param("id"))
	if err != nil && xerrors.Is(err, &repository.NotFoundError{}) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, BuildResponseDTO(j))
}

func (hdl *HTTPHandler) Update(c *gin.Context) {
	body := BodyDTO{}
	c.BindJSON(&body)

	err := hdl.jobService.Update(c.Param("id"), body.Name, body.Description)
	if err != nil && xerrors.Is(err, &repository.NotFoundError{}) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (hdl *HTTPHandler) Delete(c *gin.Context) {
	err := hdl.jobService.Delete(c.Param("id"))
	if err != nil {
		if xerrors.Is(err, &repository.NotFoundError{}) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
