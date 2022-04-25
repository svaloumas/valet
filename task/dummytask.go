package task

import (
	"github.com/svaloumas/valet"
)

// DummyParams is an example of a task params structure.
type DummyParams struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(args ...interface{}) (interface{}, error) {
	dummyParams := &DummyParams{}
	var resultsMetadata string
	valet.DecodeTaskParams(args, dummyParams)
	valet.DecodePreviousJobResults(args, &resultsMetadata)

	metadata := downloadContent(dummyParams.URL, resultsMetadata)
	return metadata, nil
}

func downloadContent(URL, bucket string) string {
	return "some metadata"
}
