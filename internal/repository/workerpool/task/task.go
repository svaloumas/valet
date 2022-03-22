package task

var (
	TaskTypes = map[string]TaskFunc{
		"dummytask": DummyTask,
	}
)

// TaskFunc is the type of the task callback.
type TaskFunc func(interface{}) ([]byte, error)
