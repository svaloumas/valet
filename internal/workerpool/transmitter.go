package workerpool

import (
	"log"
	"os"
	"time"
	"valet/internal/core/port"
)

type jobtransmitter struct {
	jobQueue     port.JobQueue
	wp           port.WorkerPool
	done         chan struct{}
	tickInterval int
}

// NewTransmitter initializes and returns a new transmitter concrete implementation.
func NewTransmitter(jobQueue port.JobQueue, wp port.WorkerPool, tickInterval int) *jobtransmitter {
	return &jobtransmitter{
		jobQueue:     jobQueue,
		wp:           wp,
		done:         make(chan struct{}),
		tickInterval: tickInterval,
	}
}

// Transmit transmits the job from the job queue to the worker pool.
func (t *jobtransmitter) Transmit() {
	logger := log.New(os.Stderr, "[transmitter] ", log.LstdFlags)
	ticker := time.NewTicker(time.Duration(t.tickInterval))
	for {
		select {
		case <-t.done:
			logger.Printf("exiting...")
			return
		case j := <-t.jobQueue.Pop():
			// TODO: revisit this.
			logger.Printf("sending job with ID: %s to worker pool", j.ID)
			err := t.wp.Send(j)
			if err != nil {
				logger.Print("worker pool backlog is full, pushing job back to queue")
				if ok := t.jobQueue.Push(j); !ok {
					logger.Print("job queue is full - try again later")
				}
			}
		case <-ticker.C:
		}
	}
}

func (t *jobtransmitter) Stop() {
	t.done <- struct{}{}
}
