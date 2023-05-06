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

func (srv *schedulerservice) NewSteward(timeout time.Duration, startSchedulerFn port.StartGoroutineFn) port.StartGoroutineFn {
	return func(ctx context.Context, pulseInterval time.Duration) <-chan struct{} {
		heartbeat := make(chan struct{})
		go func() {
			defer close(heartbeat)

			var wardCtx context.Context
			var cancel func()
			var wardHeartbeat <-chan struct{}
			startWard := func() {
				wardCtx, cancel = context.WithCancel(context.Background())
				wardHeartbeat = startSchedulerFn(wardCtx, timeout/10)
			}
			startWard()
			pulse := time.NewTicker(pulseInterval)

		MonitorLoop:
			for {
				timeoutSignal := time.After(timeout)

				for {
					select {
					case <-pulse.C:
						sendPulse(heartbeat)
					case <-wardHeartbeat:
						continue MonitorLoop
					case <-timeoutSignal:
						srv.logger.Println("steward: dispatch ward unhealthy - restaring...")
						cancel()
						startWard()
						continue MonitorLoop
					case <-ctx.Done():
						pulse.Stop()
						return
					}
				}
			}
		}()
		return heartbeat
	}
}

// Dispatch listens to the job queue for messages, consumes them and
// dispatches the work items to the worker pool for execution.
func (srv *schedulerservice) Dispatch(duration time.Duration) port.StartGoroutineFn {
	return func(ctx context.Context, pulseInterval time.Duration) <-chan struct{} {
		heartbeat := make(chan struct{})

		go func() {
			defer close(heartbeat)
			ticker := time.NewTicker(duration)
			// Hearbeat set as half of the dispatching interval.
			pulse := time.NewTicker(pulseInterval)

		DispatchLoop:
			for {
				select {
				case <-ctx.Done():
					ticker.Stop()
					pulse.Stop()
					srv.logger.Info("exiting...")
					return
				case <-ticker.C:
					j := srv.jobQueue.Pop()
					if j == nil {
						sendPulse(heartbeat)
						continue
					}
					w := srv.workService.CreateWork(j)
				HeartbeatLoop:
					for {
						select {
						case srv.workService.Dispatch() <- w:
							message := fmt.Sprintf("job with ID: %s", j.ID)
							if j.BelongsToPipeline() {
								message = fmt.Sprintf("pipeline with ID: %s", j.PipelineID)
							}
							srv.logger.Infof("sent work for %s to worker pool", message)
							continue DispatchLoop
						case <-pulse.C:
							sendPulse(heartbeat)
							continue HeartbeatLoop
						case <-ctx.Done():
							continue DispatchLoop
						}
					}
				case <-pulse.C:
					sendPulse(heartbeat)
				}
			}
		}()
		return heartbeat
	}
}

// Schedule polls the storage in given interval and schedules due jobs for execution.
func (srv *schedulerservice) Schedule(duration time.Duration) port.StartGoroutineFn {
	return func(ctx context.Context, pulseInterval time.Duration) <-chan struct{} {
		heartbeat := make(chan struct{})

		go func() {
			defer close(heartbeat)
			ticker := time.NewTicker(duration)
			// Hearbeat set as half of the dispatching interval.
			pulse := time.NewTicker(pulseInterval)

		ScheduleLoop:
			for {
				select {
				case <-ctx.Done():
					ticker.Stop()
					pulse.Stop()
					srv.logger.Info("exiting...")
					return
				case <-ticker.C:
					dueJobs, err := srv.storage.GetDueJobs()
					if err != nil {
						srv.logger.Errorf("could not get due jobs from storage: %s", err)
						continue
					}
				DueJobsLoop:
					for _, j := range dueJobs {
						if j.BelongsToPipeline() {
							for job := j; job.HasNext(); job = job.Next {
								job.Next, err = srv.storage.GetJob(job.NextJobID)
								if err != nil {
									srv.logger.Errorf("could not get piped due job from storage: %s", err)
									continue DueJobsLoop
								}
							}
						}
						w := srv.workService.CreateWork(j)
					HeartbeatLoop:
						for {
							select {
							case srv.workService.Dispatch() <- w:
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
								continue DueJobsLoop
							case <-pulse.C:
								sendPulse(heartbeat)
								continue HeartbeatLoop
							case <-ctx.Done():
								continue ScheduleLoop
							}
						}
					}
				case <-pulse.C:
					sendPulse(heartbeat)
				}
			}
		}()
		return heartbeat
	}
}

func sendPulse(heartbeat chan<- struct{}) {
	select {
	case heartbeat <- struct{}{}:
	default:
	}
}
