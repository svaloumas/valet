package consumersrv

import (
	"log"

	"valet/internal/core/port"
)

var _ port.ConsumerService = &consumerservice{}

type consumerservice struct {
	jobQueue    port.JobQueue
	workService port.WorkService
	done        chan struct{}
	logger      *log.Logger
}

// New creates a new consumer service.
func New(jobQueue port.JobQueue, workService port.WorkService, logger *log.Logger) *consumerservice {
	return &consumerservice{
		jobQueue:    jobQueue,
		workService: workService,
		done:        make(chan struct{}),
		logger:      logger,
	}
}

// Consume listens to the job queue for messages, consumes them and
// schedules the job items for execution.
func (srv *consumerservice) Consume() {
	for {
		select {
		case <-srv.done:
			srv.logger.Println("exiting...")
			return
		case j := <-srv.jobQueue.Pop():
			w := srv.workService.CreateWork(j)
			// Blocks until worker pool backlog has some space.
			srv.workService.Send(w)
			srv.logger.Printf("sent work for job with ID: %s to worker pool", j.ID)
		}
	}
}

// Stop terminates the consumer service.
func (srv *consumerservice) Stop() {
	srv.done <- struct{}{}
}
