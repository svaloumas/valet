package task

// TaskFunc is the type of the task callback.
type TaskFunc func(interface{}) ([]byte, error)
