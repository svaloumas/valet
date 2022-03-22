package domain

type JobItem struct {
	Job         *Job
	ResultQueue chan JobResult
}
