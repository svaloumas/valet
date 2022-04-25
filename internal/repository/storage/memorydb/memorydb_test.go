package memorydb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
)

func TestMemoryDBCreateJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}

	memorydb := New()
	err := memorydb.CreateJob(job)
	if err != nil {
		t.Errorf("CreateJob returned error: got %#v want nil", err)
	}

	serializedJob := memorydb.jobdb[job.ID]

	expected, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	if eq := reflect.DeepEqual(serializedJob, expected); !eq {
		t.Errorf("CreateJob stored wrong job: got %#v want %#v", string(serializedJob), string(expected))
	}
}

func TestMemoryDBGetJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}
	invalidID := "invalid_id"

	memorydb := New()
	serializedJob, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[expected.ID] = serializedJob

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expected.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "job"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := memorydb.GetJob(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("GetJob returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if eq := reflect.DeepEqual(job, expected); !eq {
					t.Errorf("GetJob returned wrong job: got %#v want %#v", job, expected)
				}
			}
		})
	}
}

func TestMemoryDBGetJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	runAt := freezed.Now()
	pendingJob := &domain.Job{
		ID:        "pending_job_id",
		TaskName:  "some task",
		Status:    domain.Pending,
		CreatedAt: &createdAt,
		RunAt:     &runAt,
	}
	createdAt2 := createdAt.Add(1 * time.Minute)
	scheduledAt := freezed.Now()
	scheduledJob := &domain.Job{
		ID:          "scheduled_job_id",
		TaskName:    "some other task",
		Status:      domain.Scheduled,
		CreatedAt:   &createdAt2,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt3 := createdAt2.Add(1 * time.Minute)
	completedAt := freezed.Now()
	startedAt := freezed.Now()
	inprogressJob := &domain.Job{
		ID:          "inprogress_job_id",
		TaskName:    "some task",
		Status:      domain.InProgress,
		CreatedAt:   &createdAt3,
		StartedAt:   &startedAt,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt4 := createdAt3.Add(1 * time.Minute)
	completedJob := &domain.Job{
		ID:          "completed_job_id",
		TaskName:    "some task",
		Status:      domain.Completed,
		CreatedAt:   &createdAt4,
		StartedAt:   &startedAt,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
		CompletedAt: &completedAt,
	}
	createdAt5 := createdAt4.Add(1 * time.Minute)
	failedJob := &domain.Job{
		ID:            "failed_job_id",
		TaskName:      "some task",
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		CreatedAt:     &createdAt5,
		StartedAt:     &startedAt,
		RunAt:         &runAt,
		ScheduledAt:   &scheduledAt,
		CompletedAt:   &completedAt,
	}

	memorydb := New()
	serializedPendingJob, err := json.Marshal(pendingJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedScheduledJob, err := json.Marshal(scheduledJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedInProgressJob, err := json.Marshal(inprogressJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedCompletedJob, err := json.Marshal(completedJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedFailedJob, err := json.Marshal(failedJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[pendingJob.ID] = serializedPendingJob
	memorydb.jobdb[scheduledJob.ID] = serializedScheduledJob
	memorydb.jobdb[inprogressJob.ID] = serializedInProgressJob
	memorydb.jobdb[completedJob.ID] = serializedCompletedJob
	memorydb.jobdb[failedJob.ID] = serializedFailedJob

	tests := []struct {
		name     string
		status   domain.JobStatus
		expected []*domain.Job
	}{
		{
			"all",
			domain.Undefined,
			[]*domain.Job{
				pendingJob,
				scheduledJob,
				inprogressJob,
				completedJob,
				failedJob,
			},
		},
		{
			"pending",
			domain.Pending,
			[]*domain.Job{
				pendingJob,
			},
		},
		{
			"scheduled",
			domain.Scheduled,
			[]*domain.Job{
				scheduledJob,
			},
		},
		{
			"in progress",
			domain.InProgress,
			[]*domain.Job{
				inprogressJob,
			},
		},
		{
			"completed",
			domain.Completed,
			[]*domain.Job{
				completedJob,
			},
		},
		{
			"failed",
			domain.Failed,
			[]*domain.Job{
				failedJob,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jobs, err := memorydb.GetJobs(tt.status)
			if err != nil {
				t.Errorf("GetJobs returned unexpected error: got %#v want nil", err)
			}

			if len(tt.expected) != len(jobs) {
				t.Fatalf("expected %#v jobs got %#v instead", len(tt.expected), len(jobs))
			}

			for i := range tt.expected {
				if !reflect.DeepEqual(tt.expected[i], jobs[i]) {
					t.Fatalf("expected %#v got %#v instead", tt.expected[i], jobs[i])
				}
			}
		})
	}
}

func TestMemoryDBGetJobsByPipelineID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testTime := freezed.Now()
	job1 := &domain.Job{
		ID:        "job1_id",
		TaskName:  "some task",
		Status:    domain.Pending,
		CreatedAt: &testTime,
		RunAt:     &testTime,
	}
	createdAt2 := testTime.Add(1 * time.Minute)
	job2 := &domain.Job{
		ID:        "job2_id",
		TaskName:  "some other task",
		Status:    domain.Scheduled,
		CreatedAt: &createdAt2,
		RunAt:     &testTime,
	}
	createdAt3 := createdAt2.Add(1 * time.Minute)
	job3 := &domain.Job{
		ID:        "job3_id",
		TaskName:  "some task",
		Status:    domain.InProgress,
		CreatedAt: &createdAt3,
		StartedAt: &testTime,
		RunAt:     &testTime,
	}

	notExistingPipelineID := "invalid_id"

	job1.NextJobID = job2.ID
	job2.NextJobID = job3.ID

	pipelineID := "pipeline_id"
	job1.PipelineID = pipelineID
	job2.PipelineID = pipelineID
	job3.PipelineID = pipelineID

	jobs := []*domain.Job{job1, job2, job3}

	p := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "pipeline_name",
		Description: "some description",
		Jobs:        jobs,
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}
	memorydb := New()
	err := memorydb.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	tests := []struct {
		name       string
		pipelineID string
		expected   []*domain.Job
	}{
		{
			"ok",
			p.ID,
			[]*domain.Job{
				job1,
				job2,
				job3,
			},
		},
		{
			"empty jobs list",
			notExistingPipelineID,
			[]*domain.Job{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jobs, err := memorydb.GetJobsByPipelineID(tt.pipelineID)
			if err != nil {
				t.Errorf("GetJobsByPipelineID returned unexpected error: got %#v want nil", err)
			}

			if len(tt.expected) != len(jobs) {
				t.Fatalf("expected %#v jobs got %#v instead", len(tt.expected), len(jobs))
			}

			for i := range tt.expected {
				if !reflect.DeepEqual(tt.expected[i], jobs[i]) {
					t.Fatalf("expected %#v got %#v instead", tt.expected[i], jobs[i])
				}
			}
		})
	}
}
func TestMemoryDBUpdateJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}

	memorydb := New()
	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[job.ID] = serializedJob

	updatedJob := &domain.Job{}
	*updatedJob = *job
	updatedJob.Name = "updated job_name"
	expected, err := json.Marshal(updatedJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	err = memorydb.UpdateJob(updatedJob.ID, updatedJob)
	if err != nil {
		t.Errorf("UpdateJob returned error: got %#v want nil", err)
	}
	serializedUpdatedJob := memorydb.jobdb[updatedJob.ID]
	if eq := reflect.DeepEqual(serializedUpdatedJob, expected); !eq {
		t.Errorf("UpdateJob updated wrong job: got %#v want %#v", string(serializedUpdatedJob), string(expected))
	}
}

func TestMemoryDBDeleteJob(t *testing.T) {
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status: domain.Pending,
	}
	result := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}
	invalidID := "invalid_id"

	memorydb := New()
	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[job.ID] = serializedJob
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobresultdb[result.JobID] = serializedJobResult

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			job.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "job"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := memorydb.DeleteJob(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("DeleteJob returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if job := memorydb.jobdb[job.ID]; job != nil {
					t.Errorf("DeletJob did not delete job: got %#v want nil", job)
				}
				if result := memorydb.jobresultdb[result.JobID]; result != nil {
					t.Errorf("DeletJob did not delete job result (CASCADE): got %#v want nil", job)
				}
			}
		})
	}
}

func TestMemoryDBGetDueJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)

	runAt := freezed.Now()
	j1 := &domain.Job{
		ID:       "due_job1_id",
		TaskName: "some task",
		Status:   domain.Pending,
		RunAt:    &runAt,
	}
	runAt2 := runAt.Add(1 * time.Minute)
	j2 := &domain.Job{
		ID:       "due_job2_id",
		TaskName: "some other task",
		Status:   domain.Pending,
		RunAt:    &runAt2,
	}
	runAt3 := runAt2.Add(1 * time.Minute)
	scheduledAt := freezed.Now()
	scheduledJob := &domain.Job{
		ID:          "scheduled_job_id",
		TaskName:    "some third task",
		Status:      domain.Scheduled,
		RunAt:       &runAt3,
		ScheduledAt: &scheduledAt,
	}

	expected := []*domain.Job{j1, j2}

	memorydb := New()
	serializedJob1, err := json.Marshal(j1)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedJob2, err := json.Marshal(j2)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedScheduledJob, err := json.Marshal(scheduledJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[j1.ID] = serializedJob1
	memorydb.jobdb[j2.ID] = serializedJob2
	memorydb.jobdb[scheduledJob.ID] = serializedScheduledJob

	dueJobs, err := memorydb.GetDueJobs()
	if err != nil {
		t.Errorf("GetDueJobs returned error: got %#v want nil", err)
	}
	if len(expected) != len(dueJobs) {
		t.Fatalf("expected %#v due jobs got %#v instead", len(expected), len(dueJobs))
	}

	for i := range expected {
		if !reflect.DeepEqual(expected[i], dueJobs[i]) {
			t.Fatalf("expected %#v got %#v instead", expected[i], dueJobs[i])
		}
	}
}

