package domain

// JobResult contains the result of a job.
type JobResult struct {
	Metadata []byte
	Error    error
}

// FutureJobResult is a WorkResult that may not yet
// have become available and can be Wait()'ed on.
type FutureJobResult struct {
	ResultQueue <-chan JobResult
}

// Wait waits for JobResult to become available and returns it.
func (f FutureJobResult) Wait() JobResult {
	r, ok := <-f.ResultQueue
	if !ok {
		// This should never happen, reading from the result
		// channel is exclusive to this future
		panic("failed to read from result channel")
	}
	return r
}
