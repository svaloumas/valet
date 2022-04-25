package tasksrv

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestRegister(t *testing.T) {
	taskName := "test_task"
	testTaskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	taskService := New()
	taskService.Register(taskName, testTaskFunc)

	taskFunc, _ := taskService.taskrepo.GetTaskFunc(taskName)
	taskFuncName := runtime.FuncForPC(reflect.ValueOf(taskFunc).Pointer()).Name()
	expected := runtime.FuncForPC(reflect.ValueOf(testTaskFunc).Pointer()).Name()

	if taskFuncName != expected {
		t.Errorf("task repository Register did not register the task properly, got %v want %v", taskFuncName, expected)
	}
}

func TestGetTaskRepository(t *testing.T) {
	taskService := New()
	taskrepo := taskService.GetTaskRepository()

	expectedTypeName := "&taskrepo.TaskRepository{}"
	actualTypeName := fmt.Sprintf("%#v", taskrepo)
	if actualTypeName != expectedTypeName {
		t.Errorf("service get task repository did not return the expected type: got %v want %v", actualTypeName, expectedTypeName)
	}
}
