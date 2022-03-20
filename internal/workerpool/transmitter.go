package workerpool

import (
	"log"
	"os"
	"time"
	"valet/internal/core/ports"
)

type jobtransmitter struct {
	jobQueue     ports.JobQueue
	wp           ports.WorkerPool
	done         chan struct{}
	tickInterval int
}

// NewTransmitter initializes and returns a new transmitter concrete implementation.
func NewTransmitter(jobQueue ports.JobQueue, wp ports.WorkerPool, tickInterval int) *jobtransmitter {
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
			err := t.wp.Send(j)
			if err != nil {
				if ok := t.jobQueue.Push(j); !ok {
					logger.Printf("job queue is full - try again later")
				}
			}
		case <-ticker.C:
		}
	}
}

func (t *jobtransmitter) Stop() {
	t.done <- struct{}{}
}
