package memorydb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/mock"
	"valet/pkg/apperrors"
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
				t.Fatalf("expected %#v accounts got %#v instead", len(tt.expected), len(jobs))
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
		t.Fatalf("expected %#v accounts got %#v instead", len(expected), len(dueJobs))
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
