package domain

import (
	"errors"
	"testing"
	"time"

	"valet/internal/core/domain/task"
)

var validTasks = map[string]task.TaskFunc{
	"test_task": func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	},
}

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
		name string
		job  *Job
		desc string
	}{
		{"empty payload", &Job{}, "name, task_type required"},
		{"only name given", &Job{Name: "a name"}, "task_type required"},
		{"only task_type given", &Job{TaskType: "test_task"}, "name required"},
		{
			"wrong stauts",
			&Job{Name: "a name", TaskType: "test_task", Description: "some_description", Status: 7},
			"7 is not a valid job status, valid statuses: map[PENDING:1 IN_PROGRESS:2 COMPLETED:3 FAILED:4]",
		},
		{
			"wrong task type",
			&Job{Name: "a name", TaskType: "wrongtask", Description: "some_description", Status: 2},
			"wrongtask is not a valid task type - valid task types: [test_task]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate(validTasks)
			if err != nil && err.Error() != tt.desc {
				t.Errorf("validator returned wrong error: got %v want %v", err.Error(), tt.desc)
			}
		})
	}
}