func TestMemoryDBCreateJobResult(t *testing.T) {
	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	memorydb := New()
	err := memorydb.CreateJobResult(result)
	if err != nil {
		t.Errorf("CreateJobResult returned error: got %#v want nil", err)
	}

	serializedResult := memorydb.jobresultdb[result.JobID]

	expected, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	if eq := reflect.DeepEqual(serializedResult, expected); !eq {
		t.Errorf("CreateJobResult stored wrong job: got %#v want %#v", string(serializedResult), string(expected))
	}
}

func TestMemoryDBGetJobResult(t *testing.T) {
	expected := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}
	invalidID := "invalid_id"

	memorydb := New()
	serializedResult, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobresultdb[expected.JobID] = serializedResult

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expected.JobID,
			nil,
		},
		{
			"not found",
			invalidID,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "job result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := memorydb.GetJobResult(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("GetJobResult returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if eq := reflect.DeepEqual(job, expected); !eq {
					t.Errorf("GetJobResult returned wrong job: got %#v want %#v", job, expected)
				}
			}
		})
	}
}

func TestMemoryDBUpdateJobResult(t *testing.T) {
	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	memorydb := New()
	serializedJob, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobresultdb[result.JobID] = serializedJob

	updatedResult := &domain.JobResult{}
	*updatedResult = *result
	updatedResult.Metadata = "updated metadata"
	expected, err := json.Marshal(updatedResult)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	err = memorydb.UpdateJobResult(updatedResult.JobID, updatedResult)
	if err != nil {
		t.Errorf("UpdateJobResult returned error: got %#v want nil", err)
	}
	serializedUpdatedResult := memorydb.jobresultdb[updatedResult.JobID]
	if eq := reflect.DeepEqual(serializedUpdatedResult, expected); !eq {
		t.Errorf("UpdateJobResult updated wrong job: got %#v want %#v", string(serializedUpdatedResult), string(expected))
	}
}

