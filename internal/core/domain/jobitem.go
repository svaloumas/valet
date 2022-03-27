package domain

import (
	"valet/internal/core/domain/task"
)

type JobItem struct {
	Job      *Job
	Result   chan JobResult
	TaskFunc task.TaskFunc
}
