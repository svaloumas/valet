package task

import (
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
	log.Printf("I got some metadata: %#v", taskMetadata)
	time.Sleep(1 * time.Second)
	return make([]byte, 0), nil
}
