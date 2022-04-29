package jobqueue

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/mock"
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
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}

	jobqueue := NewFIFOQueue(1)

	if err := jobqueue.Push(job); err != nil {
		t.Errorf("fifoqueue could not push job to queue: got %#v want nil", err)
	}
	if err := jobqueue.Push(job); err == nil {
		t.Errorf("fifoqueue pushed job to queue: got %#v want some err", err)
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
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}

	jobqueue := NewFIFOQueue(1)

	if err := jobqueue.Push(expected); err != nil {
		t.Errorf("fifoqueue could not push job to queue: got %#v want nil", err)
	}
	job := jobqueue.Pop()
	if job == nil {
		t.Errorf("fifoqueue pop did not return job: got nil want %#v", job)
	} else {
		if eq := reflect.DeepEqual(job, expected); !eq {
			t.Errorf("fifoqueue pop returned wrong job: got %#v want %#v", job, expected)
		}
	}
}

func TestFIFOQueueCheckHealth(t *testing.T) {
	q := NewFIFOQueue(1)
	healthy := q.CheckHealth()

	if healthy != true {
		t.Errorf("expected true, got %v instead", healthy)
	}
}

func TestFIFOQueueClose(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("fifoqueue push on closed channel did not panic")
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
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status: domain.Pending,
	}

	jobqueue := NewFIFOQueue(1)

	jobqueue.Close()
	jobqueue.Push(expected)
}
