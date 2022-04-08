package schedulersrv

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"valet/internal/core/port"
	rtime "valet/pkg/time"
)

var _ port.Scheduler = &schedulerservice{}

type schedulerservice struct {
	storage     port.Storage
	workService port.WorkService
	time        rtime.Time
	logger      *logrus.Logger
}

// New creates a new scheduler service.
func New(
	storage port.Storage,
	workService port.WorkService,
	time rtime.Time,
	logger *logrus.Logger) *schedulerservice {

	return &schedulerservice{
		storage:     storage,
		workService: workService,
		time:        time,
		logger:      logger,
	}
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
					w := srv.workService.CreateWork(j)
					// Blocks until worker pool backlog has some space.
					srv.workService.Send(w)

					scheduledAt := srv.time.Now()
					j.MarkScheduled(&scheduledAt)
					if err := srv.storage.UpdateJob(j.ID, j); err != nil {
						srv.logger.Errorf("could not update job: %s", err)
					}
					srv.logger.Infof("scheduled work for job with ID: %s to worker pool", j.ID)
				}
			}
		}
	}()
}
