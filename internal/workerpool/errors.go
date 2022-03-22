package workerpool

// FullWorkerPoolBacklog is an error indicating that the
// the worker pool backlog queue is full.
type FullWorkerPoolBacklog struct{}

func (e *FullWorkerPoolBacklog) Error() string {
	return "worker pool backlog is full"
}
