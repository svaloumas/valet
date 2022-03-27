package domain

import (
	"fmt"
	"strconv"
)

// JobStatus holds a value for job status ranging from 1 to 5.
type JobStatus int

const (
	Undefined  JobStatus = iota // 0
	Pending                     // 1
	InProgress                  // 2
	Completed                   // 3
	Failed                      // 4

	PENDING     = "PENDING"
	IN_PROGRESS = "IN_PROGRESS"
	COMPLETED   = "COMPLETED"
	FAILED      = "FAILED"
)

// String converts the type to a string.
func (js JobStatus) String() string {
	return [...]string{PENDING, IN_PROGRESS, COMPLETED, FAILED}[js-1]
}

// Index returns the integer representation of a JobStatus.
func (js JobStatus) Index() int {
	return int(js)
}

// Marshaling for JSON representation.
func (js *JobStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + js.String() + `"`), nil
}

// Unmarshaling for JSON representation.
func (js *JobStatus) UnmarshalJSON(data []byte) error {
	var err error
	jobStatuses := map[string]JobStatus{
		PENDING:     Pending,
		IN_PROGRESS: InProgress,
		COMPLETED:   Completed,
		FAILED:      Failed,
	}

	unquotedJobStatus, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("unquoting job status data returned error: %s", err)
	}

	jobStatus, ok := jobStatuses[unquotedJobStatus]
	if !ok {
		return fmt.Errorf("invalid job status: %s", string(data))
	}

	*js = jobStatus
	return err
}

// Validate makes a sanity check on JobStatus.
func (js JobStatus) Validate() error {
	var err error
	validJobStatuses := map[JobStatus]int{
		Pending:    Pending.Index(),
		InProgress: InProgress.Index(),
		Completed:  Completed.Index(),
		Failed:     Failed.Index(),
	}
	if _, ok := validJobStatuses[js]; !ok {
		err = fmt.Errorf("%d is not a valid job status, valid statuses: %v", js, validJobStatuses)
	}
	return err
}
