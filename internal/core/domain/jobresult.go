package domain

// JobResult contains the result of a job.
type JobResult struct {
	JobID    string      `json:"job_id"`
	Metadata interface{} `json:"metadata"`
	Error    error       `json:"error,omitempty"`
}
