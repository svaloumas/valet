package resultsrv

import (
	"errors"
	"reflect"
	"testing"
	"valet/internal/core/domain"
	"valet/mock"

	"github.com/golang/mock/gomock"
)

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedResult := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	invalidJobID := "invalid_job_id"
	resultRepositoryErr := errors.New("some job result repository error")

	resultRepository := mock.NewMockResultRepository(ctrl)
	resultRepository.
		EXPECT().
		Get(expectedResult.JobID).
		Return(expectedResult, nil).
		Times(1)
	resultRepository.
		EXPECT().
		Get(invalidJobID).
		Return(nil, resultRepositoryErr).
		Times(1)

	service := New(resultRepository)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expectedResult.JobID,
			nil,
		},
		{
			"repository error",
			invalidJobID,
			resultRepositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Get(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service get returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if eq := reflect.DeepEqual(result, expectedResult); !eq {
					t.Errorf("service get returned wrong job: got %#v want %#v", result, expectedResult)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedResult := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	invalidJobID := "invalid_job_id"
	resultRepositoryErr := errors.New("some job result repository error")

	resultRepository := mock.NewMockResultRepository(ctrl)
	resultRepository.
		EXPECT().
		Delete(expectedResult.JobID).
		Return(nil).
		Times(1)
	resultRepository.
		EXPECT().
		Delete(invalidJobID).
		Return(resultRepositoryErr).
		Times(1)

	service := New(resultRepository)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expectedResult.JobID,
			nil,
		},
		{
			"repository error",
			invalidJobID,
			resultRepositoryErr,
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
