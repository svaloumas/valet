package domain

type JobItem struct {
	Job    *Job
	Result chan JobResult
}
