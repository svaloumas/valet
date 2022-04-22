package valet

import (
	"testing"
)

func TestValetRegisterTask(t *testing.T) {

	v := New("internal/config/testdata/test_config.yaml")
	taskrepo := v.taskService.GetTaskRepository()

	if _, err := taskrepo.GetTaskFunc("dummytask"); err == nil {
		t.Errorf("taskrepo did not return expected error")
	}

	v.RegisterTask("dummytask", func(i ...interface{}) (interface{}, error) {
		return nil, nil
	})

	if _, err := taskrepo.GetTaskFunc("dummytask"); err != nil {
		t.Errorf("taskrepo returned unexpected error: got %s want nil", err)
	}
}

func TestDecodeTaskParams(t *testing.T) {
	type test struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	testStruct := test{}

	taskParams := map[string]interface{}{
		"name":        "test_name",
		"description": "test_description",
	}

	args := []interface{}{
		taskParams,
	}

	DecodeTaskParams(args, &testStruct)

	if testStruct.Name != taskParams["name"] {
		t.Errorf("DecodeTaskParaams decoded wrong struct name field: got %#v want #%v", testStruct.Name, taskParams["name"])
	}
	if testStruct.Description != taskParams["description"] {
		t.Errorf("DecodeTaskParaams decoded wrong struct description field: gow %#v want %#v", testStruct.Description, taskParams["description"])
	}
}

func TestDecodePreviousJobResults(t *testing.T) {
	var results string

	taskParams := map[string]interface{}{
		"name":        "test_name",
		"description": "test_description",
	}

	args := []interface{}{
		taskParams,
		"some metadata",
	}

	DecodePreviousJobResults(args, &results)

	if results != "some metadata" {
		t.Errorf("DecodePreviousJobResults decoded wrong results: got %#v want %#v", results, "some metadata")
	}
}
