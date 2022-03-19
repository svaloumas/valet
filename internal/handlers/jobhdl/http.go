package jobhdl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"

	"valet/internal/core/ports"
	"valet/internal/repositories"
)

type HTTPHandler struct {
	jobService ports.JobService
}

func NewHTTPHandler(jobService ports.JobService) *HTTPHandler {
	return &HTTPHandler{
		jobService: jobService,
	}
}

func (hdl *HTTPHandler) Create(c *gin.Context) {
	body := CreateBody{}
	c.BindJSON(&body)

	j, err := hdl.jobService.Create(body.Name, body.Description)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, BuildCreateResponse(j))
}

func (hdl *HTTPHandler) Get(c *gin.Context) {
	j, err := hdl.jobService.Get(c.Param("id"))
	if err != nil {
		if xerrors.Is(err, &repositories.NotFoundError{}) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, j)
}

func (hdl *HTTPHandler) Delete(c *gin.Context) {
	err := hdl.jobService.Delete(c.Param("id"))
	if err != nil {
		if xerrors.Is(err, &repositories.NotFoundError{}) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
