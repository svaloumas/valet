package jobrepo

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

func TestJobDBCreate(t *testing.T) {
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
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobdb := NewJobDB()
	err := jobdb.Create(job)
	if err != nil {
		t.Errorf("jobdb create returned error: got %#v want nil", err)
	}

	serializedJob := jobdb.db[job.ID]

	expected, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	if eq := reflect.DeepEqual(serializedJob, expected); !eq {
		t.Errorf("jobdb create stored wrong job: got %#v want %#v", string(serializedJob), string(expected))
	}
}

func TestJobDBGet(t *testing.T) {
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
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	invalidID := "invalid_id"

	jobdb := NewJobDB()
	serializedJob, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	jobdb.db[expected.ID] = serializedJob

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
			job, err := jobdb.Get(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("jobdb get returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if eq := reflect.DeepEqual(job, expected); !eq {
					t.Errorf("jobdb get returned wrong job: got %#v want %#v", job, expected)
				}
			}
		})
	}
}

func TestJobDBUpdate(t *testing.T) {
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
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobdb := NewJobDB()
	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	jobdb.db[job.ID] = serializedJob

	updatedJob := &domain.Job{}
	*updatedJob = *job
	updatedJob.Name = "updated job_name"
	expected, err := json.Marshal(updatedJob)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	err = jobdb.Update(updatedJob.ID, updatedJob)
	if err != nil {
		t.Errorf("jobdb update returned error: got %#v want nil", err)
	}
	serializedUpdatedJob := jobdb.db[updatedJob.ID]
	if eq := reflect.DeepEqual(serializedUpdatedJob, expected); !eq {
		t.Errorf("jobdb update updated wrong job: got %#v want %#v", string(serializedUpdatedJob), string(expected))
	}
}

func TestJobDBDelete(t *testing.T) {
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams:  "some task params",
		Status:      domain.Pending,
	}
	invalidID := "invalid_id"

	jobdb := NewJobDB()
	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	jobdb.db[job.ID] = serializedJob

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
			err := jobdb.Delete(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("jobdb delete returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if job := jobdb.db[job.ID]; job != nil {
					t.Errorf("jobdb delete did not delete job: got %#v want nil", job)
				}
			}
		})
	}
}

func TestJobDBGetDueJobs(t *testing.T) {
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
		Status:      domain.Failed,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}

	expected := []*domain.Job{j1, j2}

	jobdb := NewJobDB()
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
	jobdb.db[j1.ID] = serializedJob1
	jobdb.db[j2.ID] = serializedJob2
	jobdb.db[scheduledJob.ID] = serializedScheduledJob

	dueJobs, err := jobdb.GetDueJobs()
	if err != nil {
		t.Errorf("jobdb get due jobs returned error: got %#v want nil", err)
	}
	if len(dueJobs) != 2 {
		t.Errorf("jobdb get due jobs returned wrong number of due jobs: got %#v want 2", len(dueJobs))
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
		t.Errorf("jobdb get due jobs returned wrong due jobs: got %#v want %#v", dueJobs, expected)
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
		t.Errorf("jobdb get due jobs returned wrong due jobs: got %#v want %#v", dueJobs, expected)
	}
}
