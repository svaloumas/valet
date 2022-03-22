package resultrepo

import (
	"encoding/json"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/pkg/apperrors"
)

var _ port.ResultRepository = &resultdb{}

type resultdb struct {
	db map[string][]byte
}

// NewResultDB creates a new resultdb instance.
func NewResultDB() *resultdb {
	return &resultdb{db: make(map[string][]byte)}
}

// Create adds new job result to the repository.
func (repo *resultdb) Create(result *domain.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	repo.db[result.JobID] = serializedJobResult
	return nil
}

// Get fetches a job result from the repository.
func (repo *resultdb) Get(id string) (*domain.JobResult, error) {
	serializedJobResult, ok := repo.db[id]
	if !ok {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job result"}
	}
	result := &domain.JobResult{}
	json.Unmarshal(serializedJobResult, result)
	return result, nil
}

// Update updates a job result to the repository.
func (repo *resultdb) Update(id string, result *domain.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	repo.db[result.JobID] = serializedJobResult
	return nil
}

// Delete deletes a job result from the repository.
func (repo *resultdb) Delete(id string) error {
	if _, ok := repo.db[id]; !ok {
		return &apperrors.NotFoundErr{ID: id, ResourceName: "job result"}
	}
	delete(repo.db, id)
	return nil
}
