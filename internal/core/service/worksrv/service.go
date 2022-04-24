package worksrv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/internal/core/service/worksrv/work"
	rtime "github.com/svaloumas/valet/pkg/time"
)

const (
	WorkTypeTask                    = "task"
	WorkTypePipeline                = "pipeline"
	DefaultJobTimeout time.Duration = 84600 * time.Second
)

var _ port.WorkService = &workservice{}

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
		go srv.startWorker(i, srv.queue, &srv.wg)
	}
	srv.logger.Infof("set up %d workers with a queue of backlog %d", srv.concurrency, srv.backlog)
}

// Send sends a work to the worker pool.
func (srv *workservice) Send(w work.Work) {
	workType := WorkTypeTask
	for job := w.Job; ; job, workType = job.Next, WorkTypePipeline {
		go func() {
			futureResult := domain.FutureJobResult{Result: w.Result}
			result, ok := futureResult.Wait()

			if ok {
				if err := srv.storage.CreateJobResult(&result); err != nil {
					srv.logger.Errorf("could not create job result to the repository %s", err)
				}
			}
		}()

		if !job.HasNext() {
			break
		}
	}
	w.Type = workType

	srv.queue <- w
}

// CreateWork creates and return a new Work instance.
func (srv *workservice) CreateWork(j *domain.Job) work.Work {
	resultChan := make(chan domain.JobResult, 1)
	return work.NewWork(j, resultChan, srv.timeoutUnit)
}

// Stop signals the workers to stop working gracefully.
func (srv *workservice) Stop() {
	close(srv.queue)
	srv.logger.Info("waiting for ongoing tasks to finish...")
	srv.wg.Wait()
}

// ExecJobWork executes the job work.
func (srv *workservice) ExecJobWork(ctx context.Context, w work.Work) error {
	// Do not let the go-routines wait for result in case of early exit.
	defer close(w.Result)

	startedAt := srv.time.Now()
	w.Job.MarkStarted(&startedAt)
	if err := srv.storage.UpdateJob(w.Job.ID, w.Job); err != nil {
		return err
	}
	timeout := DefaultJobTimeout
	if w.Job.Timeout > 0 {
		timeout = time.Duration(w.Job.Timeout) * w.TimeoutUnit
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	jobResultChan := make(chan domain.JobResult, 1)

	srv.work(w.Job, jobResultChan, nil)

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

	return nil
}

// ExecPipelineWork executes the pipeline work.
func (srv *workservice) ExecPipelineWork(ctx context.Context, w work.Work) error {
	// Do not let the go-routines wait for result in case of early exit.
	defer close(w.Result)

	p, err := srv.storage.GetPipeline(w.Job.PipelineID)
	if err != nil {
		return err
	}
	var jobResult domain.JobResult

	for job, i := w.Job, 0; ; job, i = job.Next, i+1 {
		startedAt := srv.time.Now()
		job.MarkStarted(&startedAt)
		if err := srv.storage.UpdateJob(job.ID, job); err != nil {
			return err
		}
		if i == 0 {
			p.MarkStarted(&startedAt)
			if err := srv.storage.UpdatePipeline(p.ID, p); err != nil {
				return err
			}
		}

		timeout := DefaultJobTimeout
		if job.Timeout > 0 {
			timeout = time.Duration(job.Timeout) * w.TimeoutUnit
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		jobResultChan := make(chan domain.JobResult, 1)

		srv.work(job, jobResultChan, jobResult.Metadata)

		select {
		case <-ctx.Done():
			failedAt := srv.time.Now()
			job.MarkFailed(&failedAt, ctx.Err().Error())
			p.MarkFailed(&failedAt)

			jobResult = domain.JobResult{
				JobID:    job.ID,
				Metadata: nil,
				Error:    ctx.Err().Error(),
			}
		case jobResult = <-jobResultChan:
			if jobResult.Error != "" {
				failedAt := srv.time.Now()
				job.MarkFailed(&failedAt, jobResult.Error)
				p.MarkFailed(&failedAt)
			} else {
				completedAt := srv.time.Now()
				job.MarkCompleted(&completedAt)

				if !job.HasNext() {
					p.MarkCompleted(&completedAt)
				}
			}
		}
		// Reset timeout.
		cancel()

		if err := srv.storage.UpdateJob(job.ID, job); err != nil {
			return err
		}
		if p.Status == domain.Failed || p.Status == domain.Completed {
			if err := srv.storage.UpdatePipeline(p.ID, p); err != nil {
				return err
			}
		}
		w.Result <- jobResult

		// Stop the pipeline execution on failure.
		if job.Status == domain.Failed {
			break
		}

		// Stop the pipeline execution if there's no other job.
		if !job.HasNext() {
			break
		}
	}

	return nil
}

func (srv *workservice) work(
	job *domain.Job,
	jobResultChan chan domain.JobResult,
	previousJobResultsMetadata interface{}) {

	go func() {
		defer func() {
			if p := recover(); p != nil {
				result := domain.JobResult{
					JobID:    job.ID,
					Metadata: nil,
					Error:    fmt.Errorf("%v", p).Error(),
				}
				jobResultChan <- result
				close(jobResultChan)
			}
		}()
		var errMsg string

		// Should be already validated.
		taskFunc, _ := srv.taskrepo.GetTaskFunc(job.TaskName)

		params := []interface{}{
			job.TaskParams,
		}
		if job.UsePreviousResults && previousJobResultsMetadata != nil {
			params = append(params, previousJobResultsMetadata)
		}
		// Perform the actual work.
		resultMetadata, jobErr := taskFunc(params...)
		if jobErr != nil {
			errMsg = jobErr.Error()
		}

		result := domain.JobResult{
			JobID:    job.ID,
			Metadata: resultMetadata,
			Error:    errMsg,
		}
		jobResultChan <- result
		close(jobResultChan)
	}()
}

func (srv *workservice) Exec(ctx context.Context, w work.Work) error {
	if w.Type == WorkTypePipeline {
		return srv.ExecPipelineWork(ctx, w)
	}
	return srv.ExecJobWork(ctx, w)
}

func (srv *workservice) startWorker(id int, queue <-chan work.Work, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", id)
	for work := range queue {
		srv.logger.Infof("%s executing %s...", logPrefix, work.Type)
		if err := srv.Exec(context.Background(), work); err != nil {
			srv.logger.Errorf("could not update job status: %s", err)
		}
		srv.logger.Infof("%s %s finished!", logPrefix, work.Type)
	}
	srv.logger.Infof("%s exiting...", logPrefix)
}
