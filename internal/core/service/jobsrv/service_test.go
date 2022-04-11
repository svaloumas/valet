package jobsrv

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/internal/core/domain/taskrepo"
	"valet/mock"
	"valet/pkg/apperrors"
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
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	storageErr := errors.New("some storage error")
	jobQueueErr := &apperrors.FullQueueErr{}
	jobTaskNameErr := &apperrors.ResourceValidationErr{Message: "wrongtask is not a valid task name - valid tasks: [test_task]"}
	parseTimeErr := &apperrors.ParseTimeErr{
		Message: "parsing time \"invalid_timestamp_format\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"invalid_timestamp_format\" as \"2006\""}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(5)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(1)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		CreateJob(job).
		Return(storageErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(job).
		Return(nil).
		Times(1)
	jobQueue.
		EXPECT().
		Push(job).
		Return(jobQueueErr).
		Times(1)

	taskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)
	service := New(storage, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name     string
		jobName  string
		taskName string
		runAt    string
		err      error
	}{
		{
			"job validation error",
			"",
			"test_task",
			"",
			jobValidateErr,
		},
		{
			"storage error",
			"job_name",
			"test_task",
			"",
			storageErr,
		},
		{
			"job queue error",
			"job_name",
			"test_task",
			"",
			jobQueueErr,
		},
		{
			"job task type error",
			"job_name",
			"wrongtask",
			"",
			jobTaskNameErr,
		},
		{
			"parse time error",
			"job_name",
			"test_task",
			"invalid_timestamp_format",
			parseTimeErr,
		},
		{
			"uuid generator error",
			"job_name",
			"test_task",
			"",
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(tt.jobName, tt.taskName, job.Description, tt.runAt, job.Timeout, job.TaskParams)
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
		Times(3)

	runAt := "2022-01-02T15:04:05.999999999Z"
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAt)
	createdAt := freezed.Now()
	jobWithoutSchedule := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}

	jobWithSchedule := &domain.Job{}
	*jobWithSchedule = *jobWithoutSchedule
	jobWithSchedule.RunAt = &runAtTime

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(jobWithoutSchedule.ID, nil).
		Times(2)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		CreateJob(jobWithoutSchedule).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		CreateJob(jobWithSchedule).
		Return(nil).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(jobWithoutSchedule).
		Return(nil).
		Times(1)

	taskFunc := func(i interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("test_task", taskFunc)

	service := New(storage, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name        string
		jobName     string
		taskName    string
		description string
		runAt       string
		timeout     int
		taskParams  map[string]interface{}
		expected    *domain.Job
	}{
		{
			"job without schedule",
			jobWithoutSchedule.Name,
			jobWithoutSchedule.TaskName,
			jobWithoutSchedule.Description,
			"",
			jobWithoutSchedule.Timeout,
			jobWithoutSchedule.TaskParams,
			jobWithoutSchedule,
		},
		{
			"job with schedule",
			jobWithSchedule.Name,
			jobWithSchedule.TaskName,
			jobWithSchedule.Description,
			"2022-01-02T15:04:05.999999999Z",
			jobWithSchedule.Timeout,
			jobWithSchedule.TaskParams,
			jobWithSchedule,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j, err := service.Create(tt.jobName, tt.taskName, tt.description, tt.runAt, tt.timeout, tt.taskParams)
			if err != nil {
				t.Errorf("service create returned unexpected error: %#v", err)
			}
			if eq := reflect.DeepEqual(j, tt.expected); !eq {
				t.Errorf("service create returned wrong job, got %#v want %#v", j, tt.expected)
			}
		})
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
		Times(2)

	createdAt := freezed.Now()
	pendingJob := &domain.Job{
		ID:       "pending_job_id",
		Name:     "pending_job_name",
		TaskName: "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	startedAt := freezed.Now()
	completedAt := startedAt.Add(10 * time.Minute)
	failedJob := &domain.Job{
		ID:       "failed_job_id",
		Name:     "failed_job_name",
		TaskName: "test_task",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Description:   "some description",
		FailureReason: "some failure reason",
		Status:        domain.Failed,
		CreatedAt:     &createdAt,
		StartedAt:     &startedAt,
		CompletedAt:   &completedAt,
	}
	failedJobWithDuration := &domain.Job{}
	*failedJobWithDuration = *failedJob
	failedJobWithDuration.SetDuration()

	storageErr := errors.New("some storage error")
	invalidID := "invalid_id"
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetJob(pendingJob.ID).
		Return(pendingJob, nil).
		Times(1)
	storage.
		EXPECT().
		GetJob(failedJob.ID).
		Return(failedJob, nil).
		Times(1)
	storage.
		EXPECT().
		GetJob(invalidID).
		Return(nil, storageErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	taskrepo := taskrepo.NewTaskRepository()
	service := New(storage, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		job  *domain.Job
		err  error
	}{
		{
			"ok",
			pendingJob.ID,
			pendingJob,
			nil,
		},
		{
			"ok with duration",
			failedJob.ID,
			failedJobWithDuration,
			nil,
		},
		{
			"storage error",
			invalidID,
			nil,
			storageErr,
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
				if eq := reflect.DeepEqual(j, tt.job); !eq {
					t.Errorf("service get returned wrong job: got %#v want %#v", j, tt.job)
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

	updatedJob := &domain.Job{}
	*updatedJob = *job
	updatedJob.Name = "updated job_name"
	updatedJob.Description = "updated description"

	invalidID := "invalid_id"

	storageErr := errors.New("some storage error")
	jobNotFoundErr := errors.New("job not found")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetJob(job.ID).
		Return(job, nil).
		Times(2)
	storage.
		EXPECT().
		UpdateJob(job.ID, updatedJob).
		Return(storageErr).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(job.ID, updatedJob).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		GetJob(invalidID).
		Return(nil, jobNotFoundErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	taskrepo := taskrepo.NewTaskRepository()
	service := New(storage, jobQueue, taskrepo, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"storage error",
			job.ID,
			storageErr,
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

	invalidID := "invalid_id"
	notExistingID := "not_existing_id"
	notFoundErr := &apperrors.NotFoundErr{ID: invalidID, ResourceName: "job"}
	storageErr := errors.New("some storage error")
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetJob(expectedJob.ID).
		Return(expectedJob, nil).
		Times(1)
	storage.
		EXPECT().
		GetJob(notExistingID).
		Return(nil, notFoundErr).
		Times(1)
	storage.
		EXPECT().
		DeleteJob(expectedJob.ID).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		GetJob(invalidID).
		Return(expectedJob, nil).
		Times(1)
	storage.
		EXPECT().
		DeleteJob(invalidID).
		Return(storageErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	taskrepo := taskrepo.NewTaskRepository()
	service := New(storage, jobQueue, taskrepo, uuidGen, freezed)

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
			"not found",
			notExistingID,
			notFoundErr,
		},
		{
			"storage error",
			invalidID,
			storageErr,
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
