package work

import (
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
)

// Work is the task to be executed by the workers.
type Work struct {
	Type        string
	Job         *domain.Job
	Result      chan domain.JobResult
	TimeoutUnit time.Duration
}

// NewWork initializes and returns a new Work instance.
func NewWork(
	j *domain.Job,
	resultChan chan domain.JobResult,
	timeoutUnit time.Duration) Work {

	return Work{
		Job:         j,
		Result:      resultChan,
		TimeoutUnit: timeoutUnit,
	}
}
