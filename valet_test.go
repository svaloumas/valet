package valet

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	if err := DecodeTaskParams(args, &testStruct); err != nil {
		t.Errorf("DecodeTaskParaams returned unexpected error: got %s want nil", err)
	}

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

	if err := DecodePreviousJobResults(args, &results); err != nil {
		t.Errorf("DecodePreviousJobResults returned unexpected error: got %s want nil", err)
	}

	if results != "some metadata" {
		t.Errorf("DecodePreviousJobResults decoded wrong results: got %#v want %#v", results, "some metadata")
	}
}

func TestRun(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	v := New("internal/config/testdata/test_config.yaml")
	v.logger = &logrus.Logger{Out: ioutil.Discard}

	v.RegisterTask("dummytask", func(i ...interface{}) (interface{}, error) {
		return "some_metadata", nil
	})

	go v.Run()
	// Give some time for the server to run.
	time.Sleep(200 * time.Millisecond)
	res, err := http.Get("http://localhost:8080/api/status")
	if err != nil {
		t.Errorf("error calling valet status endpoint: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("valet server responded with wrong status code: got %v want %v", res.StatusCode, http.StatusOK)
	}
	v.Stop()
}
