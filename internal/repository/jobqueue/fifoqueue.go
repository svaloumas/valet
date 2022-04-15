package jobqueue

import (
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/pkg/apperrors"
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

// Push adds a job to the queue.
func (q *fifoqueue) Push(j *domain.Job) error {
	select {
	case q.jobs <- j:
		return nil
	default:
		return &apperrors.FullQueueErr{}
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

// Close liberates the bound resources of the job queue.
func (q *fifoqueue) Close() {
	close(q.jobs)
}
