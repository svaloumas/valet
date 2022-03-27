package domain

import (
	"testing"
	"time"
)

func TestJobMarkStarted(t *testing.T) {
	j := new(Job)
	createdAt := time.Now()
	j.CreatedAt = &createdAt

	status := InProgress
	startedAt := time.Now()
	j.MarkStarted(&startedAt)

	if j.Status != status {
		t.Errorf("expected job status %s, got %s instead", j.Status, status)
	}
	if *j.StartedAt != startedAt {
		t.Errorf("expected job started_at %s, got %s instead", j.StartedAt, startedAt)
	}
}

func TestJobMarkCompleted(t *testing.T) {
	j := new(Job)
	createdAt := time.Now()
	j.CreatedAt = &createdAt

	startedAt := time.Now()
	j.StartedAt = &startedAt

	status := Completed
	completedAt := j.StartedAt.Add(10 * time.Second)

	j.MarkCompleted(&completedAt)

	if j.Status != status {
		t.Errorf("expected job status %s, got %s instead", j.Status, status)
	}
	if *j.CompletedAt != completedAt {
		t.Errorf("expected job completed_at %s, got %s instead", j.CompletedAt, completedAt)
	}
}

func TestJobMarkFailed(t *testing.T) {
	j := new(Job)
	createdAt := time.Now()
	j.CreatedAt = &createdAt

	startedAt := time.Now()
	j.StartedAt = &startedAt

	status := Failed
	reason := "some_error"

	failedAt := j.StartedAt.Add(10 * time.Second)
	j.MarkFailed(&failedAt, reason)

	if j.Status != status {
		t.Errorf("expected job status %s, got %s instead", j.Status, status)
	}
	if j.FailureReason != reason {
		t.Errorf("expected job reason %s, got %s instead", j.FailureReason, reason)
	}
	if *j.CompletedAt != failedAt {
		t.Errorf("expected job completed_at %s, got %s instead", j.CompletedAt, failedAt)
	}
}

func TestJobValidate(t *testing.T) {
	tests := []struct {
		job  *Job
		desc string
	}{
		{&Job{}, "name, task_type required"},
		{&Job{Name: "a name"}, "task_type required"},
		{&Job{TaskType: "dummytask"}, "name required"},
		{
			&Job{Name: "a name", TaskType: "dummytask", Description: "some_description", Status: 7},
			"7 is not a valid job status, valid statuses: map[PENDING:1 IN_PROGRESS:2 COMPLETED:3 FAILED:4]",
		},
		{
			&Job{Name: "a name", TaskType: "wrongtask", Description: "some_description", Status: 2},
			"wrongtask is not a valid task type - valid task types: [dummytask]",
		},
	}

	for _, tt := range tests {
		err := tt.job.Validate()
		if err != nil && err.Error() != tt.desc {
			t.Errorf("validator returned wrong error: got %v want %v", err.Error(), tt.desc)
		}
	}
}
