package jobqueue

import (
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
		TaskType:    "test_task",
		Description: "some description",
		Metadata:    "some metadata",
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
