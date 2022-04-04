package worksrv

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/port"
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

	jobRepository    port.JobRepository
	resultRepository port.ResultRepository
	taskrepo         *taskrepo.TaskRepository
	time             rtime.Time
	queue            chan domain.Work
	wg               sync.WaitGroup
	Log              *log.Logger
}

// New creates a new work service.
func New(
	jobRepository port.JobRepository,
	resultRepository port.ResultRepository,
	taskrepo *taskrepo.TaskRepository,
	time rtime.Time, timeoutUnit time.Duration,
	concurrency int, backlog int) *workservice {

	logger := log.New(os.Stderr, "[worker-pool] ", log.LstdFlags)
	return &workservice{
		jobRepository:    jobRepository,
		resultRepository: resultRepository,
		taskrepo:         taskrepo,
		concurrency:      concurrency,
		backlog:          backlog,
		timeoutUnit:      timeoutUnit,
		queue:            make(chan domain.Work, backlog),
		time:             time,
		Log:              logger,
	}
}

// Start starts the worker pool.
func (srv *workservice) Start() {
	for i := 0; i < srv.concurrency; i++ {
		srv.wg.Add(1)
		go srv.work(i, srv.queue, &srv.wg)
	}
	srv.Log.Printf("set up %d workers with a queue of backlog %d", srv.concurrency, srv.backlog)
}

// Send sends a work to the worker pool.
func (srv *workservice) Send(w domain.Work) {
	srv.queue <- w
	go func() {
		futureResult := domain.FutureJobResult{Result: w.Result}
		result := futureResult.Wait()

		if err := srv.resultRepository.Create(&result); err != nil {
			srv.Log.Printf("could not create job result to the repository")
		}
	}()
}

// CreateWork creates and return a new Work instance.
func (srv *workservice) CreateWork(j *domain.Job) domain.Work {
	// Should be already validated.
	taskFunc, _ := srv.taskrepo.GetTaskFunc(j.TaskName)
	resultChan := make(chan domain.JobResult, 1)
	return domain.NewWork(j, resultChan, taskFunc, srv.timeoutUnit)
}

// Stop signals the workers to stop working gracefully.
func (srv *workservice) Stop() {
	close(srv.queue)
	srv.Log.Println("waiting for ongoing tasks to finish...")
	srv.wg.Wait()
}

// Exec executes the work.
func (srv *workservice) Exec(ctx context.Context, w domain.Work) error {
	startedAt := srv.time.Now()
	w.Job.MarkStarted(&startedAt)
	if err := srv.jobRepository.Update(w.Job.ID, w.Job); err != nil {
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
	if err := srv.jobRepository.Update(w.Job.ID, w.Job); err != nil {
		return err
	}
	w.Result <- jobResult
	close(w.Result)

	return nil
}

func (srv *workservice) work(id int, queue <-chan domain.Work, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", id)
	for work := range queue {
		srv.Log.Printf("%s executing task...", logPrefix)
		if err := srv.Exec(context.Background(), work); err != nil {
			srv.Log.Printf("could not update job status: %s", err)
		}
		srv.Log.Printf("%s task finished!", logPrefix)
	}
	srv.Log.Printf("%s exiting...", logPrefix)
}
