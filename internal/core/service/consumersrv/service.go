package consumersrv

import (
	"context"
	"log"
	"time"

	"valet/internal/core/port"
)

var _ port.Consumer = &consumerservice{}

type consumerservice struct {
	jobQueue    port.JobQueue
	workService port.WorkService
	logger      *log.Logger
}

// New creates a new consumer service.
func New(
	jobQueue port.JobQueue,
	workService port.WorkService,
	logger *log.Logger) *consumerservice {

	return &consumerservice{
		jobQueue:    jobQueue,
		workService: workService,
		logger:      logger,
	}
}

// Consume listens to the job queue for messages, consumes them and
// schedules the job items for execution.
func (srv *consumerservice) Consume(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				srv.logger.Println("exiting...")
				return
			case <-ticker.C:
				j := srv.jobQueue.Pop()
				if j == nil {
					continue
				}
				w := srv.workService.CreateWork(j)
				// Blocks until worker pool backlog has some space.
				srv.workService.Send(w)
				srv.logger.Printf("sent work for job with ID: %s to worker pool", j.ID)
			}
		}
	}()
}
