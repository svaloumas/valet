package consumersrv

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"valet/internal/core/domain"
	"valet/mock"
)

func TestConsume(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.ID = "job_1"
	j.TaskName = "test_task"
	j.Status = domain.Pending
	j.RunAt = nil

	w := domain.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := &logrus.Logger{Out: ioutil.Discard}

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(j).
		Times(1)
	workService := mock.NewMockWorkService(ctrl)
	workService.
		EXPECT().
		Send(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)

	consumerService := New(jobQueue, workService, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerService.Consume(ctx, 8*time.Millisecond)

	// give some time for the scheduler to consume the job
	time.Sleep(10 * time.Millisecond)
}

func TestConsumeJobJobInQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := &logrus.Logger{Out: ioutil.Discard}

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(nil).
		Times(2)
	workService := mock.NewMockWorkService(ctrl)

	consumerService := New(jobQueue, workService, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerService.Consume(ctx, 5*time.Millisecond)

	// give some time for the scheduler to try to consume two jobs
	time.Sleep(13 * time.Millisecond)
}
