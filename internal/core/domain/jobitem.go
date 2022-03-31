package domain

import (
	"time"
)

type JobItem struct {
	Job         *Job
	Result      chan JobResult
	TimeoutUnit time.Duration
}

// NewJobItem initializes and returns a new JobItem instance.
func NewJobItem(
	j *Job,
	resultChan chan JobResult,
	timeoutUnit time.Duration) JobItem {

	return JobItem{
		Job:         j,
		Result:      resultChan,
		TimeoutUnit: timeoutUnit,
	}
}
