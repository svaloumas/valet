package storage

import (
	"encoding/json"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
)

var _ port.Storage = &memorydb{}

type memorydb struct {
	jobdb       map[string][]byte
	jobresultdb map[string][]byte
}

// NewMemoryDB creates a new memorydb instance.
func NewMemoryDB() *memorydb {
	return &memorydb{
		jobdb:       make(map[string][]byte),
		jobresultdb: make(map[string][]byte),
	}
}

// CreateJob adds new job to the repository.
func (storage *memorydb) CreateJob(j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	storage.jobdb[j.ID] = serializedJob
	return nil
}

// GetJob fetches a job from the repository.
func (storage *memorydb) GetJob(id string) (*domain.Job, error) {
	serializedJob, ok := storage.jobdb[id]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	j := &domain.Job{}
	if err := json.Unmarshal(serializedJob, j); err != nil {
		return nil, err
	}
	return j, nil
}

// UpdateJob updates a job to the repository.
func (storage *memorydb) UpdateJob(id string, j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	storage.jobdb[j.ID] = serializedJob
	return nil
}

// DeleteJob deletes a job from the repository.
func (storage *memorydb) DeleteJob(id string) error {
	if _, ok := storage.jobdb[id]; !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	delete(storage.jobdb, id)
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (storage *memorydb) GetDueJobs() ([]*domain.Job, error) {
	dueJobs := []*domain.Job{}
	for _, serializedJob := range storage.jobdb {
		j := &domain.Job{}
		if err := json.Unmarshal(serializedJob, j); err != nil {
			return nil, err
		}
		if j.RunAt != nil {
			if j.RunAt.Before(time.Now()) && j.ScheduledAt == nil {
				dueJobs = append(dueJobs, j)
			}
		}
	}
	return dueJobs, nil
}

// CreateJobResult adds new job result to the repository.
func (storage *memorydb) CreateJobResult(result *domain.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	storage.jobresultdb[result.JobID] = serializedJobResult
	return nil
}

// GetJobResult fetches a job result from the repository.
func (storage *memorydb) GetJobResult(id string) (*domain.JobResult, error) {
	serializedJobResult, ok := storage.jobresultdb[id]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job result"}
	}
	result := &domain.JobResult{}
	json.Unmarshal(serializedJobResult, result)
	return result, nil
}

// UpdateJobResult updates a job result to the repository.
func (storage *memorydb) UpdateJobResult(id string, result *domain.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	storage.jobresultdb[result.JobID] = serializedJobResult
	return nil
}

// DeleteJobResult deletes a job result from the repository.
func (storage *memorydb) DeleteJobResult(id string) error {
	if _, ok := storage.jobresultdb[id]; !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "job result"}
	}
	delete(storage.jobresultdb, id)
	return nil
}
