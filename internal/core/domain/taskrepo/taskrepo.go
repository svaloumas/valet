package taskrepo

import (
	"fmt"
)

// TaskFunc is the type of the task callback.
type TaskFunc func(interface{}) (interface{}, error)

// TaskRepository is the in memory storage of tasks.
type TaskRepository struct {
	tasks map[string]TaskFunc
}

// NewTaskRepository creates and returns a new TaskRepository instance.
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[string]TaskFunc),
	}
}

// Register registers a new task in the repository.
func (repo *TaskRepository) Register(name string, taskFunc TaskFunc) {
	repo.tasks[name] = taskFunc
}

// GetTaskFunc returns the TaskFunc for a specified name if that exists in the repository.
func (repo *TaskRepository) GetTaskFunc(name string) (TaskFunc, error) {
	task, ok := repo.tasks[name]
	if !ok {
		return nil, fmt.Errorf("task with name: %s is not registered", name)
	}
	return task, nil
}

// GetTaskNames returns all the names of the tasks currently in the repository.
func (repo *TaskRepository) GetTaskNames() []string {
	names := []string{}
	for name := range repo.tasks {
		names = append(names, name)
	}
	return names
}
