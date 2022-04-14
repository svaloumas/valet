package task

import (
	"reflect"
	"testing"
)

func TestDummyTask(t *testing.T) {
	params := map[string]interface{}{
		"url": "some-url.com",
	}

	metadata, err := DummyTask(params)
	if err != nil {
		t.Errorf("dummytask returned unexpected error: got %#v want nil", err)
	}

	expected := "some metadata"

	if eq := reflect.DeepEqual(metadata, expected); !eq {
		t.Errorf("dummytask returned wrong metadata: got %#v want %#v", metadata, expected)
	}
}
