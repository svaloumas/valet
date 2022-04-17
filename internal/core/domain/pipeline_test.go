package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/svaloumas/valet/internal/core/service/tasksrv"
)

func TestPipelineMarkStarted(t *testing.T) {
	p := new(Pipeline)
	createdAt := time.Now()
	p.CreatedAt = &createdAt

	status := InProgress
	startedAt := time.Now()
	p.MarkStarted(&startedAt)

	if p.Status != status {
		t.Errorf("expected pipeline status %s, got %s instead", p.Status, status)
	}
	if *p.StartedAt != startedAt {
		t.Errorf("expected pipeline started_at %s, got %s instead", p.StartedAt, startedAt)
	}
}

func TestPipelineMarkCompleted(t *testing.T) {
	p := new(Pipeline)
	createdAt := time.Now()
	p.CreatedAt = &createdAt

	startedAt := time.Now()
	p.StartedAt = &startedAt

	status := Completed
	completedAt := p.StartedAt.Add(10 * time.Second)

	p.MarkCompleted(&completedAt)

	if p.Status != status {
		t.Errorf("expected pipeline status %s, got %s instead", p.Status, status)
	}
	if *p.CompletedAt != completedAt {
		t.Errorf("expected pipeline completed_at %s, got %s instead", p.CompletedAt, completedAt)
	}
}

func TestPipelineMarkFailed(t *testing.T) {
	p := new(Pipeline)
	createdAt := time.Now()
	p.CreatedAt = &createdAt

	startedAt := time.Now()
	p.StartedAt = &startedAt

	status := Failed

	failedAt := p.StartedAt.Add(10 * time.Second)
	p.MarkFailed(&failedAt)

	if p.Status != status {
		t.Errorf("expected pipeline status %s, got %s instead", p.Status, status)
	}
	if *p.CompletedAt != failedAt {
		t.Errorf("expected pipeline completed_at %s, got %s instead", p.CompletedAt, failedAt)
	}
}

func TestPipelineSetDuration(t *testing.T) {
	p := &Pipeline{
		Name: "pipeline_name",
	}

	startedAt := time.Now()
	p.StartedAt = &startedAt

	completedAt := startedAt.Add(10 * time.Second)
	p.CompletedAt = &completedAt

	duration := p.CompletedAt.Sub(*p.StartedAt) / time.Millisecond

	p.Status = Failed
	p.SetDuration()

	if *p.Duration != duration {
		t.Errorf("expected failed pipeline duration %s, got %s instead", p.Duration, duration)
	}

	p.Status = Completed
	p.SetDuration()
	if *p.Duration != duration {
		t.Errorf("expected completed pipeline duration %s, got %s instead", p.Duration, duration)
	}
}

func TestPipelineValidate(t *testing.T) {
	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskService := tasksrv.New()
	taskService.Register("test_task", taskFunc)

	tests := []struct {
		name     string
		pipeline *Pipeline
		desc     string
	}{
		{"empty payload", &Pipeline{Jobs: []*Job{}}, "name, jobs required"},
		{"only name given", &Pipeline{Name: "pipeline_name", Jobs: []*Job{}}, "jobs required"},
		{"only jobs given", &Pipeline{Jobs: []*Job{{Name: "job_name", TaskName: "test_task"}, {Name: "another_job", TaskName: "test_task"}}}, "name required"},
		{"only one job given", &Pipeline{Name: "pipeline_name", Jobs: []*Job{{Name: "job_name", TaskName: "test_task"}}}, "pipeline shoud have at least 2 jobs, 1 given"},
		{
			"wrong stauts",
			&Pipeline{Name: "pipeline_name", Jobs: []*Job{{Name: "job_name", TaskName: "test_task"}, {Name: "another_job", TaskName: "test_task"}}, Status: 7},
			"7 is not a valid job status, valid statuses: map[PENDING:1 SCHEDULED:2 IN_PROGRESS:3 COMPLETED:4 FAILED:5]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pipeline.Validate()
			if err != nil && err.Error() != tt.desc {
				t.Errorf("validator returned wrong error: got %v want %v", err.Error(), tt.desc)
			}
		})
	}
}

func TestMakeNestedJobPipeline(t *testing.T) {
	p := &Pipeline{}

	jobs := []*Job{
		{Name: "first_job"},
		{Name: "second_job"},
		{Name: "third_job"},
	}

	p.Jobs = jobs

	p.CreateNestedJobPipeline()

	expected := jobs[0]
	expected.Next = jobs[1]
	expected.Next.Next = jobs[2]

	if eq := reflect.DeepEqual(jobs[0], expected); !eq {
		t.Errorf("CreateNestedJobPipeline returned wrong piped jobs: got %#v want %#v", jobs[0], expected)
	}
	if eq := reflect.DeepEqual(jobs[1], expected.Next); !eq {
		t.Errorf("CreateNestedJobPipeline returned wrong piped jobs: got %#v want %#v", jobs[1], expected.Next)
	}
	if eq := reflect.DeepEqual(jobs[2], expected.Next.Next); !eq {
		t.Errorf("CreateNestedJobPipeline returned wrong piped jobs: got %#v want %#v", jobs[2], expected.Next.Next)
	}
}

func TestPipelineIsScheduled(t *testing.T) {
	runAt := time.Now()
	scheduledPipeline := &Pipeline{RunAt: &runAt}
	notScheduledPipeline := &Pipeline{}

	if !scheduledPipeline.IsScheduled() {
		t.Errorf("IsScheduled returned wrong value: got %t, want true", scheduledPipeline.IsScheduled())
	}
	if notScheduledPipeline.IsScheduled() {
		t.Errorf("IsScheduled returned wrong value: got %t, want false", notScheduledPipeline.IsScheduled())
	}
}
