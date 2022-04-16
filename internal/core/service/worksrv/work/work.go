package work

import (
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/domain/taskrepo"
)

// Work is the task to be executed by the workers.
type Work struct {
	Job         *domain.Job
	Result      chan domain.JobResult
	TaskFunc    taskrepo.TaskFunc
	TimeoutUnit time.Duration
}

// NewWork initializes and returns a new Work instance.
func NewWork(
	j *domain.Job,
	resultChan chan domain.JobResult,
	taskFunc taskrepo.TaskFunc,
	timeoutUnit time.Duration) Work {

	return Work{
		Job:         j,
		Result:      resultChan,
		TaskFunc:    taskFunc,
		TimeoutUnit: timeoutUnit,
	}
}
