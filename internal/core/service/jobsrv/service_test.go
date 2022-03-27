package jobsrv

import (
	"errors"
	"reflect"
	"testing"
	"time"
	"valet/internal/core/domain"
	"valet/internal/core/domain/task"
	"valet/mock"
	"valet/pkg/apperrors"

	"github.com/golang/mock/gomock"
)

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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "dummytask",
		Description: "some description",
		Metadata:    &task.DummyMetadata{},
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	jobRepositoryErr := errors.New("some job repository error")
	jobQueueErr := &apperrors.FullQueueErr{}
	jobTaskTypeErr := &apperrors.ResourceValidationErr{Message: "wrongtask is not a valid task type - valid task types: [dummytask]"}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expectedJob.ID, nil).
		Times(4)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expectedJob).
		Return(jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(expectedJob).
		Return(true).
		Times(1)
	jobQueue.
		EXPECT().
		Push(expectedJob).
		Return(false).
		Times(1)

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		name     string
		taskType string
		err      error
	}{
		{
			"",
			"dummytask",
			jobValidateErr,
		},
		{
			"job_name",
			"dummytask",
			jobRepositoryErr,
		},
		{
			"job_name",
			"dummytask",
			jobQueueErr,
		},
		{
			"job_name",
			"wrongtask",
			jobTaskTypeErr,
		},
		{
			"job_name",
			"dummytask",
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		_, err := service.Create(tt.name, tt.taskType, expectedJob.Description, expectedJob.Metadata)
		if err == nil {
			t.Error("service created expected error, returned nil instead")
		}
		if err.Error() != tt.err.Error() {
			t.Errorf("service create returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
		}
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "dummytask",
		Description: "some description",
		Metadata:    "some metadata",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(expectedJob.ID, nil).
		Times(1)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Create(expectedJob).
		Return(nil).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(expectedJob).
		Return(true).
		Times(1)

	service := New(jobRepository, jobQueue, uuidGen, freezed)
	j, err := service.Create(expectedJob.Name, expectedJob.TaskType, expectedJob.Description, expectedJob.Metadata)
	if err != nil {
		t.Errorf("service create returned unexpected error: %#v", err)
	}
	if eq := reflect.DeepEqual(j, expectedJob); !eq {
		t.Errorf("service create returned wrong job, got %#v want %#v", j, expectedJob)
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "dummytask",
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
		Get(expectedJob.ID).
		Return(expectedJob, nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		id  string
		err error
	}{
		{
			expectedJob.ID,
			nil,
		},
		{
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		j, err := service.Get(tt.id)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service get returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		} else {
			if eq := reflect.DeepEqual(j, expectedJob); !eq {
				t.Errorf("service get returned wrong job: got %#v want %#v", j, expectedJob)
			}
		}
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "dummytask",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	updatedJob := &domain.Job{}
	*updatedJob = *expectedJob
	updatedJob.Name = "updated job_name"
	updatedJob.Description = "updated description"

	invalidID := "invalid_id"

	jobRepositoryErr := errors.New("some job repository error")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		Get(expectedJob.ID).
		Return(expectedJob, nil).
		Times(2)
	jobRepository.
		EXPECT().
		Update(expectedJob.ID, updatedJob).
		Return(jobRepositoryErr).
		Times(1)
	jobRepository.
		EXPECT().
		Update(expectedJob.ID, updatedJob).
		Return(nil).
		Times(1)
	jobRepository.
		EXPECT().
		Get(invalidID).
		Return(nil, jobRepositoryErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		id  string
		err error
	}{
		{
			expectedJob.ID,
			jobRepositoryErr,
		},
		{
			expectedJob.ID,
			nil,
		},
		{
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		err := service.Update(tt.id, updatedJob.Name, updatedJob.Description)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service update returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		}
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
		TaskType:    "dummytask",
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

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	tests := []struct {
		id  string
		err error
	}{
		{
			expectedJob.ID,
			nil,
		},
		{
			invalidID,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {
		err := service.Delete(tt.id)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service delete returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		}
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

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	var testTaskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}

	jobItemWithNoError := domain.JobItem{
		Job:      expectedJob,
		Result:   make(chan domain.JobResult, 1),
		TaskFunc: testTaskFunc,
	}

	expectedResultWithNoError := domain.JobResult{
		JobID:    expectedJob.ID,
		Metadata: "test_metadata",
		Error:    "",
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-jobItemWithNoError.Result
		actualResultChan <- result
	}()
	err := service.Exec(jobItemWithNoError)
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

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	var testTaskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		return nil, errors.New(failureReason)
	}

	jobItemWithError := domain.JobItem{
		Job:      expectedJob,
		Result:   make(chan domain.JobResult, 1),
		TaskFunc: testTaskFuncReturnsErr,
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
	err := service.Exec(jobItemWithError)
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

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	var testTaskFuncReturnsErr = func(metadata interface{}) (interface{}, error) {
		panic(panicMessage)
	}

	jobItemWithError := domain.JobItem{
		Job:      expectedJob,
		Result:   make(chan domain.JobResult, 1),
		TaskFunc: testTaskFuncReturnsErr,
	}
	expectedResultWithError := domain.JobResult{
		JobID:    expectedJob.ID,
		Metadata: nil,
		Error:    panicMessage,
	}

	actualResultChan := make(chan domain.JobResult, 1)
	go func() {
		result := <-jobItemWithError.Result
		actualResultChan <- result
	}()
	err := service.Exec(jobItemWithError)
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "dummytask",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()

	startedJob := &domain.Job{}
	*startedJob = *expectedJob
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

	service := New(jobRepository, jobQueue, uuidGen, freezed)

	var testTaskFunc = func(metadata interface{}) (interface{}, error) {
		return "test_metadata", nil
	}

	jobItem := domain.JobItem{
		Job:      expectedJob,
		Result:   make(chan domain.JobResult, 1),
		TaskFunc: testTaskFunc,
	}

	tests := []struct {
		item domain.JobItem
		err  error
	}{
		{
			jobItem,
			jobRepositoryErr,
		},
		{
			jobItem,
			jobRepositoryErr,
		},
	}

	for _, tt := range tests {

		actualResultChan := make(chan domain.JobResult, 1)
		go func() {
			select {
			case result := <-jobItem.Result:
				actualResultChan <- result
			default:
			}
		}()
		err := service.Exec(tt.item)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.Errorf("service exec returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		}
	}
}
