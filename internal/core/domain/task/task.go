package task

import "fmt"

var (
	TaskTypes = map[string]TaskFunc{
		"dummytask": DummyTask,
	}
)

// TaskFunc is the type of the task callback.
type TaskFunc func(interface{}) (interface{}, error)

type TaskRepository struct {
	tasks map[string]TaskFunc
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[string]TaskFunc),
	}
}

func (repo *TaskRepository) Register(name string, taskFunc TaskFunc) {
	repo.tasks[name] = taskFunc
}

func (repo *TaskRepository) GetTaskFunc(name string) (TaskFunc, error) {
	task, ok := repo.tasks[name]
	if !ok {
		return nil, fmt.Errorf("task with name: %s is not registered", name)
	}
	return task, nil
}

func (repo *TaskRepository) GetNames() []string {
	names := []string{}
	for name := range repo.tasks {
		names = append(names, name)
	}
	return names
}
