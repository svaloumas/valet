package domain

// JobResult contains the result of a job.
type JobResult struct {
	JobID    string      `json:"job_id"`
	Metadata interface{} `json:"metadata,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// FutureJobResult is a JobResult that may not yet
// have become available and can be Wait()'ed on.
type FutureJobResult struct {
	Result chan JobResult
}

// Wait waits for JobResult to become available and returns it.
func (f FutureJobResult) Wait() JobResult {
	return <-f.Result
}
