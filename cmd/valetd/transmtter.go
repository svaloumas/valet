package main

import (
	"log"
	"os"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/port"
)

type jobtransmitter struct {
	jobQueue     port.JobQueue
	wp           port.WorkerPool
	done         chan struct{}
	tickInterval int
}

// NewTransmitter initializes and returns a new transmitter concrete implementation.
func NewTransmitter(
	jobQueue port.JobQueue,
	wp port.WorkerPool,
	tickInterval int) *jobtransmitter {

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
			// TODO: Implement a proper re-try mechanism.
			logger.Printf("sending job with ID: %s to worker pool", j.ID)
			resultChan := make(chan domain.JobResult, 1)
			// TODO: Consider making timeout unit configurable.
			jobItem := domain.JobItem{
				Job:         j,
				Result:      resultChan,
				TimeoutUnit: time.Second,
			}
			err := t.wp.Send(jobItem)
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

// Stop stops the transmitter.
func (t *jobtransmitter) Stop() {
	t.done <- struct{}{}
}
