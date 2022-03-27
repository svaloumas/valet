package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
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
	// CORS: Allow all origins - Revisit in production
	r.Use(cors.Default())

	r.POST("/api/jobs", jobHandhler.Create)
	r.GET("/api/jobs/:id", jobHandhler.Get)
	r.PATCH("/api/jobs/:id", jobHandhler.Update)
	r.DELETE("/api/jobs/:id", jobHandhler.Delete)

	r.GET("/api/jobs/:id/results", resultHandhler.Get)
	r.DELETE("/api/jobs/:id/results", resultHandhler.Delete)
	return r
}
