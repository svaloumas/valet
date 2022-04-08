package domain

import (
	"errors"
	"testing"
	"time"

	"valet/internal/core/domain/taskrepo"
)

var validTasks = map[string]taskrepo.TaskFunc{
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

func TestJobMarkScheduled(t *testing.T) {
	j := new(Job)
	createdAt := time.Now()
	j.CreatedAt = &createdAt

	scheduledAt := createdAt.Add(10 * time.Second)
	j.ScheduledAt = &scheduledAt

	status := Scheduled

	j.MarkScheduled(&scheduledAt)

	if j.Status != status {
		t.Errorf("expected job status %s, got %s instead", j.Status, status)
	}
	if *j.ScheduledAt != scheduledAt {
		t.Errorf("expected job scheduled_at %s, got %s instead", j.ScheduledAt, scheduledAt)
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
	taskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)

	tests := []struct {
		name string
		job  *Job
		desc string
	}{
		{"empty payload", &Job{}, "name, task_name required"},
		{"only name given", &Job{Name: "a name"}, "task_name required"},
		{"only task_name given", &Job{TaskName: "test_task"}, "name required"},
		{
			"wrong stauts",
			&Job{Name: "a name", TaskName: "test_task", Description: "some_description", Status: 7},
			"7 is not a valid job status, valid statuses: map[PENDING:1 SCHEDULED:2 IN_PROGRESS:3 COMPLETED:4 FAILED:5]",
		},
		{
			"wrong task type",
			&Job{Name: "a name", TaskName: "wrongtask", Description: "some_description", Status: 2},
			"wrongtask is not a valid task name - valid tasks: [test_task]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate(taskrepo)
			if err != nil && err.Error() != tt.desc {
				t.Errorf("validator returned wrong error: got %v want %v", err.Error(), tt.desc)
			}
		})
	}
}
