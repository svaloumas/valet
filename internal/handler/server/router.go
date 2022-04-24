package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/handler/jobhdl"
	"github.com/svaloumas/valet/internal/handler/pipelinehdl"
	"github.com/svaloumas/valet/internal/handler/resulthdl"
	"github.com/svaloumas/valet/internal/handler/taskhdl"
)

var (
	version = "v0.8.0"
)

// NewRouter initializes and returns a new gin.Engine instance.
func NewRouter(
	jobService port.JobService,
	resultService port.ResultService,
	pipelineService port.PipelineService,
	taskService port.TaskService,
	jobQueue port.JobQueue,
	storage port.Storage, loggingFormat string) *gin.Engine {

	jobHandler := jobhdl.NewJobHTTPHandler(jobService, jobQueue)
	resultHandler := resulthdl.NewResultHTTPHandler(resultService)
	pipelineHandler := pipelinehdl.NewPipelineHTTPHandler(pipelineService, jobQueue)
	taskHandler := taskhdl.NewTaskHTTPHandler(taskService)

	r := gin.New()
	if loggingFormat == "text" {
		r.Use(gin.Logger())
	} else {
		r.Use(JSONLogMiddleware())
	}
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	// CORS: Allow all origins - Revisit this.
	r.Use(cors.Default())

	r.GET("/api/status", HandleStatus(storage))

	r.POST("/api/jobs", jobHandler.Create)
	r.GET("/api/jobs", jobHandler.GetJobs)
	r.GET("/api/jobs/:id", jobHandler.Get)
	r.PATCH("/api/jobs/:id", jobHandler.Update)
	r.DELETE("/api/jobs/:id", jobHandler.Delete)

	r.GET("/api/jobs/:id/results", resultHandler.Get)
	r.DELETE("/api/jobs/:id/results", resultHandler.Delete)

	r.POST("/api/pipelines", pipelineHandler.Create)
	r.GET("/api/pipelines", pipelineHandler.GetPipelines)
	r.GET("/api/pipelines/:id", pipelineHandler.Get)
	r.PATCH("/api/pipelines/:id", pipelineHandler.Update)
	r.DELETE("/api/pipelines/:id", pipelineHandler.Delete)

	r.GET("/api/pipelines/:id/jobs", pipelineHandler.GetPipelineJobs)

	r.GET("/api/tasks", taskHandler.GetTasks)
	return r
}

// HandleStatus is an endpoint providing information and the status of the server,
// usually pinged for healthchecks by other services.
func HandleStatus(storage port.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		res := map[string]interface{}{
			"storage_healthy": storage.CheckHealth(),
			"time":            now,
			"version":         version,
		}
		c.JSON(http.StatusOK, res)
	}
}

// JSONLogMiddleware logs a gin HTTP request in JSON format.
func JSONLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		start := time.Now()
		// Process Request
		c.Next()
		// Stop timer
		elapsed := time.Since(start)

		entry := logrus.WithFields(logrus.Fields{
			"method":         c.Request.Method,
			"path":           c.Request.RequestURI,
			"status":         c.Writer.Status(),
			"referrer":       c.Request.Referer(),
			"request_id":     c.Writer.Header().Get("Request-Id"),
			"remote_address": c.Request.RemoteAddr,
			"elapsed":        elapsed.String(),
		})

		var message string
		status := c.Writer.Status()
		if status >= 400 {
			errResponse := struct {
				Code    int    `json:"code"`
				Error   bool   `json:"error"`
				Message string `json:"message"`
			}{}

			json.Unmarshal(blw.body.Bytes(), &errResponse)
			message = errResponse.Message
		}

		entry.Logger.Formatter = &logrus.JSONFormatter{}
		// TODO: Revisit this.
		if status >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.WithTime(start).Info(message)
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
