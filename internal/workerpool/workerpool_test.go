package wp

import (
	"context"
	"io/ioutil"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
	"valet/internal/core/domain"
	"valet/mock"

	"github.com/golang/mock/gomock"
)

func TestBacklogLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	defer wg.Wait()

	j := new(domain.Job)
	j.TaskName = "test_task"

	resultChan1 := make(chan domain.JobResult, 1)
	jobItem1 := domain.NewJobItem(j, resultChan1, time.Millisecond)

	jDeemedToFail := new(domain.Job)
	jDeemedToFail.TaskName = "test_task"

	resultChan2 := make(chan domain.JobResult, 1)
	jobItemDeemedToFail := domain.NewJobItem(jDeemedToFail, resultChan2, time.Millisecond)
	futureResult1 := domain.FutureJobResult{Result: jobItem1.Result}

	jobService := mock.NewMockJobService(ctrl)
	resultService := mock.NewMockResultService(ctrl)

	wp := NewWorkerPoolImpl(jobService, resultService, 0, 1)
	wp.Log = log.New(ioutil.Discard, "", 0)
	wp.Start()
	defer wp.Stop()

	wg.Add(1)
	resultService.
		EXPECT().
		Create(futureResult1).
		DoAndReturn(func(futureResult domain.FutureJobResult) error {
			wg.Done()
			return nil
		})

	err := wp.Send(jobItem1)
	if err != nil {
		t.Errorf("Error sending job to workerpool: %v", err)
	}

	err = wp.Send(jobItemDeemedToFail)
	if err == nil {
		t.Fatal("Expected error, due to backlog being full")
	}
}

func TestConcurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	defer wg.Wait()

	jobService := mock.NewMockJobService(ctrl)
	resultService := mock.NewMockResultService(ctrl)

	wp := NewWorkerPoolImpl(jobService, resultService, 1, 5)
	wp.Log = log.New(ioutil.Discard, "", 0)
	wp.Start()
	defer wp.Stop()

	j1 := new(domain.Job)
	j1.ID = "job_1"
	j1.TaskName = "test_task"

	resultChan1 := make(chan domain.JobResult, 1)
	jobItem1 := domain.NewJobItem(j1, resultChan1, time.Millisecond)

	j2 := new(domain.Job)
	j2.ID = "job_2"
	j2.TaskName = "test_task"

	resultChan2 := make(chan domain.JobResult, 1)
	jobItem2 := domain.NewJobItem(j2, resultChan2, time.Millisecond)

	futureResult1 := domain.FutureJobResult{Result: jobItem1.Result}
	futureResult2 := domain.FutureJobResult{Result: jobItem2.Result}
	resultService.
		EXPECT().
		Create(futureResult1).
		Return(nil).
		Times(1)
	resultService.
		EXPECT().
		Create(futureResult2).
		Return(nil).
		Times(1)
	wg.Add(1)
	jobService.
		EXPECT().
		Exec(context.Background(), jobItem1).
		DoAndReturn(func(ctx context.Context, jobItem1 domain.JobItem) error {
			time.Sleep(50 * time.Millisecond)
			wg.Done()
			return nil
		})

	err := wp.Send(jobItem1)
	if err != nil {
		t.Errorf("Error sending job to worker: %v", err)
	}

	// sleep enough for the workerpool to start, but not as much to finish its first job
	time.Sleep(20 * time.Millisecond)

	err = wp.Send(jobItem2)
	if err != nil {
		t.Errorf("Error sending job to worker: %v", err)
	}

	// the queue should contain only 1 item, the job item for the 2nd job
	queueLength := len(wp.queue)
	if queueLength != 1 {
		t.Errorf("Expected queue with 1 item, actual length was %#v", queueLength)
	}
	select {
	case i, ok := <-wp.queue:
		if !ok {
			t.Errorf("Unexpectedly closed workerpool queue")
		}
		if eq := reflect.DeepEqual(i.Job, j2); !eq {
			t.Errorf("Expected %#v to be %#v", i.Job, j2)
		}
	default:
		t.Errorf("Expected to find a job item in the queue")
	}
}
