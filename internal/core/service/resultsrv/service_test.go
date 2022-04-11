package resultsrv

import (
	"errors"
	"reflect"
	"testing"
	"valet/internal/core/domain"
	"valet/mock"
	"valet/pkg/apperrors"

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
	storageErr := errors.New("some storage error")

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetJobResult(expectedResult.JobID).
		Return(expectedResult, nil).
		Times(1)
	storage.
		EXPECT().
		GetJobResult(invalidJobID).
		Return(nil, storageErr).
		Times(1)

	service := New(storage)

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
			"storage error",
			invalidJobID,
			storageErr,
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

	notExistingID := "not_existing_id"
	invalidJobID := "invalid_job_id"
	notFoundErr := &apperrors.NotFoundErr{ID: invalidJobID, ResourceName: "job"}
	storageErr := errors.New("some storage error")

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetJobResult(expectedResult.JobID).
		Return(expectedResult, nil).
		Times(1)
	storage.
		EXPECT().
		GetJobResult(notExistingID).
		Return(nil, notFoundErr).
		Times(1)
	storage.
		EXPECT().
		DeleteJobResult(expectedResult.JobID).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		GetJobResult(invalidJobID).
		Return(expectedResult, nil).
		Times(1)
	storage.
		EXPECT().
		DeleteJobResult(invalidJobID).
		Return(storageErr).
		Times(1)

	service := New(storage)

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
			"not found",
			notExistingID,
			notFoundErr,
		},
		{
			"storage error",
			invalidJobID,
			storageErr,
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
