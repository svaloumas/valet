package tasksrv

import (
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
)

type taskservice struct {
	taskrepo *taskrepo.TaskRepository
}

// New creates a new job service.
func New() *taskservice {
	return &taskservice{
		taskrepo: taskrepo.New(),
	}
}

// Register registers a new task in the task repository.
func (srv *taskservice) Register(name string, taskFunc taskrepo.TaskFunc) {
	srv.taskrepo.Register(name, taskFunc)
}

// GetTaskRepository returns the task repository.
func (repo *taskservice) GetTaskRepository() *taskrepo.TaskRepository {
	return repo.taskrepo
}
