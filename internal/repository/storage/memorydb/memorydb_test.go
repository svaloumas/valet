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
	j2 := &domain.Job{
		ID:       "due_job2_id",
		TaskName: "some other task",
		Status:   domain.Pending,
		RunAt:    &runAt,
	}
	scheduledAt := freezed.Now()
	scheduledJob := &domain.Job{
		ID:          "scheduled_job_id",
		TaskName:    "some third task",
		Status:      domain.Scheduled,
		RunAt:       &runAt,
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
	if len(dueJobs) != 2 {
		t.Errorf("GetDueJobs returned wrong number of due jobs: got %#v want 2", len(dueJobs))
	}
	dueJob1 := dueJobs[0]
	found := false
	for _, job := range expected {
		if eq := reflect.DeepEqual(dueJob1, job); eq {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetDueJobs returned wrong due jobs: got %#v want %#v", dueJobs, expected)
	}
	dueJob2 := dueJobs[1]
	found = false
	for _, job := range expected {
		if eq := reflect.DeepEqual(dueJob2, job); eq {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetDueJobs returned wrong due jobs: got %#v want %#v", dueJobs, expected)
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
