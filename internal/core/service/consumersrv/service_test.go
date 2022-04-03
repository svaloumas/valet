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

	go consumerService.Consume()
	defer consumerService.Stop()

	jobChan <- j

	time.Sleep(10 * time.Millisecond)
}
