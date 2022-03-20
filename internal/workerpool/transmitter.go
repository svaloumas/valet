package workerpool

import (
	"log"
	"valet/internal/core/ports"
)

type jobtransmitter struct {
	jobQueue ports.JobQueue
	wp       ports.WorkerPool
}

// NewTransmitter initializes and returns a new transmitter concrete implementation.
func NewTransmitter(jobQueue ports.JobQueue, wp ports.WorkerPool) *jobtransmitter {
	return &jobtransmitter{
		jobQueue: jobQueue,
		wp:       wp,
	}
}

// Transmit transmits the job from the job queue to the worker pool.
func (t *jobtransmitter) Transmit() {
	for {
		j := t.jobQueue.Pop()
		// TODO: revisit this.
		err := t.wp.Send(j)
		if err != nil {
			if ok := t.jobQueue.Push(j); !ok {
				log.Printf("[transmitter] job queue is full - try again later")
			}
		}
	}
}
