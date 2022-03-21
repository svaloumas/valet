package task

import (
	"log"
	"time"

	"github.com/mitchellh/mapstructure"
)

type DummyMetadata struct {
	URL string `json:"url,omitempty"`
}

func DummyTask(metadata interface{}) ([]byte, error) {
	taskMetadata := &DummyMetadata{}
	mapstructure.Decode(metadata, taskMetadata)

	log.Println("hello from dummy task")
	log.Printf("metadata: %s", taskMetadata.URL)
	time.Sleep(1)
	return make([]byte, 0), nil
}
