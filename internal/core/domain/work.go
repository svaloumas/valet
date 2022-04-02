package domain

import (
	"time"

	"valet/internal/core/domain/taskrepo"
)

type Work struct {
	Job         *Job
	Result      chan JobResult
	TaskFunc    taskrepo.TaskFunc
	TimeoutUnit time.Duration
}

// NewWork initializes and returns a new Work instance.
func NewWork(
	j *Job,
	resultChan chan JobResult,
	taskFunc taskrepo.TaskFunc,
	timeoutUnit time.Duration) Work {

	return Work{
		Job:         j,
		Result:      resultChan,
		TaskFunc:    taskFunc,
		TimeoutUnit: timeoutUnit,
	}
}
