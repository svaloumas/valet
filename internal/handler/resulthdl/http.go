package resulthdl

import (
	"net/http"
	"valet/internal/core/port"
	"valet/pkg/apperrors"

	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"
)

// ResultHTTPHandler is an HTTP handler that exposes result endpoints.
type ResultHTTPHandler struct {
	resultService port.ResultService
}

// NewResultHTTPHandler creates and returns a new ResultHTTPHandler.
func NewResultHTTPHandler(resultService port.ResultService) *ResultHTTPHandler {
	return &ResultHTTPHandler{
		resultService: resultService,
	}
}

// Get fetches a job result.
func (hdl *ResultHTTPHandler) Get(c *gin.Context) {
	result, err := hdl.resultService.Get(c.Param("id"))
	if err != nil && xerrors.Is(err, &apperrors.NotFoundErr{}) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(result))
}

// Delete deletes a job result.
func (hdl *ResultHTTPHandler) Delete(c *gin.Context) {
	err := hdl.resultService.Delete(c.Param("id"))
	if err != nil {
		if xerrors.Is(err, &apperrors.NotFoundErr{}) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
