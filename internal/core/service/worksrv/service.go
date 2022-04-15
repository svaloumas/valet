package worksrv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/port"
	"valet/internal/core/service/worksrv/work"
	rtime "valet/pkg/time"
)

var _ port.WorkService = &workservice{}
var defaultJobTimeout time.Duration = 84600 * time.Second

type workservice struct {
	// The fixed amount of goroutines that will be handling running jobs.
	concurrency int
	// The maximum capacity of the worker pool queue. If exceeded, sending new
	// tasks to the pool will return an error.
	backlog int
	// The time unit for the calculation of the timeout interval for each task.
	timeoutUnit time.Duration

	storage  port.Storage
	taskrepo *taskrepo.TaskRepository
	time     rtime.Time
	queue    chan work.Work
	wg       sync.WaitGroup
	logger   *logrus.Logger
}

// New creates a new work service.
func New(
	storage port.Storage,
	taskrepo *taskrepo.TaskRepository,
	time rtime.Time, timeoutUnit time.Duration,
	concurrency, backlog int, logger *logrus.Logger) *workservice {

	return &workservice{
		storage:     storage,
		taskrepo:    taskrepo,
		concurrency: concurrency,
		backlog:     backlog,
		timeoutUnit: timeoutUnit,
		queue:       make(chan work.Work, backlog),
		time:        time,
		logger:      logger,
	}
}

// Start starts the worker pool.
func (srv *workservice) Start() {
	for i := 0; i < srv.concurrency; i++ {
		srv.wg.Add(1)
		go srv.work(i, srv.queue, &srv.wg)
	}
	srv.logger.Infof("set up %d workers with a queue of backlog %d", srv.concurrency, srv.backlog)
}

// Send sends a work to the worker pool.
func (srv *workservice) Send(w work.Work) {
	srv.queue <- w
	go func() {
		futureResult := domain.FutureJobResult{Result: w.Result}
		result := futureResult.Wait()

		if err := srv.storage.CreateJobResult(&result); err != nil {
			srv.logger.Errorf("could not create job result to the repository %s", err)
		}
	}()
}

// CreateWork creates and return a new Work instance.
func (srv *workservice) CreateWork(j *domain.Job) work.Work {
	// Should be already validated.
	taskFunc, _ := srv.taskrepo.GetTaskFunc(j.TaskName)
	resultChan := make(chan domain.JobResult, 1)
	return work.NewWork(j, resultChan, taskFunc, srv.timeoutUnit)
}

// Stop signals the workers to stop working gracefully.
func (srv *workservice) Stop() {
	close(srv.queue)
	srv.logger.Info("waiting for ongoing tasks to finish...")
	srv.wg.Wait()
}

// Exec executes the work.
func (srv *workservice) Exec(ctx context.Context, w work.Work) error {
	startedAt := srv.time.Now()
	w.Job.MarkStarted(&startedAt)
	if err := srv.storage.UpdateJob(w.Job.ID, w.Job); err != nil {
		return err
	}
	timeout := defaultJobTimeout
	if w.Job.Timeout > 0 {
		timeout = time.Duration(w.Job.Timeout) * w.TimeoutUnit
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	jobResultChan := make(chan domain.JobResult, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				result := domain.JobResult{
					JobID:    w.Job.ID,
					Metadata: nil,
					Error:    fmt.Errorf("%v", p).Error(),
				}
				jobResultChan <- result
			}
		}()
		var errMsg string

		// Perform the actual work.
		resultMetadata, jobErr := w.TaskFunc(w.Job.TaskParams)
		if jobErr != nil {
			errMsg = jobErr.Error()
		}

		result := domain.JobResult{
			JobID:    w.Job.ID,
			Metadata: resultMetadata,
			Error:    errMsg,
		}
		jobResultChan <- result
		close(jobResultChan)
	}()

	var jobResult domain.JobResult
	select {
	case <-ctx.Done():
		failedAt := srv.time.Now()
		w.Job.MarkFailed(&failedAt, ctx.Err().Error())

		jobResult = domain.JobResult{
			JobID:    w.Job.ID,
			Metadata: nil,
			Error:    ctx.Err().Error(),
		}
	case jobResult = <-jobResultChan:
		if jobResult.Error != "" {
			failedAt := srv.time.Now()
			w.Job.MarkFailed(&failedAt, jobResult.Error)
		} else {
			completedAt := srv.time.Now()
			w.Job.MarkCompleted(&completedAt)
		}
	}
	if err := srv.storage.UpdateJob(w.Job.ID, w.Job); err != nil {
		return err
	}
	w.Result <- jobResult
	close(w.Result)

	return nil
}

func (srv *workservice) work(id int, queue <-chan work.Work, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", id)
	for work := range queue {
		srv.logger.Infof("%s executing task...", logPrefix)
		if err := srv.Exec(context.Background(), work); err != nil {
			srv.logger.Errorf("could not update job status: %s", err)
		}
		srv.logger.Infof("%s task finished!", logPrefix)
	}
	srv.logger.Infof("%s exiting...", logPrefix)
}
