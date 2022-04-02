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
	srv.workService.Start()

	for {
		select {
		case <-srv.done:
			srv.logger.Println("exiting...")
			return
		case j := <-srv.jobQueue.Pop():
			srv.logger.Printf("sending work for job with ID: %s to worker pool", j.ID)
			w := srv.workService.CreateWork(j)
			err := srv.workService.Send(w)
			if err != nil {
				srv.logger.Printf("sending job with ID: %s to dead letter queue", j.ID)
				if ok := srv.jobQueue.Push(j); !ok {
					// TODO: Implement a proper re-try mechanism.
				}
			}
		}
	}
}

// Stop terminates the job item service.
func (srv *consumerservice) Stop() {
	srv.done <- struct{}{}
	srv.jobQueue.Close()
	srv.workService.Stop()
}
