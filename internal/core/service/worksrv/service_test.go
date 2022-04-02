package worksrv

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/mock"
)

func TestBacklogLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.TaskName = "test_task"

	resultChan1 := make(chan domain.JobResult, 1)
	work := domain.NewWork(j, resultChan1, time.Millisecond)

	jDeemedToFail := new(domain.Job)
	jDeemedToFail.TaskName = "test_task"

	resultChan2 := make(chan domain.JobResult, 1)
	workDeemedToFail := domain.NewWork(jDeemedToFail, resultChan2, time.Millisecond)

	freezed := mock.NewMockTime(ctrl)
	taskrepo := taskrepo.NewTaskRepository()
	jobRepository := mock.NewMockJobRepository(ctrl)
	resultRepository := mock.NewMockResultRepository(ctrl)

	workservice := New(jobRepository, resultRepository, taskrepo, freezed, 0, 1)
	workservice.Log = log.New(ioutil.Discard, "", 0)
	workservice.Start()
	defer workservice.Stop()

	err := workservice.Send(work)
	if err != nil {
		t.Errorf("Error sending job to workerpool: %v", err)
	}

	err = workservice.Send(workDeemedToFail)
	if err == nil {
		t.Fatal("Expected error, due to backlog being full")
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
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Metadata:    "some metadata",
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

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Update(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Update(completedJob.ID, completedJob).
		Return(nil).
		Times(1)
	resultRepository := mock.NewMockResultRepository(ctrl)

	var taskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)
	service := New(jobRepository, resultRepository, taskrepo, freezed, 1, 1)

	workItemWithNoError := domain.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	expectedResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workItemWithNoError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workItemWithNoError)
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

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Update(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Update(failedJob.ID, failedJob).
		Return(nil).
		Times(1)
	resultRepository := mock.NewMockResultRepository(ctrl)

	var taskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		return nil, errors.New(failureReason)
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	service := New(jobRepository, resultRepository, taskrepo, freezed, 1, 1)

	workItemWithError := domain.Work{
		Job:         expectedJob,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    expectedJob.ID,
		Metadata: nil,
		Error:    failureReason,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workItemWithError)
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

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Update(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Update(failedJob.ID, failedJob).
		Return(nil).
		Times(1)
	resultRepository := mock.NewMockResultRepository(ctrl)

	var taskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		panic(panicMessage)
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	service := New(jobRepository, resultRepository, taskrepo, freezed, 1, 1)

	workItemWithError := domain.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    job.ID,
		Metadata: nil,
		Error:    panicMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workItemWithError)
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

	jobRepositoryErr := errors.New("some repository error")

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Update(startedJob.ID, startedJob).
		Return(jobRepositoryErr).
		Times(1)
	jobRepository.
		EXPECT().
		Update(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Update(completedJob.ID, completedJob).
		Return(jobRepositoryErr).
		Times(1)
	resultRepository := mock.NewMockResultRepository(ctrl)

	var taskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)
	service := New(jobRepository, resultRepository, taskrepo, freezed, 1, 1)

	w := domain.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	tests := []struct {
		name string
		work domain.Work
		err  error
	}{
		{
			"job repository error",
			w,
			jobRepositoryErr,
		},
		{
			"job repository error",
			w,
			jobRepositoryErr,
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

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Update(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Update(failedJob.ID, failedJob).
		Return(nil).
		Times(1)
	resultRepository := mock.NewMockResultRepository(ctrl)

	var taskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 50)
		return "some_metadata", nil
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	service := New(jobRepository, resultRepository, taskrepo, freezed, 1, 1)

	workItemWithError := domain.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    job.ID,
		Metadata: nil,
		Error:    timeoutErrorMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-workItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), workItemWithError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithError); !eq {
		t.Errorf("service exec produced wrong job result: go %#v want %#v", actualResult, expectedResultWithError)
	}
}
