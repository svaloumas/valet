package resulthdl

import (
	"net/http"

	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/pkg/apperrors"

	"github.com/gin-gonic/gin"
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
	c.JSON(http.StatusOK, BuildResponseBodyDTO(result))
}

// Delete deletes a job result.
func (hdl *ResultHTTPHandler) Delete(c *gin.Context) {
	err := hdl.resultService.Delete(c.Param("id"))
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

func (hdl *ResultHTTPHandler) handleError(c *gin.Context, code int, err error) {
	c.Error(err)
	c.AbortWithStatusJSON(code, gin.H{"error": true, "code": code, "message": err.Error()})
}
