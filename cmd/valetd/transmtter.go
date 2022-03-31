package main

import (
	"log"

	"valet/internal/core/port"
)

type jobtransmitter struct {
	jobQueue port.JobQueue
	wp       port.WorkerPool
	logger   *log.Logger
	done     chan struct{}
}

// NewTransmitter initializes and returns a new transmitter concrete implementation.
func NewTransmitter(
	jobQueue port.JobQueue,
	wp port.WorkerPool,
	logger *log.Logger) *jobtransmitter {

	return &jobtransmitter{
		jobQueue: jobQueue,
		wp:       wp,
		logger:   logger,
		done:     make(chan struct{}),
	}
}

// Transmit transmits the job from the job queue to the worker pool.
func (t *jobtransmitter) Transmit() {
	for {
		select {
		case <-t.done:
			t.logger.Printf("exiting...")
			return
		case j := <-t.jobQueue.Pop():
			t.logger.Printf("sending job item with job ID: %s to worker pool", j.ID)
			jobItem := t.wp.CreateJobItem(j)
			// TODO: Implement a proper re-try mechanism.
			err := t.wp.Send(jobItem)
			if err != nil {
				t.logger.Print("worker pool backlog is full, pushing job back to queue")
				if ok := t.jobQueue.Push(j); !ok {
					t.logger.Print("job queue is full - try again later")
				}
			}
		}
	}
}

// Stop stops the transmitter.
func (t *jobtransmitter) Stop() {
	t.done <- struct{}{}
}
