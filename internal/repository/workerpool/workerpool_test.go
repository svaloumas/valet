package workerpool

import (
	"io/ioutil"
	"log"
	"testing"
	"valet/internal/core/domain"
	"valet/internal/core/domain/task"
	"valet/mock"

	"github.com/golang/mock/gomock"
)

func TestBacklogLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobService := mock.NewMockJobService(ctrl)
	resultService := mock.NewMockResultService(ctrl)
	taskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", nil
	}
	taskrepo := task.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)

	wp := NewWorkerPoolImpl(jobService, resultService, taskrepo, 0, 1)
	wp.logger = log.New(ioutil.Discard, "", 0)
	wp.Start()
	defer wp.Stop()

	j := new(domain.Job)
	err := wp.Send(j)
	if err != nil {
		t.Errorf("Error sending job to workerpool: %v", err)
	}

	jDeemedToFail := new(domain.Job)
	err = wp.Send(jDeemedToFail)
	if err == nil {
		t.Fatal("Expected error, due to backlog being full")
	}
}

// func TestConcurrency(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	freezed := mock.NewMockTime(ctrl)
// 	freezed.
// 		EXPECT().
// 		Now().
// 		Return(time.Date(1985, 11, 04, 20, 34, 58, 651387237, time.UTC)).
// 		Times(2)

// 	jobService := mock.NewMockJobService(ctrl)
// 	resultService := mock.NewMockResultService(ctrl)

// 	wp := NewWorkerPoolImpl(jobService, resultService, 1, 5)
// 	wp.logger = log.New(ioutil.Discard, "", 0)
// 	wp.Start()
// 	defer wp.Stop()

// 	j1 := new(domain.Job)
// 	// TODO: Watch out here!
// 	j1.taskName = "test_task"
// 	createdAt1 := time.Now()
// 	j1.CreatedAt = &createdAt1
// 	j1.ID = "job_1"

// 	result := make(chan domain.JobResult, 1)
// 	callback := func(metadata interface{}) (interface{}, error) {
// 		return "some metadata", nil
// 	}
// 	jobItem := domain.JobItem{
// 		Job:         j1,
// 		Result:      result,
// 		TaskFunc:    callback,
// 		TimeoutType: time.Millisecond,
// 	}
// 	futureResult := domain.FutureJobResult{Result: jobItem.Result}
// 	resultService.
// 		EXPECT().
// 		Create(futureResult).
// 		Return(nil).
// 		Times(2)
// 	jobService.
// 		EXPECT().
// 		Exec(context.Background(), jobItem).
// 		Return(nil).
// 		Times(1)

// 	j2 := new(domain.Job)
// 	createdAt2 := time.Now()
// 	j2.CreatedAt = &createdAt2
// 	j2.ID = "job_2"

// 	err := wp.Send(j1)
// 	if err != nil {
// 		t.Errorf("Error sending job to worker: %v", err)
// 	}

// 	// sleep enough for the workerpool to start, but not as much to finish its first
// 	// job which takes at least 1 second
// 	time.Sleep(50 * time.Millisecond)

// 	err = wp.Send(j2)
// 	if err != nil {
// 		t.Errorf("Error sending job to worker: %v", err)
// 	}

// 	// the queue should contain only 1 item, the job item for the 2nd job
// 	queueLength := len(wp.queue)
// 	if queueLength != 1 {
// 		t.Errorf("Expected queue with 1 item, actual length was %#v", queueLength)
// 	}
// 	select {
// 	case i, ok := <-wp.queue:
// 		if !ok {
// 			t.Errorf("Unexpectedly closed workerpool queue")
// 		}
// 		if eq := reflect.DeepEqual(i.Job, j2); !eq {
// 			t.Errorf("Expected %#v to be %#v", i.Job, j2)
// 		}
// 	default:
// 		t.Errorf("Expected to find a job item in the queue")
// 	}
// }
