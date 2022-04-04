package schedulersrv

import (
	"context"
	"log"
	"time"

	"valet/internal/core/port"
	rtime "valet/pkg/time"
)

var _ port.Scheduler = &schedulerservice{}

type schedulerservice struct {
	jobRepository port.JobRepository
	workService   port.WorkService
	time          rtime.Time
	logger        *log.Logger
}

// New creates a new scheduler service.
func New(
	jobRepository port.JobRepository,
	workService port.WorkService,
	time rtime.Time,
	logger *log.Logger) *schedulerservice {

	return &schedulerservice{
		jobRepository: jobRepository,
		workService:   workService,
		time:          time,
		logger:        logger,
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
				srv.logger.Println("exiting...")
				return
			case <-ticker.C:
				dueJobs, err := srv.jobRepository.GetDueJobs()
				if err != nil {
					srv.logger.Printf("could not get due jobs from repository: %s", err)
					continue
				}
				for _, j := range dueJobs {
					w := srv.workService.CreateWork(j)
					// Blocks until worker pool backlog has some space.
					srv.workService.Send(w)

					scheduledAt := srv.time.Now()
					j.ScheduledAt = &scheduledAt
					if err := srv.jobRepository.Update(j.ID, j); err != nil {
						srv.logger.Printf("could not update job: %s", err)
					}
					srv.logger.Printf("scheduled work for job with ID: %s to worker pool", j.ID)
				}
			}
		}
	}()
}
