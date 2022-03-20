package task

import (
	"log"
	"time"
	"valet/internal/core/domain"
)

type dummytask struct{}

func NewDummyTask() *dummytask {
	return &dummytask{}
}

func (dt *dummytask) Run(j *domain.Job) ([]byte, error) {
	log.Printf("[dummy task] hello from dummy task")
	log.Printf("metadata: %v", j.Metadata)
	time.Sleep(1 * time.Second)
	return make([]byte, 0), nil
}

type dummymetadata struct {
	URL string `json:"url,omitempty"`
}

func NewDummyMetadata() *dummymetadata {
	return &dummymetadata{}
}

func (dm *dummymetadata) TaskMetadata() {}
