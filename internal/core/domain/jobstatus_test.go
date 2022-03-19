package domain

import (
	"reflect"
	"testing"
)

func TestJobStatusString(t *testing.T) {
	tests := []struct {
		js       JobStatus
		expected string
	}{
		{
			JobStatus(Pending),
			"PENDING",
		},
		{
			JobStatus(InProgress),
			"IN_PROGRESS",
		},
		{
			JobStatus(Completed),
			"COMPLETED",
		},
		{
			JobStatus(Failed),
			"FAILED",
		},
	}

	for _, tt := range tests {
		jsString := tt.js.String()
		if jsString != tt.expected {
			t.Fatalf("JobStatus String returned wrong string: got %s want %s", jsString, tt.expected)
		}
	}
}

func TestJobStatusMarshalJSON(t *testing.T) {
	tests := []struct {
		js       JobStatus
		expected []byte
	}{
		{
			JobStatus(Pending),
			[]byte("\"PENDING\""),
		},
		{
			JobStatus(InProgress),
			[]byte("\"IN_PROGRESS\""),
		},
		{
			JobStatus(Completed),
			[]byte("\"COMPLETED\""),
		},
		{
			JobStatus(Failed),
			[]byte("\"FAILED\""),
		},
	}

	for _, tt := range tests {
		marshaledJobStatus, _ := tt.js.MarshalJSON()
		if eq := reflect.DeepEqual(marshaledJobStatus, tt.expected); !eq {
			t.Fatalf("JobStatus MarshalJSON returned wrong []byte: got %s want %s", marshaledJobStatus, tt.expected)
		}
	}
}

func TestJobStatusUnmarshalJSON(t *testing.T) {
	tests := []struct {
		data     []byte
		expected JobStatus
	}{
		{
			[]byte("\"PENDING\""),
			JobStatus(Pending),
		},
		{
			[]byte("\"IN_PROGRESS\""),
			JobStatus(InProgress),
		},
		{
			[]byte("\"COMPLETED\""),
			JobStatus(Completed),
		},
		{
			[]byte("\"FAILED\""),
			JobStatus(Failed),
		},
	}

	for _, tt := range tests {
		js := JobStatus(0)
		err := js.UnmarshalJSON(tt.data)
		if err != nil {
			t.Fatalf("JobStatus UnmarshalJSON returned unexpected error: %s", err)
		}
		if eq := reflect.DeepEqual(js, tt.expected); !eq {
			t.Fatalf("JobStatus UnmarshalJSON returned wrong JobStatus: got %s want %s", js, tt.expected)
		}
	}
}

func TestJobStatusValidate(t *testing.T) {
	tests := []struct {
		js       JobStatus
		expected string
	}{
		{
			JobStatus(Undefined),
			"0 is not a valid job status, valid statuses: map[PENDING:1 IN_PROGRESS:2 COMPLETED:3 FAILED:4]",
		},
		{
			JobStatus(Pending),
			"",
		},
		{
			JobStatus(7),
			"7 is not a valid job status, valid statuses: map[PENDING:1 IN_PROGRESS:2 COMPLETED:3 FAILED:4]",
		},
	}

	for _, tt := range tests {
		err := tt.js.Validate()
		if err != nil && err.Error() != tt.expected {
			t.Fatalf("JobStatus Validate returned wrong error message: got %s want %s", err.Error(), tt.expected)
		}
	}
}

func TestJobStatusIndex(t *testing.T) {
	tests := []struct {
		js       JobStatus
		expected int
	}{
		{
			JobStatus(Pending),
			1,
		},
		{
			JobStatus(InProgress),
			2,
		},
		{
			JobStatus(Completed),
			3,
		},
		{
			JobStatus(Failed),
			4,
		},
	}

	for _, tt := range tests {
		index := tt.js.Index()
		if eq := reflect.DeepEqual(index, tt.expected); !eq {
			t.Fatalf("JobStatus MarshalJSON returned wrong int: got %d want %d", index, tt.expected)
		}
	}
}
