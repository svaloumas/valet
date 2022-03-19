package jobsrv

import (
	"errors"
	"reflect"
	"testing"
	"time"
	"valet/internal/core/domain"
	"valet/mocks"

	"github.com/golang/mock/gomock"
)

func TestCreateErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mocks.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(3)

	createdAt := freezed.Now()
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	jobRepositoryErr := errors.New("some job repository error")

	uuidGen := mocks.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expectedJob.ID, nil).
		Times(2)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(1)

	jobRepository := mocks.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expectedJob).
		Return(jobRepositoryErr).
		Times(1)

	service := New(jobRepository, uuidGen, freezed)

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
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		_, err := service.Create(tt.name, expectedJob.Description)
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

	freezed := mocks.NewMockTime(ctrl)
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
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mocks.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expectedJob.ID, nil).
		Times(1)

	jobRepository := mocks.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expectedJob).
		Return(nil).
		Times(1)

	service := New(jobRepository, uuidGen, freezed)
	j, err := service.Create(expectedJob.Name, expectedJob.Description)
	if err != nil {
		t.Errorf("create service returned unexpected error: %#v", err)
	}
	if eq := reflect.DeepEqual(j, expectedJob); !eq {
		t.Errorf("create service returned wrong job, got %#v want %#v", j, expectedJob)
	}
}