func TestMemoryDBDeleteJobResult(t *testing.T) {
	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}
	invalidID := "invalid_id"

	memorydb := New()
	serializedJob, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobresultdb[result.JobID] = serializedJob

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			result.JobID,
			nil,
		},
		{
			"not found",
			invalidID,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "job"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := memorydb.DeleteJobResult(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("DeleteJobResult returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if job := memorydb.jobresultdb[result.JobID]; job != nil {
					t.Errorf("DeleteJobResult did not delete job: got %#v want nil", job)
				}
			}
		})
	}
}

func TestMemoryDBCreatePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testTime := freezed.Now()
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		PipelineID:  "pipeline_id",
		NextJobID:   "second_job_id",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: false,
		Timeout:            3,
		Status:             domain.Pending,
		FailureReason:      "",
		RunAt:              &testTime,
		ScheduledAt:        &testTime,
		CreatedAt:          &testTime,
		StartedAt:          &testTime,
		CompletedAt:        &testTime,
	}

	later := testTime.Add(1 * time.Minute)
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.ID = "second_job_id"
	secondJob.Name = "second_job_name"
	secondJob.CreatedAt = &later

	jobs := []*domain.Job{job, secondJob}

	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Jobs:        jobs,
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	memorydb := New()
	err := memorydb.CreatePipeline(p)
	if err != nil {
		t.Errorf("CreateJob returned error: got %#v want nil", err)
	}

	serializedPipeline := memorydb.pipelinedb[p.ID]
	serializedJob := memorydb.jobdb[job.ID]
	serializedSecondJob := memorydb.jobdb[secondJob.ID]

	expectedPipeline, err := json.Marshal(p)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	expectedJob, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	expectedSecondJob, err := json.Marshal(secondJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	if eq := reflect.DeepEqual(serializedPipeline, expectedPipeline); !eq {
		t.Errorf("CreateJob stored wrong pipeline: got %#v want %#v", string(serializedPipeline), string(expectedPipeline))
	}
	if eq := reflect.DeepEqual(serializedJob, expectedJob); !eq {
		t.Errorf("CreateJob stored wrong job: got %#v want %#v", string(serializedJob), string(expectedJob))
	}
	if eq := reflect.DeepEqual(serializedSecondJob, expectedSecondJob); !eq {
		t.Errorf("CreateJob stored wrong job: got %#v want %#v", string(serializedSecondJob), string(expectedSecondJob))
	}
}

func TestMemoryDBGetPipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testTime := freezed.Now()
	expected := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}
	invalidID := "invalid_id"

	memorydb := New()
	serializedPipeline, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.pipelinedb[expected.ID] = serializedPipeline

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expected.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "pipeline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := memorydb.GetPipeline(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("GetPipeline returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if eq := reflect.DeepEqual(pipeline, expected); !eq {
					t.Errorf("GetPipeline returned wrong job: got %#v want %#v", pipeline, expected)
				}
			}
		})
	}
}

func TestMemoryDBGetPipelines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)

	createdAt := freezed.Now()
	runAt := freezed.Now()
	pendingPipeline := &domain.Pipeline{
		ID:        "pending_pipeline_id",
		Name:      "pending_pipeline",
		Status:    domain.Pending,
		CreatedAt: &createdAt,
		RunAt:     &runAt,
	}
	later := createdAt.Add(1 * time.Minute)
	inprogressPipeline := &domain.Pipeline{
		ID:        "inprogress_pipeline_id",
		Name:      "inprogress_pipeline",
		Status:    domain.InProgress,
		CreatedAt: &later,
		RunAt:     &runAt,
		StartedAt: &later,
	}

	memorydb := New()
	serializedPendingPipeline, err := json.Marshal(pendingPipeline)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	serializedInProgressPipeline, err := json.Marshal(inprogressPipeline)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.pipelinedb[pendingPipeline.ID] = serializedPendingPipeline
	memorydb.pipelinedb[inprogressPipeline.ID] = serializedInProgressPipeline

	tests := []struct {
		name     string
		status   domain.JobStatus
		expected []*domain.Pipeline
	}{
		{
			"all",
			domain.Undefined,
			[]*domain.Pipeline{
				pendingPipeline,
				inprogressPipeline,
			},
		},
		{
			"pending",
			domain.Pending,
			[]*domain.Pipeline{
				pendingPipeline,
			},
		},
		{
			"inprogress",
			domain.InProgress,
			[]*domain.Pipeline{
				inprogressPipeline,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pipelines, err := memorydb.GetPipelines(tt.status)
			if err != nil {
				t.Errorf("GetPipelines returned unexpected error: got %#v want nil", err)
			}

			if len(tt.expected) != len(pipelines) {
				t.Fatalf("expected %#v pipelines got %#v instead", len(tt.expected), len(pipelines))
			}

			for i := range tt.expected {
				if !reflect.DeepEqual(tt.expected[i], pipelines[i]) {
					t.Fatalf("expected %#v got %#v instead", tt.expected[i], pipelines[i])
				}
			}
		})
	}
}

func TestMemoryDBUpdatePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testTime := freezed.Now()
	p := &domain.Pipeline{
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	memorydb := New()
	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.pipelinedb[p.ID] = serializedPipeline

	updatedPipeline := &domain.Pipeline{}
	*updatedPipeline = *p
	updatedPipeline.Name = "updated pipeline_name"
	expected, err := json.Marshal(updatedPipeline)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	err = memorydb.UpdatePipeline(updatedPipeline.ID, updatedPipeline)
	if err != nil {
		t.Errorf("UpdatePipeline returned error: got %#v want nil", err)
	}
	serializedUpdatedPipeline := memorydb.pipelinedb[updatedPipeline.ID]
	if eq := reflect.DeepEqual(serializedUpdatedPipeline, expected); !eq {
		t.Errorf("UpdatePipeline updated wrong pipeline: got %#v want %#v", string(serializedUpdatedPipeline), string(expected))
	}
}

func TestMemoryDBDeletePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testTime := freezed.Now()
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		PipelineID:  "pipeline_id",
		NextJobID:   "second_job_id",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: false,
		Timeout:            3,
		Status:             domain.Pending,
		FailureReason:      "",
		RunAt:              &testTime,
		ScheduledAt:        &testTime,
		CreatedAt:          &testTime,
		StartedAt:          &testTime,
		CompletedAt:        &testTime,
	}

	later := testTime.Add(1 * time.Minute)
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.ID = "second_job_id"
	secondJob.Name = "second_job_name"
	secondJob.CreatedAt = &later

	jobs := []*domain.Job{job, secondJob}

	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Jobs:        jobs,
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	result1 := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}
	result2 := &domain.JobResult{
		JobID:    secondJob.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}
	invalidID := "invalid_id"

	memorydb := New()
	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[job.ID] = serializedJob

	serializedSecondJob, err := json.Marshal(secondJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobdb[secondJob.ID] = serializedSecondJob

	serializedJobResult1, err := json.Marshal(result1)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobresultdb[result1.JobID] = serializedJobResult1

	serializedJobResult2, err := json.Marshal(result2)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.jobresultdb[result2.JobID] = serializedJobResult2

	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	memorydb.pipelinedb[p.ID] = serializedPipeline

	tests := []struct {
		name       string
		pipelineID string
		err        error
	}{
		{
			"ok",
			p.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "pipeline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := memorydb.DeletePipeline(tt.pipelineID)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("DeletePipeline returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if job := memorydb.jobdb[job.ID]; job != nil {
					t.Errorf("DeletPipeline did not delete job: got %#v want nil", job)
				}
				if job := memorydb.jobdb[secondJob.ID]; job != nil {
					t.Errorf("DeletPipeline did not delete second job: got %#v want nil", job)
				}
				if result := memorydb.jobresultdb[result1.JobID]; result != nil {
					t.Errorf("DeletPipeline did not delete job result1 (CASCADE): got %#v want nil", job)
				}
				if result := memorydb.jobresultdb[result2.JobID]; result != nil {
					t.Errorf("DeletPipeline did not delete job result2 (CASCADE): got %#v want nil", job)
				}
				if pipeline := memorydb.jobresultdb[p.ID]; pipeline != nil {
					t.Errorf("DeletPipeline did not delete pipeline (CASCADE): got %#v want nil", job)
				}
			}
		})
	}
}
