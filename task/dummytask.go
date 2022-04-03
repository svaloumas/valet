package task

import (
	"log"

	"github.com/mitchellh/mapstructure"
)

// DummyParams is an example of a task params structure.
type DummyParams struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(taskParams interface{}) (interface{}, error) {
	params := &DummyParams{}
	mapstructure.Decode(taskParams, params)

	log.Println("Hello from dummy task")
	params.URL = "http://www.test-url.com"
	metadata, err := downloadContent(params.URL)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func downloadContent(URL string) (string, error) {
	return "some metadata", nil
}
