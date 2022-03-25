package workerpool

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/internal/repository/workerpool/task"
	"valet/pkg/apperrors"
)

var _ port.WorkerPool = &WorkerPoolImpl{}

// WorkerPoolImpl is a concrete implementation of WorkerPool.
type WorkerPoolImpl struct {
	// The task that should be run.
	task        task.TaskFunc
	taskTimeout time.Duration
	// The fixed amount of goroutines that will be handling running jobs.
	concurrency int
	// The maximum capacity of the worker pool queue. If exceeded, sending new
	// tasks to the pool will return an error.
	backlog int

	jobService    port.JobService
	resultService port.ResultService
	queue         chan domain.JobItem
	wg            sync.WaitGroup
	logger        *log.Logger
}

// NewWorkerPoolImpl initializes and returns a new worker pool.
func NewWorkerPoolImpl(
	jobService port.JobService,
	resultService port.ResultService,
	concurrency, backlog int,
	taskTimeout time.Duration,
	task task.TaskFunc) *WorkerPoolImpl {

	logger := log.New(os.Stderr, "[worker-pool] ", log.LstdFlags)
	return &WorkerPoolImpl{
		task:          task,
		taskTimeout:   taskTimeout,
		concurrency:   concurrency,
		backlog:       backlog,
		jobService:    jobService,
		resultService: resultService,
		queue:         make(chan domain.JobItem, backlog),
		logger:        logger,
	}
}

// Start starts the worker pool.
func (wp *WorkerPoolImpl) Start() {
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.schedule(i, wp.queue, &wp.wg)
	}
	wp.logger.Printf("set up %d workers with a queue of backlog %d", wp.concurrency, wp.backlog)
}

// Send schedules the job. An error is returned if the job backlog is full.
func (wp *WorkerPoolImpl) Send(j *domain.Job) error {
	result := make(chan domain.JobResult, 1)
	wi := domain.JobItem{
		Job:    j,
		Result: result,
	}

	select {
	case wp.queue <- wi:
		go func() {
			if err := wp.resultService.Create(domain.FutureJobResult{Result: wi.Result}); err != nil {
				wp.logger.Printf("could not create job result to the repository")
			}
		}()
		return nil
	default:
		return &apperrors.FullWorkerPoolBacklog{}
	}
}

// Stop signals the workers to stop working gracefully.
func (wp *WorkerPoolImpl) Stop() {
	close(wp.queue)
	wp.logger.Println("waiting for ongoing tasks to finish...")
	wp.wg.Wait()
}

func (wp *WorkerPoolImpl) schedule(id int, queue <-chan domain.JobItem, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", id)
	for item := range queue {
		wp.logger.Printf("%s executing task...", logPrefix)
		if err := wp.jobService.ExecWithTimeout(wp.taskTimeout, context.Background(), item, wp.task); err != nil {
			wp.logger.Printf("task execution error: %s", err)
		}
		wp.logger.Printf("%s task finished!", logPrefix)
	}
	wp.logger.Printf("%s exiting...", logPrefix)
}
