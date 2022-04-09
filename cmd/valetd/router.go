package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"valet/internal/core/port"
	"valet/internal/handler/jobhdl"
	"valet/internal/handler/resulthdl"
)

// NewRouter initializes and returns a new gin.Engine instance.
func NewRouter(
	jobService port.JobService,
	resultService port.ResultService,
	storage port.Storage, loggingFormat string) *gin.Engine {

	jobHandhler := jobhdl.NewJobHTTPHandler(jobService)
	resultHandhler := resulthdl.NewResultHTTPHandler(resultService)

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

	r.POST("/api/jobs", jobHandhler.Create)
	r.GET("/api/jobs/:id", jobHandhler.Get)
	r.PATCH("/api/jobs/:id", jobHandhler.Update)
	r.DELETE("/api/jobs/:id", jobHandhler.Delete)

	r.GET("/api/jobs/:id/results", resultHandhler.Get)
	r.DELETE("/api/jobs/:id/results", resultHandhler.Delete)
	return r
}

// HandleStatus is an endpoint providing information and the status of the server,
// usually pinged for healthchecks by other services.
func HandleStatus(storage port.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		res := map[string]interface{}{
			"build_time": buildTime,
			"commit":     commit,
			"time":       now,
			"version":    version,
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
