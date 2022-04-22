package handler

import (
	"github.com/gin-gonic/gin"
)

type HTTPHandler struct{}

func (hdl *HTTPHandler) HandleError(c *gin.Context, code int, err error) {
	c.Error(err)
	c.AbortWithStatusJSON(code, gin.H{"error": true, "code": code, "message": err.Error()})
}
