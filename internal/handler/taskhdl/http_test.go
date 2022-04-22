package taskhdl

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/mock"
)

func TestHTTPGetTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	gin.SetMode(gin.ReleaseMode)

	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	repo := make(map[string]taskrepo.TaskFunc)
	repo["a_test_task"] = taskFunc
	taskrepo := taskrepo.TaskRepository(repo)
	taskService := mock.NewMockTaskService(ctrl)
	taskService.
		EXPECT().
		GetTaskRepository().
		Return(&taskrepo).
		Times(1)
	hdl := NewTaskHTTPHandler(taskService)

	rr := httptest.NewRecorder()
	c, r := gin.CreateTestContext(rr)

	r.GET("/api/tasks", hdl.GetTasks)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/tasks", nil)

	r.ServeHTTP(rr, c.Request)

	b, _ := ioutil.ReadAll(rr.Body)

	var res map[string]interface{}
	json.Unmarshal(b, &res)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	expected := map[string]interface{}{
		"tasks": []interface{}{
			"a_test_task",
		},
	}
	if eq := reflect.DeepEqual(res, expected); !eq {
		t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
	}
}
