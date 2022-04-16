package worksrv

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/internal/core/service/worksrv/work"
	"github.com/svaloumas/valet/mock"
)

func TestSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	defer wg.Wait()

	j := new(domain.Job)
	j.ID = "job_id"
	j.TaskName = "test_task"

	work := work.Work{Job: j, Result: make(chan domain.JobResult, 1)}

	result := domain.JobResult{
		JobID:    j.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	work.Result <- result

	freezed := mock.NewMockTime(ctrl)
	taskrepo := &taskrepo.TaskRepository{}
	storage := mock.NewMockStorage(ctrl)
	wg.Add(1)
	storage.
		EXPECT().
		CreateJobResult(&result).
		DoAndReturn(func(result *domain.JobResult) error {
			wg.Done()
			return nil
		})

	logger := &logrus.Logger{Out: ioutil.Discard}
	workservice := New(storage, taskrepo, freezed, time.Second, 0, 1, logger)
	workservice.Start()
	defer workservice.Stop()

	workservice.Send(work)

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did not increase the queue length: got %v want 1", len(workservice.queue))
	}
}

func TestSendBacklogLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.TaskName = "test_task"

	w := work.Work{Job: j}

	jDeemedToBlock := new(domain.Job)
	jDeemedToBlock.TaskName = "test_task"

	workDeemedToBlock := work.Work{Job: jDeemedToBlock}

	freezed := mock.NewMockTime(ctrl)
	taskrepo := taskrepo.New()
	storage := mock.NewMockStorage(ctrl)

	logger := &logrus.Logger{Out: ioutil.Discard}
	workservice := New(storage, taskrepo, freezed, time.Second, 0, 1, logger)
	workservice.Start()
	defer workservice.Stop()

	workservice.Send(w)

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did not increase the queue length: got %v want 1", len(workservice.queue))
	}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				if err := fmt.Errorf("%s", p); err.Error() != "send on closed channel" {
					t.Error("work service send did not panic while sending on closed channel")
				}
			}
		}()
		// This should block forever.
		workservice.Send(workDeemedToBlock)
	}()

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did changed the queue length instead of blocking: got %v want 1", len(workservice.queue))
	}
}
func TestExecCompletedJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:       "auuid4",
		Name:     "job_name",
		TaskName: "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()

	startedJob := &domain.Job{}
	*startedJob = *job
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	completedAt := freezed.Now()

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(nil).
		Times(1)

	var taskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithNoError := work.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    taskFunc,
		TimeoutUnit: time.Millisecond,
	}

	expectedResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workWithNoError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workWithNoError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithNoError); !eq {
		t.Errorf("service exec produced wrong job result: go %#v want %#v", actualResult, expectedResultWithNoError)
	}

}

func TestExecFailedJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()

	startedJob := &domain.Job{}
	*startedJob = *expectedJob
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	failedAt := freezed.Now()
	failureReason := "some task func error"

	failedJob := &domain.Job{}
	*failedJob = *startedJob
	failedJob.Status = domain.Failed
	failedJob.FailureReason = failureReason
	failedJob.CompletedAt = &failedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(failedJob.ID, failedJob).
		Return(nil).
		Times(1)

	var taskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		return nil, errors.New(failureReason)
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithError := work.Work{
		Job:         expectedJob,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    taskFuncReturnsErr,
		TimeoutUnit: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    expectedJob.ID,
		Metadata: nil,
		Error:    failureReason,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workWithError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithError); !eq {
		t.Errorf("service exec produced wrong job result: go %#v want %#v", actualResult, expectedResultWithError)
	}
}

func TestExecPanicJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()

	startedJob := &domain.Job{}
	*startedJob = *job
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	failedAt := freezed.Now()
	panicMessage := "some panic error"

	failedJob := &domain.Job{}
	*failedJob = *startedJob
	failedJob.Status = domain.Failed
	failedJob.FailureReason = panicMessage
	failedJob.CompletedAt = &failedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(failedJob.ID, failedJob).
		Return(nil).
		Times(1)

	var taskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		panic(panicMessage)
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithError := work.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    taskFuncReturnsErr,
		TimeoutUnit: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    job.ID,
		Metadata: nil,
		Error:    panicMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workWithError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithError); !eq {
		t.Errorf("service exec produced wrong job result: go %#v want %#v", actualResult, expectedResultWithError)
	}
}

func TestExecJobUpdateErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(6)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()

	startedJob := &domain.Job{}
	*startedJob = *job
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	completedAt := freezed.Now()

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	storageErr := errors.New("some repository error")

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(storageErr).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(storageErr).
		Times(1)

	var taskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	w := work.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    taskFunc,
		TimeoutUnit: time.Millisecond,
	}

	tests := []struct {
		name string
		work work.Work
		err  error
	}{
		{
			"storage error",
			w,
			storageErr,
		},
		{
			"storage error",
			w,
			storageErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualResultChan := make(chan domain.JobResult, 1)
			go func() {
				select {
				case result := <-w.Result:
					actualResultChan <- result
				default:
				}
			}()
			err := service.Exec(context.Background(), tt.work)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service exec returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			}
		})
	}
}

func TestExecJobTimeoutExceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     20,
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()

	startedJob := &domain.Job{}
	*startedJob = *job
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	failedAt := freezed.Now()
	timeoutErrorMessage := "context deadline exceeded"

	failedJob := &domain.Job{}
	*failedJob = *startedJob
	failedJob.Status = domain.Failed
	failedJob.FailureReason = timeoutErrorMessage
	failedJob.CompletedAt = &failedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(failedJob.ID, failedJob).
		Return(nil).
		Times(1)

	var taskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 50)
		return "some_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithError := work.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    taskFuncReturnsErr,
		TimeoutUnit: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    job.ID,
		Metadata: nil,
		Error:    timeoutErrorMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workWithError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithError); !eq {
		t.Errorf("service exec produced wrong job result: go %#v want %#v", actualResult, expectedResultWithError)
	}
}
