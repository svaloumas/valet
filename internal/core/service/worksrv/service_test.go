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

func TestDispatch(t *testing.T) {
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

	workservice.Dispatch(work)

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did not increase the queue length: got %v want 1", len(workservice.queue))
	}
}

func TestDispatchMergedJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	defer wg.Wait()

	secondJobID := "second_job_id"
	thirdJobID := "third_job_id"
	job := new(domain.Job)
	job.ID = "job_id"
	job.TaskName = "test_task"
	job.NextJobID = secondJobID

	secondJob := new(domain.Job)
	*secondJob = *job
	secondJob.ID = secondJobID
	secondJob.NextJobID = thirdJobID

	thirdJob := new(domain.Job)
	*thirdJob = *secondJob
	thirdJob.ID = thirdJobID
	thirdJob.NextJobID = ""

	job.Next = secondJob
	job.Next.Next = thirdJob

	// Use a capacity of 3 to send the results in the buffered chan before testing.
	work := work.Work{Job: job, Result: make(chan domain.JobResult, 3)}

	result := domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}
	secondResult := domain.JobResult{
		JobID:    secondJob.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}
	thirdResult := domain.JobResult{
		JobID:    thirdJob.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	work.Result <- result
	work.Result <- secondResult
	work.Result <- thirdResult

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
	wg.Add(1)
	storage.
		EXPECT().
		CreateJobResult(&secondResult).
		DoAndReturn(func(result *domain.JobResult) error {
			wg.Done()
			return nil
		})
	wg.Add(1)
	storage.
		EXPECT().
		CreateJobResult(&thirdResult).
		DoAndReturn(func(result *domain.JobResult) error {
			wg.Done()
			return nil
		})

	logger := &logrus.Logger{Out: ioutil.Discard}
	workservice := New(storage, taskrepo, freezed, time.Second, 0, 1, logger)
	workservice.Start()
	defer workservice.Stop()

	workservice.Dispatch(work)

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did not increase the queue length: got %v want 1", len(workservice.queue))
	}
}

func TestDispatchNoResultCreation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.ID = "job_id"
	j.TaskName = "test_task"

	resultChan := make(chan domain.JobResult, 1)
	work := work.Work{Job: j, Result: resultChan}
	// Channel closed due to some Exec error
	close(resultChan)

	freezed := mock.NewMockTime(ctrl)
	taskrepo := &taskrepo.TaskRepository{}
	storage := mock.NewMockStorage(ctrl)

	logger := &logrus.Logger{Out: ioutil.Discard}
	workservice := New(storage, taskrepo, freezed, time.Second, 0, 1, logger)
	workservice.Start()
	defer workservice.Stop()

	workservice.Dispatch(work)

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did not increase the queue length: got %v want 1", len(workservice.queue))
	}
}

func TestDispatchBacklogLimit(t *testing.T) {
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

	workservice.Dispatch(w)

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
		workservice.Dispatch(workDeemedToBlock)
	}()

	if len(workservice.queue) != 1 {
		t.Errorf("work service send did changed the queue length instead of blocking: got %v want 1", len(workservice.queue))
	}
}
func TestExecJobWorkCompletedJob(t *testing.T) {
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

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithNoError := work.Work{
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
		result := <-workWithNoError.Result
		actualResultChan <- result
		close(actualResultChan)
	}()
	err := service.Exec(context.Background(), workWithNoError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithNoError); !eq {
		t.Errorf("service exec produced wrong job result: got %#v want %#v", actualResult, expectedResultWithNoError)
	}

}

