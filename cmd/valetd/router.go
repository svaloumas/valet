package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"valet/internal/core/port"
	"valet/internal/handler/jobhdl"
	"valet/internal/handler/resulthdl"
)

// NewRouter initializes and returns a new gin.Engine instance.
func NewRouter(jobService port.JobService, resultService port.ResultService) *gin.Engine {
	jobHandhler := jobhdl.NewJobHTTPHandler(jobService)
	resultHandhler := resulthdl.NewResultHTTPHandler(resultService)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	r.POST("/jobs", jobHandhler.Create)
	r.GET("/jobs/:id", jobHandhler.Get)
	r.DELETE("/jobs/:id", jobHandhler.Delete)

	r.GET("/jobs/:id/results", resultHandhler.Get)
	r.DELETE("/jobs/:id/results", resultHandhler.Delete)
	return r
}
