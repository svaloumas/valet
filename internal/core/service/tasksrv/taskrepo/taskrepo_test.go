package taskrepo

import (
	"errors"
	"reflect"
	"runtime"
	"testing"
)

func TestGetTaskFunc(t *testing.T) {
	taskName := "test_task"
	testTaskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	repo := make(map[string]TaskFunc)
	repo[taskName] = testTaskFunc
	taskrepo := TaskRepository(repo)

	taskFunc, err := taskrepo.GetTaskFunc(taskName)
	if err != nil {
		t.Errorf("task repository GetTaskFunc returned error: got %#v want nil", err)
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
	testTaskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	repo := make(map[string]TaskFunc)
	repo[aTaskName] = testTaskFunc
	repo[otherTaskName] = testTaskFunc
	taskrepo := TaskRepository(repo)

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

func TestRegister(t *testing.T) {
	taskName := "test_task"
	testTaskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}

	repo := make(map[string]TaskFunc)
	taskrepo := TaskRepository(repo)
	taskrepo.Register(taskName, testTaskFunc)

	taskFunc, ok := taskrepo[taskName]
	if !ok {
		t.Errorf("task repository register did not register the task")
	}
	taskFuncName := runtime.FuncForPC(reflect.ValueOf(taskFunc).Pointer()).Name()
	expected := runtime.FuncForPC(reflect.ValueOf(testTaskFunc).Pointer()).Name()

	if taskFuncName != expected {
		t.Errorf("task repository Register did not register the task properly, got %v want %v", taskFuncName, expected)
	}
}

func TestGetTaskFuncTaskNotRegistered(t *testing.T) {
	taskrepo := New()
	_, err := taskrepo.GetTaskFunc("sometask")

	expected := errors.New("task with name: sometask is not registered")
	if err == nil {
		t.Errorf("get task func did not return expected error: got nil, want %v", expected)
	}
	if err != nil {
		if err.Error() != expected.Error() {
			t.Errorf("get task func returned wrong error: got %v, want %v", err, expected)
		}
	}
}
