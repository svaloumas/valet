package taskrepo

import (
	"fmt"
)

// TaskFunc is the type of the task callback.
type TaskFunc func(interface{}) (interface{}, error)

// TaskRepository is the in memory task repository.
type TaskRepository map[string]TaskFunc

// New initializes and returns a new TaskRepository instance.
func New() *TaskRepository {
	repo := make(map[string]TaskFunc)
	taskrepo := TaskRepository(repo)
	return &taskrepo
}

// GetTaskFunc returns the TaskFunc for a specified name if that exists in the task repository.
func (repo TaskRepository) GetTaskFunc(name string) (TaskFunc, error) {
	task, ok := repo[name]
	if !ok {
		return nil, fmt.Errorf("task with name: %s is not registered", name)
	}
	return task, nil
}

// GetTaskNames returns all the names of the tasks currently in the task repository.
func (repo TaskRepository) GetTaskNames() []string {
	names := []string{}
	for name := range repo {
		names = append(names, name)
	}
	return names
}

// Register adds a new task in the repository.
func (repo TaskRepository) Register(name string, taskFunc TaskFunc) {
	repo[name] = taskFunc
}
