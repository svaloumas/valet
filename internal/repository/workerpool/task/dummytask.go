package task

import (
	"encoding/json"
	"log"
	"time"

	"github.com/mitchellh/mapstructure"
)

// DummyMetadata is an example of a task metadata structure.
type DummyMetadata struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(metadata interface{}) ([]byte, error) {
	taskMetadata := &DummyMetadata{}
	mapstructure.Decode(metadata, taskMetadata)

	log.Println("Hello from dummy task")

	taskMetadata.URL = "http://www.test-url.com"
	time.Sleep(1 * time.Second)
	serializedMetadata, err := json.Marshal(taskMetadata)
	if err != nil {
		return nil, err
	}
	return serializedMetadata, nil
}
