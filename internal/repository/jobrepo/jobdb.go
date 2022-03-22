package jobrepo

import (
	"encoding/json"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/internal/repository"
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
		return nil, &repository.NotFoundErr{ID: id, ResourceName: "job"}
	}
	j := &domain.Job{}
	json.Unmarshal(serializedJob, j)
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
		return &repository.NotFoundErr{ID: id, ResourceName: "job"}
	}
	delete(repo.db, id)
	return nil
}
