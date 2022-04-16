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
	jobdb       map[string][]byte
	jobresultdb map[string][]byte
}

// NewMemoryDB creates a new memorydb instance.
func New() *memorydb {
	return &memorydb{
		jobdb:       make(map[string][]byte),
		jobresultdb: make(map[string][]byte),
	}
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
		if j.RunAt != nil {
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

// CheckHealth checks if the storage is alive.
func (mem *memorydb) CheckHealth() bool {
	return mem.jobdb != nil && mem.jobresultdb != nil
}

// Close terminates any storage connections gracefully.
func (mem *memorydb) Close() error {
	return nil
}
