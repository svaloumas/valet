package tasksrv

import "github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"

type service struct {
	taskrepo *taskrepo.TaskRepository
}

// New creates a new job service.
func New() *service {
	return &service{
		taskrepo: taskrepo.New(),
	}
}

// Register registers a new task in the task repository.
func (srv *service) Register(name string, taskFunc taskrepo.TaskFunc) {
	srv.taskrepo.Register(name, taskFunc)
}

// GetTaskRepository returns the task repository.
func (repo *service) GetTaskRepository() *taskrepo.TaskRepository {
	return repo.taskrepo
}
