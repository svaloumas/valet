package jobqueue

import (
	"valet/internal/core/domain"
	"valet/internal/core/port"
)

var _ port.JobQueue = &fifoqueue{}

type fifoqueue struct {
	jobs     chan *domain.Job
	capacity int
}

// NewFIFOQueue creates and returns a new fifoqueue instance.
func NewFIFOQueue(capacity int) *fifoqueue {
	return &fifoqueue{
		jobs:     make(chan *domain.Job, capacity),
		capacity: capacity,
	}
}

// Push adds a job to the queue. Returns false if queue is full.
func (q *fifoqueue) Push(j *domain.Job) bool {
	select {
	case q.jobs <- j:
		return true
	default:
		return false
	}
}

// Pop removes and returns the head job from the queue.
func (q *fifoqueue) Pop() *domain.Job {
	select {
	case j := <-q.jobs:
		return j
	default:
		return nil
	}
}

// Close closes tha job queue channel.
func (q *fifoqueue) Close() {
	close(q.jobs)
}
