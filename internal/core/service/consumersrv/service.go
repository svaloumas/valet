package consumersrv

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/core/port"
)

var _ port.Consumer = &consumerservice{}

type consumerservice struct {
	jobQueue    port.JobQueue
	workService port.WorkService
	logger      *logrus.Logger
}

// New creates a new consumer service.
func New(
	jobQueue port.JobQueue,
	workService port.WorkService,
	logger *logrus.Logger) *consumerservice {

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
				srv.logger.Info("exiting...")
				return
			case <-ticker.C:
				j := srv.jobQueue.Pop()
				if j == nil {
					continue
				}
				w := srv.workService.CreateWork(j)
				// Blocks until worker pool backlog has some space.
				srv.workService.Send(w)
				message := fmt.Sprintf("job with ID: %s", j.ID)
				if j.BelongsToPipeline() {
					message = fmt.Sprintf("pipeline with ID: %s", j.PipelineID)
				}
				srv.logger.Infof("sent work for %s to worker pool", message)
			}
		}
	}()
}
