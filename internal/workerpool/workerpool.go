package wp

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
)

var _ WorkerPool = &WorkerPoolImpl{}

// WorkerPoolImpl is a concrete implementation of WorkerPool.
type WorkerPoolImpl struct {
	// The fixed amount of goroutines that will be handling running jobs.
	concurrency int
	// The maximum capacity of the worker pool queue. If exceeded, sending new
	// tasks to the pool will return an error.
	backlog int

	jobService    port.JobService
	resultService port.ResultService
	queue         chan domain.JobItem
	wg            sync.WaitGroup
	Log           *log.Logger
}

// NewWorkerPoolImpl initializes and returns a new worker pool.
func NewWorkerPoolImpl(
	jobService port.JobService,
	resultService port.ResultService,
	concurrency, backlog int) *WorkerPoolImpl {

	logger := log.New(os.Stderr, "[worker-pool] ", log.LstdFlags)
	return &WorkerPoolImpl{
		concurrency:   concurrency,
		backlog:       backlog,
		jobService:    jobService,
		resultService: resultService,
		queue:         make(chan domain.JobItem, backlog),
		Log:           logger,
	}
}

// Start starts the worker pool.
func (wp *WorkerPoolImpl) Start() {
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.schedule(i, wp.queue, &wp.wg)
	}
	wp.Log.Printf("set up %d workers with a queue of backlog %d", wp.concurrency, wp.backlog)
}

// Send schedules the job. An error is returned if the job backlog is full.
func (wp *WorkerPoolImpl) Send(jobItem domain.JobItem) error {
	select {
	case wp.queue <- jobItem:
		go func() {
			if err := wp.resultService.Create(domain.FutureJobResult{Result: jobItem.Result}); err != nil {
				wp.Log.Printf("could not create job result to the repository")
			}
		}()
		return nil
	default:
		return &apperrors.FullWorkerPoolBacklog{}
	}
}

func (wp *WorkerPoolImpl) CreateJobItem(j *domain.Job) domain.JobItem {
	// TODO: Consider making timeout unit configurable.
	return domain.JobItem{
		Job:         j,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Second,
	}
}

// Stop signals the workers to stop working gracefully.
func (wp *WorkerPoolImpl) Stop() {
	close(wp.queue)
	wp.Log.Println("waiting for ongoing tasks to finish...")
	wp.wg.Wait()
}

func (wp *WorkerPoolImpl) schedule(id int, queue <-chan domain.JobItem, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", id)
	for item := range queue {
		wp.Log.Printf("%s executing task...", logPrefix)
		if err := wp.jobService.Exec(context.Background(), item); err != nil {
			wp.Log.Printf("could not update job status: %s", err)
		}
		wp.Log.Printf("%s task finished!", logPrefix)
	}
	wp.Log.Printf("%s exiting...", logPrefix)
}
