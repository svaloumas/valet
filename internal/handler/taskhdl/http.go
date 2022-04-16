package taskhdl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svaloumas/valet/internal/core/port"
)

// TaskHTTPHandler is an HTTP handler that exposes task endpoints.
type TaskHTTPHandler struct {
	taskService port.TaskService
}

// NewTaskHTTPHandler creates and returns a new JobHTTPHandler.
func NewTaskHTTPHandler(taskService port.TaskService) *TaskHTTPHandler {
	return &TaskHTTPHandler{
		taskService: taskService,
	}
}

// GetTasks returns all registered tasks.
func (hdl *TaskHTTPHandler) GetTasks(c *gin.Context) {
	taskrepo := hdl.taskService.GetTaskRepository()
	tasks := taskrepo.GetTaskNames()
	res := map[string]interface{}{
		"tasks": tasks,
	}
	c.JSON(http.StatusOK, res)
}
