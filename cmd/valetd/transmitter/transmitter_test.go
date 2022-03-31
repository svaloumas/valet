package transmitter

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/mock"
)

func TestTransmitterTransmit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.ID = "job_1"
	j.TaskName = "test_task"

	jobChan := make(chan *domain.Job, 1)

	resultChan := make(chan domain.JobResult, 1)
	jobItem := domain.NewJobItem(j, resultChan, time.Second)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(jobChan).
		Times(2)
	wp := mock.NewMockWorkerPool(ctrl)
	wp.
		EXPECT().
		Send(jobItem).
		Return(nil).
		Times(1)
	wp.
		EXPECT().
		CreateJobItem(j).
		Return(jobItem).
		Times(1)

	logger := log.New(ioutil.Discard, "", 0)
	transmitter := NewTransmitter(jobQueue, wp, logger)

	go transmitter.Transmit()
	defer transmitter.Stop()

	jobChan <- j

	time.Sleep(10 * time.Millisecond)
}
