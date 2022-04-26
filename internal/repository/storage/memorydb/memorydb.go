package memorydb

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/pkg/apperrors"
)

var _ port.Storage = &memorydb{}

type memorydb struct {
	pipelinedb  map[string][]byte
	jobdb       map[string][]byte
	jobresultdb map[string][]byte
}

// NewMemoryDB creates a new memorydb instance.
func New() *memorydb {
	return &memorydb{
		pipelinedb:  make(map[string][]byte),
		jobdb:       make(map[string][]byte),
		jobresultdb: make(map[string][]byte),
	}
}

// CheckHealth checks if the storage is alive.
func (mem *memorydb) CheckHealth() bool {
	return mem.jobdb != nil && mem.jobresultdb != nil && mem.pipelinedb != nil
}

// Close terminates any storage connections gracefully.
func (mem *memorydb) Close() error {
	mem.jobdb = nil
	mem.jobresultdb = nil
	mem.pipelinedb = nil
	return nil
}

// CreateJob adds a new job to the repository.
func (mem *memorydb) CreateJob(j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	mem.jobdb[j.ID] = serializedJob
	return nil
}

// GetJob fetches a job from the repository.
func (mem *memorydb) GetJob(id string) (*domain.Job, error) {
	serializedJob, ok := mem.jobdb[id]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	j := &domain.Job{}
	if err := json.Unmarshal(serializedJob, j); err != nil {
		return nil, err
	}
	return j, nil
}

// GetJobs fetches all jobs from the repository, optionally filters the jobs by status.
func (mem *memorydb) GetJobs(status domain.JobStatus) ([]*domain.Job, error) {
	jobs := []*domain.Job{}
	for _, serializedJob := range mem.jobdb {
		j := &domain.Job{}
		if err := json.Unmarshal(serializedJob, j); err != nil {
			return nil, err
		}
		if status == domain.Undefined || j.Status == status {
			jobs = append(jobs, j)
		}
	}
	// ORDER BY created_at ASC
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(*jobs[j].CreatedAt)
	})
	return jobs, nil
}

// GetJobsByPipelineID fetches the jobs of the specified pipeline.
func (mem *memorydb) GetJobsByPipelineID(pipelineID string) ([]*domain.Job, error) {
	serializedPipeline, ok := mem.pipelinedb[pipelineID]
	if !ok {
		// Mimic the behavior or the relational databases.
		return []*domain.Job{}, nil
	}
	p := &domain.Pipeline{}
	if err := json.Unmarshal(serializedPipeline, p); err != nil {
		return nil, err
	}
	return p.Jobs, nil
}

// UpdateJob updates a job to the repository.
func (mem *memorydb) UpdateJob(id string, j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	mem.jobdb[j.ID] = serializedJob
	return nil
}

// DeleteJob deletes a job from the repository.
func (mem *memorydb) DeleteJob(id string) error {
	if _, ok := mem.jobdb[id]; !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	delete(mem.jobdb, id)
	// CASCADE
	delete(mem.jobresultdb, id)
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (mem *memorydb) GetDueJobs() ([]*domain.Job, error) {
	dueJobs := []*domain.Job{}
	for _, serializedJob := range mem.jobdb {
		j := &domain.Job{}
		if err := json.Unmarshal(serializedJob, j); err != nil {
			return nil, err
		}
		if j.IsScheduled() {
			if j.RunAt.Before(time.Now()) && j.Status == domain.Pending {
				dueJobs = append(dueJobs, j)
			}
		}
	}
	// ORDER BY run_at ASC
	sort.Slice(dueJobs, func(i, j int) bool {
		return dueJobs[i].RunAt.Before(*dueJobs[j].RunAt)
	})
	return dueJobs, nil
}

// CreateJobResult adds new job result to the repository.
func (mem *memorydb) CreateJobResult(result *domain.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	mem.jobresultdb[result.JobID] = serializedJobResult
	return nil
}

// GetJobResult fetches a job result from the repository.
func (mem *memorydb) GetJobResult(jobID string) (*domain.JobResult, error) {
	serializedJobResult, ok := mem.jobresultdb[jobID]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: jobID, ResourceName: "job result"}
	}
	result := &domain.JobResult{}
	json.Unmarshal(serializedJobResult, result)
	return result, nil
}

// UpdateJobResult updates a job result to the repository.
func (mem *memorydb) UpdateJobResult(jobID string, result *domain.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	mem.jobresultdb[result.JobID] = serializedJobResult
	return nil
}

// DeleteJobResult deletes a job result from the repository.
func (mem *memorydb) DeleteJobResult(id string) error {
	if _, ok := mem.jobresultdb[id]; !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "job result"}
	}
	delete(mem.jobresultdb, id)
	return nil
}

// CreatePipeline adds a new pipeline and of its jobs to the repository.
func (mem *memorydb) CreatePipeline(p *domain.Pipeline) error {
	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		return err
	}
	mem.pipelinedb[p.ID] = serializedPipeline
	for _, j := range p.Jobs {
		serializedJob, err := json.Marshal(j)
		if err != nil {
			return err
		}
		mem.jobdb[j.ID] = serializedJob
	}
	return nil
}

// GetPipeline fetches a pipeline from the repository.
func (mem *memorydb) GetPipeline(id string) (*domain.Pipeline, error) {
	serializedPipeline, ok := mem.pipelinedb[id]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "pipeline"}
	}
	p := &domain.Pipeline{}
	if err := json.Unmarshal(serializedPipeline, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetPipelines fetches all pipelines from the repository, optionally filters the pipelines by status.
func (mem *memorydb) GetPipelines(status domain.JobStatus) ([]*domain.Pipeline, error) {
	pipelines := []*domain.Pipeline{}
	for _, serializedPipeline := range mem.pipelinedb {
		p := &domain.Pipeline{}
		if err := json.Unmarshal(serializedPipeline, p); err != nil {
			return nil, err
		}
		if status == domain.Undefined || p.Status == status {
			pipelines = append(pipelines, p)
		}
	}
	// ORDER BY created_at ASC
	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].CreatedAt.Before(*pipelines[j].CreatedAt)
	})
	return pipelines, nil
}

// UpdatePipeline updates a pipeline to the repository.
func (mem *memorydb) UpdatePipeline(id string, p *domain.Pipeline) error {
	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		return err
	}
	mem.pipelinedb[p.ID] = serializedPipeline
	return nil
}

// DeletePipeline deletes a pipeline and all its jobs from the repository.
func (mem *memorydb) DeletePipeline(id string) error {
	serializedPipeline, ok := mem.pipelinedb[id]
	if !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "pipeline"}
	}
	p := &domain.Pipeline{}
	err := json.Unmarshal(serializedPipeline, &p)
	if err != nil {
		return err
	}
	for _, j := range p.Jobs {
		// CASCADE
		delete(mem.jobdb, j.ID)
		delete(mem.jobresultdb, j.ID)
	}
	delete(mem.pipelinedb, id)
	return nil
}
