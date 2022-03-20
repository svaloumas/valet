package workerpool

type FullWorkerPoolBacklog struct{}

func (e *FullWorkerPoolBacklog) Error() string {
	return "worker pool backlog is full"
}
