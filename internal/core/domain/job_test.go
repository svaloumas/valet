package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/svaloumas/valet/internal/core/service/tasksrv"
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

func TestJobSetDuration(t *testing.T) {
	j := &Job{
		Name: "job_name",
	}

	startedAt := time.Now()
	j.StartedAt = &startedAt

	completedAt := startedAt.Add(10 * time.Second)
	j.CompletedAt = &completedAt

	duration := j.CompletedAt.Sub(*j.StartedAt) / time.Millisecond

	j.Status = Failed
	j.SetDuration()

	if *j.Duration != duration {
		t.Errorf("expected failed job duration %s, got %s instead", j.Duration, duration)
	}

	j.Status = Completed
	j.SetDuration()
	if *j.Duration != duration {
		t.Errorf("expected completed job duration %s, got %s instead", j.Duration, duration)
	}
}

func TestJobValidate(t *testing.T) {
	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskService := tasksrv.New()
	taskService.Register("test_task", taskFunc)

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
			taskrepo := taskService.GetTaskRepository()

			err := tt.job.Validate(taskrepo)
			if err != nil && err.Error() != tt.desc {
				t.Errorf("validator returned wrong error: got %v want %v", err.Error(), tt.desc)
			}
		})
	}
}

func TestJobIsScheduled(t *testing.T) {
	runAt := time.Now()
	scheduledJob := &Job{RunAt: &runAt}
	notScheduledJob := &Job{}

	if !scheduledJob.IsScheduled() {
		t.Errorf("IsScheduled returned wrong value: got %t, want true", scheduledJob.IsScheduled())
	}
	if notScheduledJob.IsScheduled() {
		t.Errorf("IsScheduled returned wrong value: got %t, want false", notScheduledJob.IsScheduled())
	}
}

func TestHasNext(t *testing.T) {
	job := Job{}
	pipedJob := Job{NextJobID: "some_id"}

	if job.HasNext() {
		t.Errorf("HasNext returned wrong value: got %t, want false", job.HasNext())
	}
	if !pipedJob.HasNext() {
		t.Errorf("HasNext returned wrong value: got %t, want true", pipedJob.HasNext())
	}
}

func TestBelongsToPipeline(t *testing.T) {
	job := Job{}
	pipelineJob := Job{PipelineID: "some_id"}

	if job.BelongsToPipeline() {
		t.Errorf("BelongsToPipeline returned wrong value: got %t, want false", job.BelongsToPipeline())
	}
	if !pipelineJob.BelongsToPipeline() {
		t.Errorf("BelongsToPipeline returned wrong value: got %t, want true", pipelineJob.BelongsToPipeline())
	}
}

func TestNewJob(t *testing.T) {
	testTime := time.Now()

	expected := &Job{
		ID:       "job_id",
		Name:     "job_name",
		TaskName: "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description:        "some description",
		PipelineID:         "pipeline_id",
		Status:             Pending,
		NextJobID:          "next_job_id",
		Timeout:            10,
		RunAt:              &testTime,
		CreatedAt:          &testTime,
		UsePreviousResults: true,
	}
	job := NewJob(
		expected.ID, expected.Name, expected.TaskName, expected.Description, expected.PipelineID,
		expected.NextJobID, expected.Timeout, expected.RunAt, expected.CreatedAt, expected.UsePreviousResults, expected.TaskParams)

	if eq := reflect.DeepEqual(job, expected); !eq {
		t.Errorf("new job returned wrong job: got %v want %v", job, expected)
	}
}

func TestNewJobRunAtIsZero(t *testing.T) {
	testTime := time.Now()
	runAt := time.Time{}

	expected := &Job{
		ID:       "job_id",
		Name:     "job_name",
		TaskName: "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description:        "some description",
		PipelineID:         "pipeline_id",
		Status:             Pending,
		NextJobID:          "next_job_id",
		Timeout:            10,
		RunAt:              &runAt,
		CreatedAt:          &testTime,
		UsePreviousResults: true,
	}
	job := NewJob(
		expected.ID, expected.Name, expected.TaskName, expected.Description, expected.PipelineID,
		expected.NextJobID, expected.Timeout, expected.RunAt, expected.CreatedAt, expected.UsePreviousResults, expected.TaskParams)

	if job.RunAt != nil {
		t.Errorf("new job with zero run_at time returned wrong run_at value: got %v want nil", job.RunAt)
	}
}
