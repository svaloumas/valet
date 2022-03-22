package service

// FullQueueErr is an error to indicate that a queue is full.
type FullQueueErr struct{}

func (e *FullQueueErr) Error() string {
	return "job queue is full - try again later"

}
