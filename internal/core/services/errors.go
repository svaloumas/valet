package services

type FullQueueErr struct{}

func (e *FullQueueErr) Error() string {
	return "job queue is full - try again later"

}
