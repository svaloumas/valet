package domain

import (
	"time"
)

type Work struct {
	Job         *Job
	Result      chan JobResult
	TimeoutUnit time.Duration
}

// NewWork initializes and returns a new Work instance.
func NewWork(
	j *Job,
	resultChan chan JobResult,
	timeoutUnit time.Duration) Work {

	return Work{
		Job:         j,
		Result:      resultChan,
		TimeoutUnit: timeoutUnit,
	}
}
