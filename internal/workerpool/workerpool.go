package workerpool

import (
	"fmt"
	"log"
	"os"
	"sync"

	"valet/internal/core/domain"
	"valet/internal/core/port"
)

// WorkResult contains the result of a job.
type WorkResult struct {
	Metadata []byte
	Error    error
}

// FutureWorkResult is a WorkResult that may not yet
// have become available and can be Wait()'ed on.
type FutureWorkResult struct {
	resultQueue <-chan WorkResult
}

// Wait waits for WorkResult to become available and returns it.
func (f FutureWorkResult) Wait() WorkResult {
	r, ok := <-f.resultQueue
	if !ok {
		// This should never happen, reading from the result
		// channel is exclusive to this future
		panic("failed to read from result channel")
	}
	return r
}

type workItem struct {
	job         *domain.Job
	resultQueue chan<- WorkResult
}

// WorkerPoolImpl is a concrete implementation of WorkerPool.
type WorkerPoolImpl struct {
	// The type of task that should be run.
	task port.Task
	// The fixed amount of goroutines that will be handling running jobs.
	concurrency int
	// The maximum capacity of the worker pool queue. If exceeded, sending new
	// tasks to the pool will return an error.
	backlog int

	queue  chan workItem
	wg     sync.WaitGroup
	logger *log.Logger
}

// NewWorkerPoolImpl initializes and returns a new worker pool.
func NewWorkerPoolImpl(concurrency, backlog int, task port.Task) *WorkerPoolImpl {
	logger := log.New(os.Stderr, "[worker-pool] ", log.LstdFlags)
	return &WorkerPoolImpl{
		task:        task,
		concurrency: concurrency,
		backlog:     backlog,
		queue:       make(chan workItem, backlog),
		logger:      logger,
	}
}

func (wp *WorkerPoolImpl) Start() {
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.schedule(i, wp.queue, &wp.wg)
	}
	wp.logger.Printf("set up %d workers with a queue of backlog %d", wp.concurrency, wp.backlog)
}

// Send schedules the job. An error is returned if the job backlog is full.
func (wp *WorkerPoolImpl) Send(j *domain.Job) error {
	resultQueue := make(chan WorkResult, 1)
	wi := workItem{
		job:         j,
		resultQueue: resultQueue,
	}

	select {
	case wp.queue <- wi:
		return nil
	default:
		return &FullWorkerPoolBacklog{}
	}
}

// Stop signals the workers to stop working gracefully.
func (wp *WorkerPoolImpl) Stop() {
	close(wp.queue)
	wp.logger.Println("waiting for ongoing tasks to finish...")
	wp.wg.Wait()
	wp.logger.Println("exiting...")
}

func (wp *WorkerPoolImpl) schedule(id int, queue <-chan workItem, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", id)
	for item := range queue {
		wp.logger.Printf("%s executing work...", logPrefix)
		metadata, err := wp.task.Run(item.job.Metadata)

		select {
		case item.resultQueue <- WorkResult{Metadata: metadata, Error: err}:
			wp.logger.Printf("%s task finished!", logPrefix)
		default:
			// This should never happen as the result queue chan should be unique for this worker.
			wp.logger.Panicf("%s failed to write result to the result queue channel", logPrefix)
		}
		close(item.resultQueue)
	}
	wp.logger.Printf("%s exiting...", logPrefix)
}
