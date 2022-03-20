package jobqueue

import (
	"log"
	"os"

	"valet/internal/core/domain"
)

type fifoqueue struct {
	jobs     chan *domain.Job
	capacity int
	logger   *log.Logger
}

// NewFIFOQueue creates and returns a new fifoqueue instance.
func NewFIFOQueue(capacity int) *fifoqueue {
	logger := log.New(os.Stderr, "[job-queue] ", log.LstdFlags)
	return &fifoqueue{
		jobs:     make(chan *domain.Job, capacity),
		capacity: capacity,
		logger:   logger,
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
func (q *fifoqueue) Pop() <-chan *domain.Job {
	return q.jobs
}

func (q *fifoqueue) Close() {
	close(q.jobs)
	q.logger.Println("exiting...")
}
