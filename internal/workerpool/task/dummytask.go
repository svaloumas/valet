package task

import (
	"log"
	"time"
	"valet/internal/core/port"
)

type dummytask struct{}

func NewDummyTask() *dummytask {
	return &dummytask{}
}

func (dt *dummytask) Run(metadata port.Metadata) ([]byte, error) {
	log.Printf("[dummy task] hello from dummy task")
	log.Printf("metadata: %v", metadata)
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
