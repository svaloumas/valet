package domain

import (
	"fmt"
	"strings"
	"time"
	"valet/internal/core/domain/task"
)

// Job represents an async task.
type Job struct {
	// ID is the auto-generated job identifier in UUID4 format.
	ID string `json:"id"`

	// Name is the name of the job.
	Name string `json:"name"`

	// TaskName is the name of the task to be executed.
	TaskName string `json:"task_name"`

	// Timeout is the time in seconds after which the job task will be interrupted.
	Timeout int `json:"timeout,omitempty"`

	// Description gives some information about the job.
	Description string `json:"description,omitempty"`

	// Status represents the status of the job.
	Status JobStatus `json:"status"`

	// FailureReason holds the error message that led to the job failure, if any.
	FailureReason string `json:"failure_reason,omitempty"`

	// CreatedAt is the UTC timestamp of the job creation.
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// StartedAt is the UTC timestamp of the moment the job started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is the UTC timestamp of the moment the job finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Metadata is the payload provided for the specific job.
	Metadata interface{} `json:"metadata"`
}

// NewJob initializes and returns a new Job instance.
func NewJob(
	uuid, name, taskName, description string, timeout int,
	createdAt *time.Time, metadata interface{}) *Job {

	return &Job{
		ID:          uuid,
		Name:        name,
		TaskName:    taskName,
		Timeout:     timeout,
		Description: description,
		Metadata:    metadata,
		Status:      Pending,
		CreatedAt:   createdAt,
	}
}

// MarkStarted updates the status and timestamp at the moment the job started.
func (j *Job) MarkStarted(startedAt *time.Time) {
	j.Status = InProgress
	j.StartedAt = startedAt
}

// MarkCompleted updates the status and timestamp at the moment the job finished.
func (j *Job) MarkCompleted(completedAt *time.Time) {

	j.Status = Completed
	j.CompletedAt = completedAt
}

// MarkFailed updates the status and timestamp at the moment the job failed.
func (j *Job) MarkFailed(failedAt *time.Time, reason string) {

	j.Status = Failed
	j.FailureReason = reason
	j.CompletedAt = failedAt
}

// Validate perfoms basic sanity checks on the job request payload.
func (job *Job) Validate(taskrepo *task.TaskRepository) error {
	var required []string

	if job.Name == "" {
		required = append(required, "name")
	}

	if job.TaskName == "" {
		required = append(required, "task_name")
	}

	if len(required) > 0 {
		return fmt.Errorf(strings.Join(required, ", ") + " required")
	}

	_, err := taskrepo.GetTaskFunc(job.TaskName)
	if err != nil {
		taskNames := taskrepo.GetTaskNames()
		return fmt.Errorf("%s is not a valid task name - valid tasks: %v", job.TaskName, taskNames)
	}

	if job.Status != Undefined {
		err := job.Status.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}
