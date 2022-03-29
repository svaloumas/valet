package jobsrv

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/internal/core/domain/task"
	"valet/mock"
	"valet/pkg/apperrors"
)

var validTasks = map[string]task.TaskFunc{
	"test_task": func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	},
}

func TestCreateErrorCases(t *testing.T) {
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
		TaskType:    "test_task",
		Timeout:     10,
		Description: "some description",
		Metadata:    "some metadata",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	jobRepositoryErr := errors.New("some job repository error")
	jobQueueErr := &apperrors.FullQueueErr{}
	jobTaskTypeErr := &apperrors.ResourceValidationErr{Message: "wrongtask is not a valid task type - valid task types: [test_task]"}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(4)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(job).
		Return(jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(job).
		Return(true).
		Times(1)
	jobQueue.
		EXPECT().
		Push(job).
		Return(false).
		Times(1)

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	tests := []struct {
		name     string
		jobname  string
		taskType string
		err      error
	}{
		{
			"job validation error",
			"",
			"test_task",
			jobValidateErr,
		},
		{
			"job repository error",
			"job_name",
			"test_task",
			jobRepositoryErr,
		},
		{
			"job queue error",
			"job_name",
			"test_task",
			jobQueueErr,
		},
		{
			"job task type error",
			"job_name",
			"wrongtask",
			jobTaskTypeErr,
		},
		{
			"uuid generator error",
			"job_name",
			"test_task",
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(tt.jobname, tt.taskType, job.Description, job.Timeout, job.Metadata)
			if err == nil {
				t.Error("service created expected error, returned nil instead")
			}
			if err.Error() != tt.err.Error() {
				t.Errorf("service create returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)

	createdAt := freezed.Now()
	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "test_task",
		Timeout:     10,
		Description: "some description",
		Metadata:    "some metadata",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expected.ID, nil).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expected).
		Return(nil).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(expected).
		Return(true).
		Times(1)

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)
	j, err := service.Create(expected.Name, expected.TaskType, expected.Description, expected.Timeout, expected.Metadata)
	if err != nil {
		t.Errorf("service create returned unexpected error: %#v", err)
	}
	if eq := reflect.DeepEqual(j, expected); !eq {
		t.Errorf("service create returned wrong job, got %#v want %#v", j, expected)
	}
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "test_task",
		Metadata:    "some metadata",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobRepositoryErr := errors.New("some repository error")
	invalidID := "invalid_id"
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Get(expected.ID).
		Return(expected, nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expected.ID,
			nil,
		},
		{
			"job repository error",
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j, err := service.Get(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service get returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if eq := reflect.DeepEqual(j, expected); !eq {
					t.Errorf("service get returned wrong job: got %#v want %#v", j, expected)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
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
		Metadata:    "some metadata",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	updatedJob := &domain.Job{}
	*updatedJob = *job
	updatedJob.Name = "updated job_name"
	updatedJob.Description = "updated description"

	invalidID := "invalid_id"

	jobRepositoryErr := errors.New("some job repository error")
	jobNotFoundErr := errors.New("job not found")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Get(job.ID).
		Return(job, nil).
		Times(2)
	jobRepository.
		EXPECT().
		Update(job.ID, updatedJob).
		Return(jobRepositoryErr).
		Times(1)
	jobRepository.
		EXPECT().
		Update(job.ID, updatedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobNotFoundErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"job repository error",
			job.ID,
			jobRepositoryErr,
		},
		{
			"ok",
			job.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			jobNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(tt.id, updatedJob.Name, updatedJob.Description)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service update returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "test_task",
		Metadata:    "some metadata",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	invalidID := "invalid_id"
	jobRepositoryErr := errors.New("some repository error")
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Delete(expectedJob.ID).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Delete(invalidID).
		Return(jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			expectedJob.ID,
			nil,
		},
		{
			"job repository error",
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service delete returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			}
		})
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
		TaskType:    "test_task",
		Metadata:    "some metadata",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)

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

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	var testTaskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}

	jobItemWithNoError := domain.JobItem{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    testTaskFunc,
		TimeoutType: time.Millisecond,
	}

	expectedResultWithNoError := domain.JobResult{
		JobID:    job.ID,
		Metadata: "test_metadata",
		Error:    "",
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-jobItemWithNoError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), jobItemWithNoError)
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
		TaskType:    "test_func",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)

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

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	var testTaskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		return nil, errors.New(failureReason)
	}

	jobItemWithError := domain.JobItem{
		Job:         expectedJob,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    testTaskFuncReturnsErr,
		TimeoutType: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    expectedJob.ID,
		Metadata: nil,
		Error:    failureReason,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-jobItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), jobItemWithError)
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
		TaskType:    "test_func",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)

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

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	var testTaskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		panic(panicMessage)
	}

	jobItemWithError := domain.JobItem{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    testTaskFuncReturnsErr,
		TimeoutType: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    job.ID,
		Metadata: nil,
		Error:    panicMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-jobItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), jobItemWithError)
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
		TaskType:    "dummytask",
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

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)

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

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	var testTaskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}

	jobItem := domain.JobItem{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    testTaskFunc,
		TimeoutType: time.Millisecond,
	}

	tests := []struct {
		name string
		item domain.JobItem
		err  error
	}{
		{
			"job repository error",
			jobItem,
			jobRepositoryErr,
		},
		{
			"job repository error",
			jobItem,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualResultChan := make(chan domain.JobResult, 1)
			go func() {
				select {
				case result := <-jobItem.Result:
					actualResultChan <- result
				default:
				}
			}()
			err := service.Exec(context.Background(), tt.item)
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
		TaskType:    "test_func",
		Timeout:     20,
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)

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

	service := New(jobRepository, jobQueue, validTasks, uuidGen, freezed)

	var testTaskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 50)
		return "some_metadata", nil
	}

	jobItemWithError := domain.JobItem{
		Job:         job,
		Result:      make(chan domain.JobResult, 1),
		TaskFunc:    testTaskFuncReturnsErr,
		TimeoutType: time.Millisecond,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    job.ID,
		Metadata: nil,
		Error:    timeoutErrorMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-jobItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(context.Background(), jobItemWithError)
	if err != nil {
		t.Errorf("service exec returned error: got %#v want nil", err.Error())
	}
	actualResult := <-actualResultChan
	if eq := reflect.DeepEqual(actualResult, expectedResultWithError); !eq {
		t.Errorf("service exec produced wrong job result: go %#v want %#v", actualResult, expectedResultWithError)
	}
}
