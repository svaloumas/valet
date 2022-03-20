package jobrepo

import (
	"encoding/json"

	"valet/internal/core/domain"
	"valet/internal/repositories"
)

type memdb struct {
	db map[string][]byte
}

func NewMemDB() *memdb {
	return &memdb{db: make(map[string][]byte)}
}

func (repo *memdb) Create(j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	repo.db[j.ID] = serializedJob
	return nil
}

func (repo *memdb) Get(id string) (*domain.Job, error) {
	serializedJob, ok := repo.db[id]
	if !ok {
		return nil, &repositories.NotFoundError{ID: id, ResourceName: "job"}
	}
	j := &domain.Job{}
	json.Unmarshal(serializedJob, j)
	return j, nil
}

func (repo *memdb) Update(id string, j *domain.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	repo.db[j.ID] = serializedJob
	return nil
}

func (repo *memdb) Delete(id string) error {
	if _, ok := repo.db[id]; !ok {
		return &repositories.NotFoundError{ID: id, ResourceName: "job"}
	}
	delete(repo.db, id)
	return nil
}