func TestExecJobWorkFailedJob(t *testing.T) {
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

	var taskFuncReturnsErr = func(...interface{}) (interface{}, error) {
		return nil, errors.New(failureReason)
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithError := work.Work{
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
		result := <-workWithError.Result
		actualResultChan <- result
		close(actualResultChan)
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

func TestExecJobWorkPanicJob(t *testing.T) {
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

	var taskFuncReturnsErr = func(...interface{}) (interface{}, error) {
		panic(panicMessage)
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithError := work.Work{
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
		result := <-workWithError.Result
		actualResultChan <- result
		close(actualResultChan)
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

func TestExecJobWorkJobUpdateErrorCases(t *testing.T) {
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

	storageErr := errors.New("some storage error")

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

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	w1 := work.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	w2 := work.Work{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	tests := []struct {
		name string
		work work.Work
		err  error
	}{
		{
			"storage error",
			w1,
			storageErr,
		},
		{
			"storage error",
			w2,
			storageErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				<-tt.work.Result
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

func TestExecJobWorkJobTimeoutExceeded(t *testing.T) {
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

	var taskFuncReturnsErr = func(...interface{}) (interface{}, error) {
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
		close(actualResultChan)
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

func TestExecPipelineWorkCompletedJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(9)

	createdAt := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	thirdJob := &domain.Job{
		ID:         "third_job_id",
		PipelineID: p.ID,
		Name:       "third_job",
		TaskName:   "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	secondJob := &domain.Job{
		ID:         "second_job_id",
		PipelineID: p.ID,
		Name:       "second_job",
		TaskName:   "test_task",
		NextJobID:  thirdJob.ID,
		Next:       thirdJob,
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	job := &domain.Job{
		ID:         "job_id",
		PipelineID: p.ID,
		Name:       "job_name",
		TaskName:   "test_task",
		NextJobID:  secondJob.ID,
		Next:       secondJob,
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

	startedSecondJob := &domain.Job{}
	*startedSecondJob = *secondJob
	startedSecondJob.Status = domain.InProgress
	startedSecondJob.StartedAt = &startedAt

	startedThirdJob := &domain.Job{}
	*startedThirdJob = *thirdJob
	startedThirdJob.Status = domain.InProgress
	startedThirdJob.StartedAt = &startedAt

	startedPipeline := &domain.Pipeline{}
	*startedPipeline = *p
	startedPipeline.Status = domain.InProgress
	startedPipeline.StartedAt = &startedAt

	completedAt := freezed.Now()

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	completedSecondJob := &domain.Job{}
	*completedSecondJob = *startedSecondJob
	completedSecondJob.Status = domain.Completed
	completedSecondJob.CompletedAt = &completedAt

	completedThirdJob := &domain.Job{}
	*completedThirdJob = *startedThirdJob
	completedThirdJob.Status = domain.Completed
	completedThirdJob.CompletedAt = &completedAt

	completedPipeline := &domain.Pipeline{}
	*completedPipeline = *startedPipeline
	completedPipeline.Status = domain.Completed
	completedPipeline.CompletedAt = &completedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedSecondJob.ID, startedSecondJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedSecondJob.ID, completedSecondJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedThirdJob.ID, startedThirdJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedThirdJob.ID, completedThirdJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(completedPipeline.ID, completedPipeline).
		Return(nil).
		Times(1)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithNoError := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	jobResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}
	secondResultWithNoError := domain.JobResult{
		JobID:    secondJob.ID,
		Metadata: "test_metadata",
		Error:    "",
	}
	thirdResultWithNoError := domain.JobResult{
		JobID:    thirdJob.ID,
		Metadata: "test_metadata",
		Error:    "",
	}
	expectedResults := []domain.JobResult{
		jobResultWithNoError, secondResultWithNoError, thirdResultWithNoError,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		for result := range workWithNoError.Result {
			actualResultChan <- result
		}
		close(actualResultChan)
	}()
	err := service.Exec(context.Background(), workWithNoError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	for actualResult := range actualResultChan {
		found := false
		for _, excpected := range expectedResults {

			if eq := reflect.DeepEqual(actualResult, excpected); eq {
				found = true
			}
		}
		if !found {
			t.Errorf("service exec produced wrong job result: got %#v want %#v", actualResult, expectedResults)
		}
	}
}

func TestExecPipelineWorkFailedJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(8)

	createdAt := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	thirdJob := &domain.Job{
		ID:         "third_job_id",
		PipelineID: p.ID,
		Name:       "third_job",
		TaskName:   "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	secondJob := &domain.Job{
		ID:         "second_job_id",
		PipelineID: p.ID,
		Name:       "second_job",
		TaskName:   "test_task_error",
		NextJobID:  thirdJob.ID,
		Next:       thirdJob,
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	job := &domain.Job{
		ID:         "job_id",
		PipelineID: p.ID,
		Name:       "job_name",
		TaskName:   "test_task",
		NextJobID:  secondJob.ID,
		Next:       secondJob,
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

	startedSecondJob := &domain.Job{}
	*startedSecondJob = *secondJob
	startedSecondJob.Status = domain.InProgress
	startedSecondJob.StartedAt = &startedAt

	startedPipeline := &domain.Pipeline{}
	*startedPipeline = *p
	startedPipeline.Status = domain.InProgress
	startedPipeline.StartedAt = &startedAt

	completedAt := freezed.Now()
	failedAt := freezed.Now()

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	failureReason := "some failure reason"
	failedSecondJob := &domain.Job{}
	*failedSecondJob = *startedSecondJob
	failedSecondJob.FailureReason = failureReason
	failedSecondJob.Status = domain.Failed
	failedSecondJob.CompletedAt = &failedAt

	failedPipeline := &domain.Pipeline{}
	*failedPipeline = *startedPipeline
	failedPipeline.Status = domain.Failed
	failedPipeline.CompletedAt = &failedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedSecondJob.ID, startedSecondJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(failedSecondJob.ID, failedSecondJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdatePipeline(failedPipeline.ID, failedPipeline).
		Return(nil).
		Times(1)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	var taskFuncReturnsErr = func(...interface{}) (interface{}, error) {
		return nil, errors.New(failureReason)
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	taskrepo.Register("test_task_error", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithNoError := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	jobResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}
	secondResultWithError := domain.JobResult{
		JobID:    failedSecondJob.ID,
		Metadata: nil,
		Error:    failureReason,
	}
	expectedResults := []domain.JobResult{
		jobResultWithNoError, secondResultWithError,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		for result := range workWithNoError.Result {
			actualResultChan <- result
		}
		close(actualResultChan)
	}()
	err := service.Exec(context.Background(), workWithNoError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	for actualResult := range actualResultChan {
		found := false
		for _, excpected := range expectedResults {

			if eq := reflect.DeepEqual(actualResult, excpected); eq {
				found = true
			}
		}
		if !found {
			t.Errorf("service exec produced wrong job result: got %#v want %#v", actualResult, expectedResults)
		}
	}
}

func TestExecPipelineWorkPanicJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(8)

	createdAt := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	thirdJob := &domain.Job{
		ID:         "third_job_id",
		PipelineID: p.ID,
		Name:       "third_job",
		TaskName:   "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	secondJob := &domain.Job{
		ID:         "second_job_id",
		PipelineID: p.ID,
		Name:       "second_job",
		TaskName:   "test_task_panic",
		NextJobID:  thirdJob.ID,
		Next:       thirdJob,
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	job := &domain.Job{
		ID:         "job_id",
		PipelineID: p.ID,
		Name:       "job_name",
		TaskName:   "test_task",
		NextJobID:  secondJob.ID,
		Next:       secondJob,
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

	startedSecondJob := &domain.Job{}
	*startedSecondJob = *secondJob
	startedSecondJob.Status = domain.InProgress
	startedSecondJob.StartedAt = &startedAt

	startedPipeline := &domain.Pipeline{}
	*startedPipeline = *p
	startedPipeline.Status = domain.InProgress
	startedPipeline.StartedAt = &startedAt

	completedAt := freezed.Now()
	failedAt := freezed.Now()

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	panicMessage := "some panic error"
	failedSecondJob := &domain.Job{}
	*failedSecondJob = *startedSecondJob
	failedSecondJob.FailureReason = panicMessage
	failedSecondJob.Status = domain.Failed
	failedSecondJob.CompletedAt = &failedAt

	failedPipeline := &domain.Pipeline{}
	*failedPipeline = *startedPipeline
	failedPipeline.Status = domain.Failed
	failedPipeline.CompletedAt = &failedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedSecondJob.ID, startedSecondJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(failedSecondJob.ID, failedSecondJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdatePipeline(failedPipeline.ID, failedPipeline).
		Return(nil).
		Times(1)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	var taskFuncReturnsErr = func(...interface{}) (interface{}, error) {
		panic(panicMessage)
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	taskrepo.Register("test_task_panic", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithNoError := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	jobResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}
	secondResultWithPanic := domain.JobResult{
		JobID:    failedSecondJob.ID,
		Metadata: nil,
		Error:    panicMessage,
	}
	expectedResults := []domain.JobResult{
		jobResultWithNoError, secondResultWithPanic,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		for result := range workWithNoError.Result {
			actualResultChan <- result
		}
		close(actualResultChan)
	}()
	err := service.Exec(context.Background(), workWithNoError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	for actualResult := range actualResultChan {
		found := false
		for _, excpected := range expectedResults {

			if eq := reflect.DeepEqual(actualResult, excpected); eq {
				found = true
			}
		}
		if !found {
			t.Errorf("service exec produced wrong job result: got %#v want %#v", actualResult, expectedResults)
		}
	}
}

func TestExecPipelineWorkTimeoutExceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(8)

	createdAt := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	thirdJob := &domain.Job{
		ID:         "third_job_id",
		PipelineID: p.ID,
		Name:       "third_job",
		TaskName:   "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	secondJob := &domain.Job{
		ID:         "second_job_id",
		PipelineID: p.ID,
		Name:       "second_job",
		TaskName:   "test_task_panic",
		Timeout:    20,
		NextJobID:  thirdJob.ID,
		Next:       thirdJob,
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	job := &domain.Job{
		ID:         "job_id",
		PipelineID: p.ID,
		Name:       "job_name",
		TaskName:   "test_task",
		NextJobID:  secondJob.ID,
		Next:       secondJob,
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

	startedSecondJob := &domain.Job{}
	*startedSecondJob = *secondJob
	startedSecondJob.Status = domain.InProgress
	startedSecondJob.StartedAt = &startedAt

	startedPipeline := &domain.Pipeline{}
	*startedPipeline = *p
	startedPipeline.Status = domain.InProgress
	startedPipeline.StartedAt = &startedAt

	completedAt := freezed.Now()
	failedAt := freezed.Now()

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	timeoutMessage := "context deadline exceeded"
	failedSecondJob := &domain.Job{}
	*failedSecondJob = *startedSecondJob
	failedSecondJob.FailureReason = timeoutMessage
	failedSecondJob.Status = domain.Failed
	failedSecondJob.CompletedAt = &failedAt

	failedPipeline := &domain.Pipeline{}
	*failedPipeline = *startedPipeline
	failedPipeline.Status = domain.Failed
	failedPipeline.CompletedAt = &failedAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdateJob(startedSecondJob.ID, startedSecondJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(failedSecondJob.ID, failedSecondJob).
		Return(nil).
		Times(1)

	storage.
		EXPECT().
		UpdatePipeline(failedPipeline.ID, failedPipeline).
		Return(nil).
		Times(1)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	var taskFuncReturnsErr = func(...interface{}) (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return "some metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	taskrepo.Register("test_task_panic", taskFuncReturnsErr)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	workWithNoError := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	jobResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}
	secondResultWithPanic := domain.JobResult{
		JobID:    failedSecondJob.ID,
		Metadata: nil,
		Error:    timeoutMessage,
	}
	expectedResults := []domain.JobResult{
		jobResultWithNoError, secondResultWithPanic,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		for result := range workWithNoError.Result {
			actualResultChan <- result
		}
		close(actualResultChan)
	}()
	err := service.Exec(context.Background(), workWithNoError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	for actualResult := range actualResultChan {
		found := false
		for _, excpected := range expectedResults {

			if eq := reflect.DeepEqual(actualResult, excpected); eq {
				found = true
			}
		}
		if !found {
			t.Errorf("service exec produced wrong job result: got %#v want %#v", actualResult, expectedResults)
		}
	}
}

func TestExecPipelineWorkJobAndPipelineUpdateErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(7)

	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	job := &domain.Job{
		ID:         "job_id",
		PipelineID: p.ID,
		Name:       "job_name",
		TaskName:   "test_task",
		NextJobID:  "",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedJob := &domain.Job{}
	*startedJob = *job
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	startedPipeline := &domain.Pipeline{}
	*startedPipeline = *p
	startedPipeline.Status = domain.InProgress
	startedPipeline.StartedAt = &startedAt

	storageErr := errors.New("some storage error")

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(nil, storageErr).
		Times(1)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(3)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(storageErr).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(2)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(storageErr).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(nil).
		Times(1)

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(storageErr).
		Times(1)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	w1 := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	w2 := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	w3 := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}
	w4 := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	tests := []struct {
		name string
		work work.Work
		err  error
	}{
		{
			"get pipeline storage error",
			w1,
			storageErr,
		},
		{
			"update started job storage error",
			w2,
			storageErr,
		},
		{
			"update started pipeline storage error",
			w3,
			storageErr,
		},
		{
			"update completed job storage error",
			w4,
			storageErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				<-tt.work.Result
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

func TestExecPipelineWorkCompletedPipelineUpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	job := &domain.Job{
		ID:         "job_id",
		PipelineID: p.ID,
		Name:       "job_name",
		TaskName:   "test_task",
		NextJobID:  "",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedJob := &domain.Job{}
	*startedJob = *job
	startedJob.Status = domain.InProgress
	startedJob.StartedAt = &startedAt

	startedPipeline := &domain.Pipeline{}
	*startedPipeline = *p
	startedPipeline.Status = domain.InProgress
	startedPipeline.StartedAt = &startedAt

	completedJob := &domain.Job{}
	*completedJob = *startedJob
	completedJob.Status = domain.Completed
	completedJob.CompletedAt = &completedAt

	completedPipeline := &domain.Pipeline{}
	*completedPipeline = *startedPipeline
	completedPipeline.Status = domain.Completed
	completedPipeline.CompletedAt = &completedAt

	storageErr := errors.New("some storage error")

	storage := mock.NewMockStorage(ctrl)

	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(startedJob.ID, startedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(startedPipeline.ID, startedPipeline).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(completedJob.ID, completedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(completedPipeline.ID, completedPipeline).
		Return(storageErr).
		Times(1)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "test_metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}
	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)

	w := work.Work{
		Type:        WorkTypePipeline,
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TimeoutUnit: time.Millisecond,
	}

	tests := []struct {
		name string
		work work.Work
		err  error
	}{
		{
			"update completed pipeline storage error",
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
					close(actualResultChan)
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

func TestWork(t *testing.T) {
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
		ID:       "job_id",
		Name:     "job_name",
		TaskName: "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobToUsePreviousResults := &domain.Job{}
	*jobToUsePreviousResults = *job
	jobToUsePreviousResults.UsePreviousResults = true

	storage := mock.NewMockStorage(ctrl)

	var taskFunc = func(...interface{}) (interface{}, error) {
		return "task metadata", nil
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	logger := &logrus.Logger{Out: ioutil.Discard}

	previousResults := "previous results metadata"

	jobResult := domain.JobResult{
		JobID:    job.ID,
		Metadata: "task metadata",
		Error:    "",
	}

	service := New(storage, taskrepo, freezed, time.Second, 1, 1, logger)
	tests := []struct {
		name            string
		job             *domain.Job
		expected        domain.JobResult
		previousResults interface{}
	}{
		{
			"work without previous results",
			job,
			jobResult,
			nil,
		},
		{
			"work with use of previous results",
			jobToUsePreviousResults,
			jobResult,
			previousResults,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualResultChan := make(chan domain.JobResult, 1)

			service.work(tt.job, actualResultChan, tt.previousResults)

			result := <-actualResultChan

			if eq := reflect.DeepEqual(result, tt.expected); !eq {
				t.Errorf("service work returned wront result job ID: got %#v want %#v", result, tt.expected)
			}
		})
	}
}
