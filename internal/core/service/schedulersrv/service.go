package schedulersrv

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/core/port"
	rtime "github.com/svaloumas/valet/pkg/time"
)

var _ port.Scheduler = &schedulerservice{}

type schedulerservice struct {
	jobQueue    port.JobQueue
	storage     port.Storage
	workService port.WorkService
	time        rtime.Time
	logger      *logrus.Logger
}

// New creates a new scheduler service.
func New(
	jobQueue port.JobQueue,
	storage port.Storage,
	workService port.WorkService,
	time rtime.Time,
	logger *logrus.Logger) *schedulerservice {

	return &schedulerservice{
		jobQueue:    jobQueue,
		storage:     storage,
		workService: workService,
		time:        time,
		logger:      logger,
	}
}

// Dispatch listens to the job queue for messages, consumes them and
// dispatches the work items to the worker pool for execution.
func (srv *schedulerservice) Dispatch(ctx context.Context, duration time.Duration) {
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

// Schedule polls the repository in given interval and schedules due jobs for execution.
func (srv *schedulerservice) Schedule(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				srv.logger.Info("exiting...")
				return
			case <-ticker.C:
				dueJobs, err := srv.storage.GetDueJobs()
				if err != nil {
					srv.logger.Errorf("could not get due jobs from repository: %s", err)
					continue
				}
				for _, j := range dueJobs {
					if j.BelongsToPipeline() {
						for job := j; job.HasNext(); job = job.Next {
							job.Next, err = srv.storage.GetJob(job.NextJobID)
							if err != nil {
								srv.logger.Errorf("could not get piped due job from repository: %s", err)
								continue
							}
						}
					}
					w := srv.workService.CreateWork(j)
					// Blocks until worker pool backlog has some space.
					srv.workService.Send(w)

					scheduledAt := srv.time.Now()
					j.MarkScheduled(&scheduledAt)
					if err := srv.storage.UpdateJob(j.ID, j); err != nil {
						srv.logger.Errorf("could not update job: %s", err)
					}
					message := fmt.Sprintf("job with ID: %s", j.ID)
					if j.BelongsToPipeline() {
						message = fmt.Sprintf("pipeline with ID: %s", j.PipelineID)
					}
					srv.logger.Infof("scheduled work for %s to worker pool", message)
				}
			}
		}
	}()
}
