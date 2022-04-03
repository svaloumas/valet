package jobqueue

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/mock"
)

func TestFIFOQueuePush(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobqueue := NewFIFOQueue(1)

	if ok := jobqueue.Push(job); !ok {
		t.Errorf("fifoqueue could not push job to queue: got %#v want true", ok)
	}
	if ok := jobqueue.Push(job); ok {
		t.Errorf("fifoqueue pushed job to queue: got %#v want false", ok)
	}
}

func TestFIFOQueuePop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobqueue := NewFIFOQueue(1)

	if ok := jobqueue.Push(expected); !ok {
		t.Errorf("fifoqueue could not push job to queue: got %#v want true", ok)
	}
	job := <-jobqueue.Pop()
	if job == nil {
		t.Errorf("fifoqueue pop did not return job: got nil want %#v", job)
	} else {
		if eq := reflect.DeepEqual(job, expected); !eq {
			t.Errorf("fifoqueue pop returned wrong job: got %#v want %#v", job, expected)
		}
	}
}

func TestFIFOQueueClose(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("fifoqueue send on closed channel did not panic")
		} else {
			expected := "send on closed channel"
			panicMsg := fmt.Sprintf("%s", p)
			if eq := reflect.DeepEqual(panicMsg, expected); !eq {
				t.Errorf("fifoqueue close paniced with wrong message: got %#v want %#v", panicMsg, expected)
			}
		}
	}()

	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams:  "some task params",
		Status:      domain.Pending,
	}

	jobqueue := NewFIFOQueue(1)
	jobqueue.logger = log.New(ioutil.Discard, "", 0)

	jobqueue.Close()
	jobqueue.Push(expected)
}
