package jobsrv

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/mock"
	"valet/pkg/apperrors"
)

func TestCreateErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	jobRepositoryErr := errors.New("some job repository error")
	jobQueueErr := &apperrors.FullQueueErr{}
	jobTaskNameErr := &apperrors.ResourceValidationErr{Message: "wrongtask is not a valid task name - valid tasks: [test_task]"}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(4)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(job).
		Return(jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(job).
		Return(true).
		Times(1)
	jobQueue.
		EXPECT().
		Push(job).
		Return(false).
		Times(1)

	taskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)
	service := New(jobRepository, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name     string
		jobname  string
		taskName string
		err      error
	}{
		{
			"job validation error",
			"",
			"test_task",
			jobValidateErr,
		},
		{
			"job repository error",
			"job_name",
			"test_task",
			jobRepositoryErr,
		},
		{
			"job queue error",
			"job_name",
			"test_task",
			jobQueueErr,
		},
		{
			"job task type error",
			"job_name",
			"wrongtask",
			jobTaskNameErr,
		},
		{
			"uuid generator error",
			"job_name",
			"test_task",
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(tt.jobname, tt.taskName, job.Description, job.Timeout, job.TaskParams)
			if err == nil {
				t.Error("service created expected error, returned nil instead")
			}
			if err.Error() != tt.err.Error() {
				t.Errorf("service create returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)

	createdAt := freezed.Now()
	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expected.ID, nil).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expected).
		Return(nil).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(expected).
		Return(true).
		Times(1)

	taskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)

	service := New(jobRepository, jobQueue, taskrepo, uuidGen, freezed)

	j, err := service.Create(expected.Name, expected.TaskName, expected.Description, expected.Timeout, expected.TaskParams)
	if err != nil {
		t.Errorf("service create returned unexpected error: %#v", err)
	}
	if eq := reflect.DeepEqual(j, expected); !eq {
		t.Errorf("service create returned wrong job, got %#v want %#v", j, expected)
	}
}

func TestGet(t *testing.T) {
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
		TaskParams:  "some task params",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobRepositoryErr := errors.New("some repository error")
	invalidID := "invalid_id"
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Get(expected.ID).
		Return(expected, nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	taskrepo := taskrepo.NewTaskRepository()
	service := New(jobRepository, jobQueue, taskrepo, uuidGen, freezed)

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
			"job repository error",
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j, err := service.Get(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service get returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if eq := reflect.DeepEqual(j, expected); !eq {
					t.Errorf("service get returned wrong job: got %#v want %#v", j, expected)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
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
		TaskParams:  "some task params",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	updatedJob := &domain.Job{}
	*updatedJob = *job
	updatedJob.Name = "updated job_name"
	updatedJob.Description = "updated description"

	invalidID := "invalid_id"

	jobRepositoryErr := errors.New("some job repository error")
	jobNotFoundErr := errors.New("job not found")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Get(job.ID).
		Return(job, nil).
		Times(2)
	jobRepository.
		EXPECT().
		Update(job.ID, updatedJob).
		Return(jobRepositoryErr).
		Times(1)
	jobRepository.
		EXPECT().
		Update(job.ID, updatedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobNotFoundErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	taskrepo := taskrepo.NewTaskRepository()
	service := New(jobRepository, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"job repository error",
			job.ID,
			jobRepositoryErr,
		},
		{
			"ok",
			job.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			jobNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(tt.id, updatedJob.Name, updatedJob.Description)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service update returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		TaskParams:  "some task params",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	invalidID := "invalid_id"
	jobRepositoryErr := errors.New("some repository error")
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Delete(expectedJob.ID).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Delete(invalidID).
		Return(jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	taskrepo := taskrepo.NewTaskRepository()
	service := New(jobRepository, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expectedJob.ID,
			nil,
		},
		{
			"job repository error",
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service delete returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			}
		})
	}
}
