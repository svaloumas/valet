package consumersrv

import (
	"log"
	"valet/internal/core/port"
	wp "valet/internal/workerpool"
)

type consumerservice struct {
	jobQueue port.JobQueue
	wp       wp.WorkerPool
	done     chan struct{}
	logger   *log.Logger
}

// New creates a new job item service.
func New(jobQueue port.JobQueue, wp wp.WorkerPool, logger *log.Logger) *consumerservice {
	return &consumerservice{
		jobQueue: jobQueue,
		wp:       wp,
		done:     make(chan struct{}),
		logger:   logger,
	}
}

// Consume listens to the job queue for messages, consumes them and
// schedules the job items for execution.
func (srv *consumerservice) Consume() {
	srv.wp.Start()

	for {
		select {
		case <-srv.done:
			srv.logger.Println("exiting...")
			return
		case j := <-srv.jobQueue.Pop():
			srv.logger.Printf("sending job item for job with ID: %s to worker pool", j.ID)
			jobItem := srv.wp.CreateJobItem(j)
			err := srv.wp.Send(jobItem)
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
	srv.wp.Stop()
}
