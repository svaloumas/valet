package resultrepo

import (
	"encoding/json"
	"reflect"
	"testing"

	"valet/internal/core/domain"
	"valet/pkg/apperrors"
)

func TestResultDBCreate(t *testing.T) {
	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	resultdb := NewResultDB()
	err := resultdb.Create(result)
	if err != nil {
		t.Errorf("resultdb create returned error: got %#v want nil", err)
	}

	serializedResult := resultdb.db[result.JobID]

	expected, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	if eq := reflect.DeepEqual(serializedResult, expected); !eq {
		t.Errorf("resultdb create stored wrong job: got %#v want %#v", string(serializedResult), string(expected))
	}
}

func TestResultDBGet(t *testing.T) {
	expected := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}
	invalidID := "invalid_id"

	resultdb := NewResultDB()
	serializedResult, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	resultdb.db[expected.JobID] = serializedResult

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
			job, err := resultdb.Get(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("resultdb get returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if eq := reflect.DeepEqual(job, expected); !eq {
					t.Errorf("resultdb get returned wrong job: got %#v want %#v", job, expected)
				}
			}
		})
	}
}

func TestResultDBUpdate(t *testing.T) {
	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	resultdb := NewResultDB()
	serializedJob, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	resultdb.db[result.JobID] = serializedJob

	updatedResult := &domain.JobResult{}
	*updatedResult = *result
	updatedResult.Metadata = "updated metadata"
	expected, err := json.Marshal(updatedResult)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}

	err = resultdb.Update(updatedResult.JobID, updatedResult)
	if err != nil {
		t.Errorf("resultdb update returned error: got %#v want nil", err)
	}
	serializedUpdatedResult := resultdb.db[updatedResult.JobID]
	if eq := reflect.DeepEqual(serializedUpdatedResult, expected); !eq {
		t.Errorf("resultdb update updated wrong job: got %#v want %#v", string(serializedUpdatedResult), string(expected))
	}
}

func TestResultDBDelete(t *testing.T) {
	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}
	invalidID := "invalid_id"

	resultdb := NewResultDB()
	serializedJob, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json marshal returned error: got %#v want nil", err)
	}
	resultdb.db[result.JobID] = serializedJob

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
			err := resultdb.Delete(tt.id)
			if err != nil {
				errValue, ok := err.(*apperrors.NotFoundErr)
				if !ok {
					t.Errorf("resultdb delete returned wrong error: got %#v want %#v", errValue, tt.err)
				}
			} else {
				if job := resultdb.db[result.JobID]; job != nil {
					t.Errorf("resultdb delete did not delete job: got %#v want nil", job)
				}
			}
		})
	}
}
