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

	jobs := []*Job{
		{
			Name:     "job_name",
			TaskName: "test_task",
		},
		{
			Name:     "another_job",
			TaskName: "test_task",
		},
	}
	createdAt := time.Now()
	validPipeline := NewPipeline("pipeline_id", "pipeline_name", "some description", jobs, &createdAt)

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
		{"ok", validPipeline, ""},
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

func TestMergeJobsInOne(t *testing.T) {
	firstJob := &Job{
		Name: "first_job",
	}
	secondJob := &Job{
		Name: "second_job",
	}
	thirdJob := &Job{
		Name: "third_job",
	}
	fourthJob := &Job{
		Name: "fourth_job",
	}
	fifthJob := &Job{
		Name: "fifth_job",
	}
	p := &Pipeline{
		Jobs: []*Job{
			firstJob,
			secondJob,
			thirdJob,
			fourthJob,
			fifthJob,
		},
	}

	p.MergeJobsInOne()

	if eq := reflect.DeepEqual(firstJob.Next, secondJob); !eq {
		t.Errorf("MergeJobsInOne did not link the first job's next with the second job: got %#v want %#v", firstJob.Next, secondJob)
	}
	if eq := reflect.DeepEqual(firstJob.Next.Next, thirdJob); !eq {
		t.Errorf("MergeJobsInOne did not link the second job's next with the third job: got %#v want %#v", secondJob.Next, thirdJob)
	}
	if eq := reflect.DeepEqual(firstJob.Next.Next.Next, fourthJob); !eq {
		t.Errorf("MergeJobsInOne did not link the third job's next with the fourth job: got %#v want %#v", thirdJob.Next, fourthJob)
	}
	if eq := reflect.DeepEqual(firstJob.Next.Next.Next.Next, fifthJob); !eq {
		t.Errorf("MergeJobsInOne did not link the fourth job's next with the fifth job: got %#v want %#v", fourthJob.Next, fifthJob)
	}
	if firstJob.Next.Next.Next.Next.Next != nil {
		t.Errorf("MergeJobsInOne linked the last job's next with a job: got %#v want nil", fifthJob.Next)
	}
}

func TestUnmergeJobs(t *testing.T) {
	firstJob := &Job{
		Name: "first_job",
	}
	secondJob := &Job{
		Name: "second_job",
	}
	p := &Pipeline{
		Jobs: []*Job{
			firstJob,
			secondJob,
		},
	}

	p.MergeJobsInOne()

	p.UnmergeJobs()

	if firstJob.Next != nil {
		t.Errorf("UnmergeJobs did not unlink the first job's next with the second job: got %#v want %#v", firstJob.Next, secondJob)
	}
}
