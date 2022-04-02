package consumersrv

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/mock"
)

func TestConsume(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.ID = "job_1"
	j.TaskName = "test_task"

	jobChan := make(chan *domain.Job, 1)

	w := domain.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := log.New(ioutil.Discard, "", 0)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(jobChan).
		Times(2)
	jobQueue.
		EXPECT().
		Close().
		Return().
		Times(1)
	wp := mock.NewMockWorkService(ctrl)
	wp.
		EXPECT().
		Start().
		Return().
		Times(1)
	wp.
		EXPECT().
		Send(w).
		Return(nil).
		Times(1)
	wp.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)
	wp.
		EXPECT().
		Stop().
		Return().
		Times(1)

	consumerService := New(jobQueue, wp, logger)

	go consumerService.Consume()
	defer consumerService.Stop()

	jobChan <- j

	time.Sleep(10 * time.Millisecond)
}
