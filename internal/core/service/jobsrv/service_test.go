package jobsrv

import (
	"errors"
	"reflect"
	"testing"
	"time"
	"valet/internal/core/domain"
	"valet/internal/repository/workerpool/task"
	"valet/mock"
	"valet/pkg/apperrors"

	"github.com/golang/mock/gomock"
)

func TestCreateErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(4)

	createdAt := freezed.Now()
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		Description: "some description",
		Metadata:    &task.DummyMetadata{},
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	jobRepositoryErr := errors.New("some job repository error")
	jobQueueErr := &apperrors.FullQueueErr{}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expectedJob.ID, nil).
		Times(3)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expectedJob).
		Return(jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(expectedJob).
		Return(true).
		Times(1)
	jobQueue.
		EXPECT().
		Push(expectedJob).
		Return(false).
		Times(1)

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		name string
		err  error
	}{
		{
			"",
			jobValidateErr,
		},
		{
			"job_name",
			jobRepositoryErr,
		},
		{
			"job_name",
			jobQueueErr,
		},
		{
			"job_name",
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		_, err := service.Create(tt.name, expectedJob.Description, expectedJob.Metadata)
		if err == nil {
			t.Error("service created expected error, returned nil instead")
		}
		if err.Error() != tt.err.Error() {
			t.Errorf("service create returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
		}
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		Description: "some description",
		Metadata:    &task.DummyMetadata{},
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expectedJob.ID, nil).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expectedJob).
		Return(nil).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(expectedJob).
		Return(true).
		Times(1)

	service := New(jobRepository, jobQueue, uuidGen, freezed)
	j, err := service.Create(expectedJob.Name, expectedJob.Description, expectedJob.Metadata)
	if err != nil {
		t.Errorf("service create returned unexpected error: %#v", err)
	}
	if eq := reflect.DeepEqual(j, expectedJob); !eq {
		t.Errorf("service create returned wrong job, got %#v want %#v", j, expectedJob)
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
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
		Get(expectedJob.ID).
		Return(expectedJob, nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		id  string
		err error
	}{
		{
			expectedJob.ID,
			nil,
		},
		{
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		j, err := service.Get(tt.id)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service get returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		} else {
			if eq := reflect.DeepEqual(j, expectedJob); !eq {
				t.Errorf("service get returned wrong job: got %#v want %#v", j, expectedJob)
			}
		}
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	updatedJob := &domain.Job{}
	*updatedJob = *expectedJob
	updatedJob.Name = "updated job_name"
	updatedJob.Description = "updated description"

	invalidID := "invalid_id"

	jobRepositoryErr := errors.New("some job repository error")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Get(expectedJob.ID).
		Return(expectedJob, nil).
		Times(2)
	jobRepository.
		EXPECT().
		Update(expectedJob.ID, updatedJob).
		Return(jobRepositoryErr).
		Times(1)
	jobRepository.
		EXPECT().
		Update(expectedJob.ID, updatedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		id  string
		err error
	}{
		{
			expectedJob.ID,
			jobRepositoryErr,
		},
		{
			expectedJob.ID,
			nil,
		},
		{
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		err := service.Update(tt.id, updatedJob.Name, updatedJob.Description)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service update returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		}
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

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		id  string
		err error
	}{
		{
			expectedJob.ID,
			nil,
		},
		{
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		err := service.Delete(tt.id)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service delete returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		}
	}
}

func TestExec(t *testing.T) {
	// TODO: Implement this.
}
