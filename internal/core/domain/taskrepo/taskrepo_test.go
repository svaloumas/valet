package taskrepo

import (
	"errors"
	"reflect"
	"runtime"
	"testing"
)

func TestRegister(t *testing.T) {
	taskName := "test_task"
	testTaskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	taskrepo := NewTaskRepository()
	taskrepo.Register(taskName, testTaskFunc)

	taskFunc := taskrepo.tasks[taskName]
	taskFuncName := runtime.FuncForPC(reflect.ValueOf(taskFunc).Pointer()).Name()
	expected := runtime.FuncForPC(reflect.ValueOf(testTaskFunc).Pointer()).Name()

	if taskFuncName != expected {
		t.Errorf("task repository Register did not register the task properly, got %v want %v", taskFuncName, expected)
	}
}

func TestGetTaskFunc(t *testing.T) {
	taskName := "test_task"
	testTaskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	taskrepo := NewTaskRepository()
	taskrepo.tasks[taskName] = testTaskFunc

	taskFunc, err := taskrepo.GetTaskFunc(taskName)
	if err != nil {
		t.Errorf("tesk repository GetTaskFunc returned error: got %#v want nil", err)
	}
	taskFuncName := runtime.FuncForPC(reflect.ValueOf(taskFunc).Pointer()).Name()
	expected := runtime.FuncForPC(reflect.ValueOf(testTaskFunc).Pointer()).Name()

	if taskFuncName != expected {
		t.Errorf("task repository Register did not register the task properly, got %v want %v", taskFuncName, expected)
	}
}

func TestGetTaskNames(t *testing.T) {
	aTaskName := "some task name"
	otherTaskName := "another task name"
	testTaskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	taskrepo := NewTaskRepository()
	taskrepo.tasks[aTaskName] = testTaskFunc
	taskrepo.tasks[otherTaskName] = testTaskFunc

	expectedNames := []string{aTaskName, otherTaskName}
	actualNames := taskrepo.GetTaskNames()

	names := make(map[string]bool)
	for _, name := range actualNames {
		names[name] = true
	}
	for _, expected := range expectedNames {
		if ok := names[expected]; !ok {
			t.Errorf("task repository GetTaskNames returned wrong name: %#v", expected)
		}
	}
}
