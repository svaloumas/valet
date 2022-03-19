package domain

import (
	"errors"
	"strings"
	"time"
)

// Job represents an async task.
type Job struct {
	// ID is the auto-generated job identifier in UUID4 format.
	ID string

	// Name is the name of the job.
	Name string

	// Description gives some information about the job.
	Description string

	// Status represents the status of the job.
	Status JobStatus

	// FailureReason holds the error message that led to the job failure, if any.
	FailureReason string

	// CreatedAt is the UTC timestamp of the job creation.
	CreatedAt *time.Time

	// StartedAt is the UTC timestamp of the moment the job started.
	StartedAt *time.Time

	// CompletedAt is the UTC timestamp of the moment the job finished.
	CompletedAt *time.Time
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
func (job *Job) Validate() error {
	var required []string

	if job.Name == "" {
		required = append(required, "name")
	}

	if len(required) > 0 {
		return errors.New(strings.Join(required, ", ") + " required")
	}

	if job.Status != Undefined {
		err := job.Status.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}
