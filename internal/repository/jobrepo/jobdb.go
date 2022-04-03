package jobrepo

import (
	"encoding/json"
	"time"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
)

var _ port.JobRepository = &jobdb{}

type jobdb struct {
	db map[string][]byte
}

// NewJobDB creates a new jobdb instance.
func NewJobDB() *jobdb {
	return &jobdb{db: make(map[string][]byte)}
}

// Create adds new job to the repository.
func (repo *jobdb) Create(j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	repo.db[j.ID] = serializedJob
	return nil
}

// Get fetches a job from the repository.
func (repo *jobdb) Get(id string) (*domain.Job, error) {
	serializedJob, ok := repo.db[id]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	j := &domain.Job{}
	if err := json.Unmarshal(serializedJob, j); err != nil {
		return nil, err
	}
	return j, nil
}

// Update updates a job to the repository.
func (repo *jobdb) Update(id string, j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	repo.db[j.ID] = serializedJob
	return nil
}

// Delete deletes a job from the repository.
func (repo *jobdb) Delete(id string) error {
	if _, ok := repo.db[id]; !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	delete(repo.db, id)
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (repo *jobdb) GetDueJobs() ([]*domain.Job, error) {
	dueJobs := []*domain.Job{}
	for _, serializedJob := range repo.db {
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
