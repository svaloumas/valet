package domain

import (
	"reflect"
	"testing"
)

func TestJobStatusString(t *testing.T) {
	tests := []struct {
		name     string
		js       JobStatus
		expected string
	}{
		{
			"pending",
			JobStatus(Pending),
			"PENDING",
		},
		{
			"scheduled",
			JobStatus(Scheduled),
			"SCHEDULED",
		},
		{
			"in progress",
			JobStatus(InProgress),
			"IN_PROGRESS",
		},
		{
			"completed",
			JobStatus(Completed),
			"COMPLETED",
		},
		{
			"failed",
			JobStatus(Failed),
			"FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsString := tt.js.String()
			if jsString != tt.expected {
				t.Fatalf("JobStatus String returned wrong string: got %s want %s", jsString, tt.expected)
			}
		})
	}
}

func TestJobStatusMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		js       JobStatus
		expected []byte
	}{
		{
			"pending",
			JobStatus(Pending),
			[]byte("\"PENDING\""),
		},
		{
			"scheduled",
			JobStatus(Scheduled),
			[]byte("\"SCHEDULED\""),
		},
		{
			"in progress",
			JobStatus(InProgress),
			[]byte("\"IN_PROGRESS\""),
		},
		{
			"completed",
			JobStatus(Completed),
			[]byte("\"COMPLETED\""),
		},
		{
			"failed",
			JobStatus(Failed),
			[]byte("\"FAILED\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaledJobStatus, _ := tt.js.MarshalJSON()
			if eq := reflect.DeepEqual(marshaledJobStatus, tt.expected); !eq {
				t.Fatalf("JobStatus MarshalJSON returned wrong []byte: got %s want %s", marshaledJobStatus, tt.expected)
			}
		})
	}
}

func TestJobStatusUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected JobStatus
	}{
		{
			"pending",
			[]byte("\"PENDING\""),
			JobStatus(Pending),
		},
		{
			"scheduled",
			[]byte("\"SCHEDULED\""),
			JobStatus(Scheduled),
		},
		{
			"in progress",
			[]byte("\"IN_PROGRESS\""),
			JobStatus(InProgress),
		},
		{
			"completed",
			[]byte("\"COMPLETED\""),
			JobStatus(Completed),
		},
		{
			"failed",
			[]byte("\"FAILED\""),
			JobStatus(Failed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := JobStatus(0)
			err := js.UnmarshalJSON(tt.data)
			if err != nil {
				t.Fatalf("JobStatus UnmarshalJSON returned unexpected error: %s", err)
			}
			if eq := reflect.DeepEqual(js, tt.expected); !eq {
				t.Fatalf("JobStatus UnmarshalJSON returned wrong JobStatus: got %s want %s", js, tt.expected)
			}
		})
	}
}

func TestJobStatusValidate(t *testing.T) {
	tests := []struct {
		name     string
		js       JobStatus
		expected string
	}{
		{
			"zero status",
			JobStatus(Undefined),
			"0 is not a valid job status, valid statuses: map[PENDING:1 SCHEDULED:2 IN_PROGRESS:3 COMPLETED:4 FAILED:5]",
		},
		{
			"ok",
			JobStatus(Pending),
			"",
		},
		{
			"seven status",
			JobStatus(7),
			"7 is not a valid job status, valid statuses: map[PENDING:1 SCHEDULED:2 IN_PROGRESS:3 COMPLETED:4 FAILED:5]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.js.Validate()
			if err != nil && err.Error() != tt.expected {
				t.Fatalf("JobStatus Validate returned wrong error message: got %s want %s", err.Error(), tt.expected)
			}
		})
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
			JobStatus(Scheduled),
			2,
		},
		{
			JobStatus(InProgress),
			3,
		},
		{
			JobStatus(Completed),
			4,
		},
		{
			JobStatus(Failed),
			5,
		},
	}

	for _, tt := range tests {
		index := tt.js.Index()
		if eq := reflect.DeepEqual(index, tt.expected); !eq {
			t.Fatalf("JobStatus MarshalJSON returned wrong int: got %d want %d", index, tt.expected)
		}
	}
}
