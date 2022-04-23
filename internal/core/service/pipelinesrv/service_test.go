package pipelinesrv

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
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

	secondJobID := "second_job_id"
	pipelineID := "pipeline_id"
	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		NextJobID:   secondJobID,
		PipelineID:  pipelineID,
		RunAt:       nil,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: false,
		CreatedAt:          &createdAt,
	}
	secondJob := &domain.Job{
		ID:                 secondJobID,
		Name:               "second_job",
		TaskName:           "test_task",
		NextJobID:          "",
		PipelineID:         pipelineID,
		RunAt:              nil,
		Timeout:            10,
		Description:        "some description",
		TaskParams:         nil,
		Status:             domain.Pending,
		UsePreviousResults: true,
		CreatedAt:          &createdAt,
	}

	noNameJob := &domain.Job{}
	*noNameJob = *job
	noNameJob.Name = ""

	p := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Pending,
		Jobs: []*domain.Job{
			job, secondJob,
		},
		CreatedAt: &createdAt,
	}
	uuidGenErr := errors.New("some uuid generator error")
	jobValidateErr := errors.New("name required")
	pipelineValidateErr := errors.New("name required")
	parseTimeErr := &apperrors.ParseTimeErr{
		Message: "parsing time \"invalid_timestamp_format\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"invalid_timestamp_format\" as \"2006\""}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(pipelineID, nil).
		Times(4)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(3)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(secondJobID, nil).
		Times(2)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return("", uuidGenErr).
		Times(2)

	storage := mock.NewMockStorage(ctrl)

	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	service := New(storage, taskrepo, uuidGen, freezed)

	tests := []struct {
		name         string
		pipelineName string
		description  string
		runAt        string
		jobs         []*domain.Job
		err          error
	}{
		{
			"job validation error",
			p.Name,
			p.Description,
			"",
			[]*domain.Job{
				noNameJob,
			},
			jobValidateErr,
		},
		{
			"pipeline validation error",
			"",
			p.Description,
			"",
			p.Jobs,
			pipelineValidateErr,
		},
		{
			"parse time error",
			p.Name,
			p.Description,
			"invalid_timestamp_format",
			p.Jobs,
			parseTimeErr,
		},
		{
			"job uuid generator error",
			p.Name,
			p.Description,
			"",
			p.Jobs,
			uuidGenErr,
		},
		{
			"pipeline uuid generator error",
			p.Name,
			p.Description,
			"",
			p.Jobs,
			uuidGenErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(tt.pipelineName, job.Description, tt.runAt, tt.jobs)
			if err == nil {
				t.Error("service created expected error, returned nil instead")
			}
			if err.Error() != tt.err.Error() {
				t.Errorf("service create returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
			}
		})
	}
}

func TestCreateStorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(4)

	secondJobID := "second_job_id"
	pipelineID := "pipeline_id"
	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		NextJobID:   secondJobID,
		PipelineID:  pipelineID,
		RunAt:       nil,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: false,
		CreatedAt:          &createdAt,
	}
	secondJob := &domain.Job{
		ID:                 secondJobID,
		Name:               "second_job",
		TaskName:           "test_task",
		NextJobID:          "",
		PipelineID:         pipelineID,
		RunAt:              nil,
		Timeout:            10,
		Description:        "some description",
		TaskParams:         nil,
		Status:             domain.Pending,
		UsePreviousResults: true,
		CreatedAt:          &createdAt,
	}

	noNameJob := &domain.Job{}
	*noNameJob = *job
	noNameJob.Name = ""

	p := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Pending,
		Jobs: []*domain.Job{
			job, secondJob,
		},
		CreatedAt: &createdAt,
	}
	storageErr := errors.New("some storage error")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(pipelineID, nil).
		Times(1)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(1)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(secondJobID, nil).
		Times(1)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		CreatePipeline(p).
		Return(storageErr).
		Times(1)

	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	service := New(storage, taskrepo, uuidGen, freezed)

	_, err := service.Create(p.Name, p.Description, "", p.Jobs)
	if err == nil {
		t.Error("service created expected error, returned nil instead")
	}
	if err.Error() != storageErr.Error() {
		t.Errorf("service create returned wrong error: got %#v want %#v", err.Error(), storageErr.Error())
	}
}

func TestCreateWithoutSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(4)

	secondJobID := "second_job_id"
	pipelineID := "pipeline_id"
	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		NextJobID:   secondJobID,
		PipelineID:  pipelineID,
		RunAt:       nil,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: false,
		CreatedAt:          &createdAt,
	}
	secondJob := &domain.Job{
		ID:          secondJobID,
		Name:        "second_job",
		TaskName:    "test_task",
		NextJobID:   "",
		PipelineID:  pipelineID,
		RunAt:       nil,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: true,
		CreatedAt:          &createdAt,
	}

	expectedJobs := []*domain.Job{
		job, secondJob,
	}

	expected := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Pending,
		Jobs:        expectedJobs,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(pipelineID, nil).
		Times(1)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(1)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(secondJobID, nil).
		Times(1)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		CreatePipeline(expected).
		Return(nil).
		Times(1)

	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	service := New(storage, taskrepo, uuidGen, freezed)

	p, err := service.Create(expected.Name, expected.Description, "", expected.Jobs)
	if err != nil {
		t.Errorf("service create returned unexpected error: got %#v want nil", err)
	}

	for i := range expected.Jobs {
		if eq := reflect.DeepEqual(p.Jobs[i], expected.Jobs[i]); !eq {
			t.Errorf("service create returned wrong pipeline job: got %#v want %#v", p.Jobs[i], expected.Jobs[i])
		}
	}

	if eq := reflect.DeepEqual(p, expected); !eq {
		t.Errorf("service create returned wrong pipeline: got %#v want %#v", p, expected)
	}
}

func TestCreateWithSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(4)

	secondJobID := "second_job_id"
	pipelineID := "pipeline_id"
	createdAt := freezed.Now()
	runAt := "2022-01-02T15:04:05.999999999Z"
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAt)
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		NextJobID:   secondJobID,
		PipelineID:  pipelineID,
		RunAt:       &runAtTime,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: false,
		CreatedAt:          &createdAt,
	}
	secondJob := &domain.Job{
		ID:          secondJobID,
		Name:        "second_job",
		TaskName:    "test_task",
		NextJobID:   "",
		PipelineID:  pipelineID,
		RunAt:       nil,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: true,
		CreatedAt:          &createdAt,
	}

	expectedJobs := []*domain.Job{
		job, secondJob,
	}

	expected := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "a_pipeline",
		Description: "some description",
		RunAt:       &runAtTime,
		Status:      domain.Pending,
		Jobs:        expectedJobs,
		CreatedAt:   &createdAt,
	}

	uuidGen := mock.NewMockUUIDGenerator(ctrl)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(pipelineID, nil).
		Times(1)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(job.ID, nil).
		Times(1)
	uuidGen.
		EXPECT().
		GenerateRandomUUIDString().
		Return(secondJobID, nil).
		Times(1)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		CreatePipeline(expected).
		Return(nil).
		Times(1)

	taskFunc := func(...interface{}) (interface{}, error) {
		return "some metadata", errors.New("some task error")
	}
	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", taskFunc)
	service := New(storage, taskrepo, uuidGen, freezed)

	p, err := service.Create(expected.Name, expected.Description, runAt, expected.Jobs)
	if err != nil {
		t.Errorf("service create returned unexpected error: got %#v want nil", err)
	}

	for i := range expected.Jobs {
		if eq := reflect.DeepEqual(p.Jobs[i], expected.Jobs[i]); !eq {
			t.Errorf("service create returned wrong pipeline job: got %#v want %#v", p.Jobs[i], expected.Jobs[i])
		}
	}

	if eq := reflect.DeepEqual(p, expected); !eq {
		t.Errorf("service create returned wrong pipeline: got %#v want %#v", p, expected)
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

	pipelineID := "pipeline_id"
	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := startedAt.Add(1 * time.Minute)
	runAt := "2022-01-02T15:04:05.999999999Z"
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAt)

	expected := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "a_pipeline",
		Description: "some description",
		RunAt:       &runAtTime,
		Status:      domain.Failed,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}

	storageErr := errors.New("some storage error")
	invalidID := "invalid_id"
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(expected.ID).
		Return(expected, nil).
		Times(1)
	storage.
		EXPECT().
		GetPipeline(invalidID).
		Return(nil, storageErr).
		Times(1)

	taskrepo := taskrepo.New()
	service := New(storage, taskrepo, uuidGen, freezed)

	tests := []struct {
		name     string
		id       string
		pipeline *domain.Pipeline
		err      error
	}{
		{
			"ok",
			expected.ID,
			expected,
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
				if eq := reflect.DeepEqual(j, tt.pipeline); !eq {
					t.Errorf("service get returned wrong job: got %#v want %#v", j, tt.pipeline)
				}
			}
		})
	}
}

func TestGetPipelines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)

	pipelineID1 := "pipeline_id"
	pipelineID2 := "another_pipeline_id"

	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := startedAt.Add(1 * time.Minute)
	runAt := "2022-01-02T15:04:05.999999999Z"
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAt)

	failedPipeline := &domain.Pipeline{
		ID:          pipelineID1,
		Name:        "a_pipeline",
		Description: "some description",
		RunAt:       &runAtTime,
		Status:      domain.Failed,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}

	completedPipeline := &domain.Pipeline{
		ID:          pipelineID2,
		Name:        "another_pipeline",
		Description: "some description",
		RunAt:       &runAtTime,
		Status:      domain.Completed,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}

	failedPipelines := []*domain.Pipeline{failedPipeline}
	completedPipelines := []*domain.Pipeline{completedPipeline}
	allPipelines := []*domain.Pipeline{failedPipeline, completedPipeline}

	storageErr := errors.New("some storage error")
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipelines(domain.Undefined).
		Return(allPipelines, nil).
		Times(1)
	storage.
		EXPECT().
		GetPipelines(failedPipeline.Status).
		Return(failedPipelines, nil).
		Times(1)
	storage.
		EXPECT().
		GetPipelines(completedPipeline.Status).
		Return(completedPipelines, nil).
		Times(1)
	storage.
		EXPECT().
		GetPipelines(failedPipeline.Status).
		Return(nil, storageErr).
		Times(1)

	taskrepo := taskrepo.New()
	service := New(storage, taskrepo, uuidGen, freezed)

	tests := []struct {
		name     string
		status   string
		expected []*domain.Pipeline
		err      error
	}{
		{
			"all",
			"",
			allPipelines,
			nil,
		},
		{
			"failed",
			"failed",
			failedPipelines,
			nil,
		},
		{
			"completed",
			"completed",
			completedPipelines,
			nil,
		},
		{
			"failed",
			"failed",
			nil,
			storageErr,
		},
		{
			"invalid status",
			"not_started",
			nil,
			errors.New("invalid status: \"NOT_STARTED\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipelines, err := service.GetPipelines(tt.status)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service get pipelines returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if eq := reflect.DeepEqual(pipelines, tt.expected); !eq {
					t.Errorf("service get pipelines returned wrong jobs: got %#v want %#v", pipelines, tt.expected)
				}
			}
		})
	}
}

func TestGetPipelineJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	secondJobID := "second_job_id"
	pipelineID := "pipeline_id"
	createdAt := freezed.Now()
	completedAt := createdAt.Add(1 * time.Minute)
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		NextJobID:   secondJobID,
		PipelineID:  pipelineID,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: false,
		CreatedAt:          &createdAt,
	}
	job.SetDuration()
	secondJob := &domain.Job{
		ID:          secondJobID,
		Name:        "second_job",
		TaskName:    "test_task",
		NextJobID:   "",
		PipelineID:  pipelineID,
		RunAt:       nil,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: true,
		CreatedAt:          &createdAt,
		StartedAt:          &createdAt,
		CompletedAt:        &completedAt,
	}
	secondJob.SetDuration()

	jobs := []*domain.Job{
		job, secondJob,
	}

	p := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Completed,
		CreatedAt:   &createdAt,
		StartedAt:   &createdAt,
		CompletedAt: &completedAt,
	}

	notExistingID := "not_existing_id"
	notFoundErr := &apperrors.NotFoundErr{ID: notExistingID, ResourceName: "pipeline"}
	storageErr := errors.New("some storage error")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(pipelineID).
		Return(p, nil).
		Times(2)
	storage.
		EXPECT().
		GetJobsByPipelineID(pipelineID).
		Return(jobs, nil).
		Times(1)
	storage.
		EXPECT().
		GetJobsByPipelineID(pipelineID).
		Return(nil, storageErr).
		Times(1)
	storage.
		EXPECT().
		GetPipeline(notExistingID).
		Return(nil, notFoundErr).
		Times(1)

	taskrepo := taskrepo.New()
	service := New(storage, taskrepo, uuidGen, freezed)

	tests := []struct {
		name     string
		id       string
		expected []*domain.Job
		err      error
	}{
		{
			"ok",
			pipelineID,
			jobs,
			nil,
		},
		{
			"storage error",
			pipelineID,
			nil,
			storageErr,
		},
		{
			"not found",
			notExistingID,
			nil,
			notFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs, err := service.GetPipelineJobs(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("service get pipelines jobs returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if eq := reflect.DeepEqual(jobs, tt.expected); !eq {
					t.Errorf("service get pipelines jobs returned wrong jobs: got %#v want %#v", jobs, tt.expected)
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

	testTime := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Completed,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	updatedPipeline := &domain.Pipeline{}
	*updatedPipeline = *p
	updatedPipeline.Name = "updated pipeline_name"
	updatedPipeline.Description = "updated description"

	invalidID := "invalid_id"

	storageErr := errors.New("some storage error")
	pipelineNotFoundErr := errors.New("pipeline not found")

	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(2)
	storage.
		EXPECT().
		UpdatePipeline(p.ID, updatedPipeline).
		Return(storageErr).
		Times(1)
	storage.
		EXPECT().
		UpdatePipeline(p.ID, updatedPipeline).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		GetPipeline(invalidID).
		Return(nil, pipelineNotFoundErr).
		Times(1)

	taskrepo := taskrepo.New()
	service := New(storage, taskrepo, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"storage error",
			p.ID,
			storageErr,
		},
		{
			"ok",
			p.ID,
			nil,
		},
		{
			"not found",
			invalidID,
			pipelineNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(tt.id, updatedPipeline.Name, updatedPipeline.Description)
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

	testTime := freezed.Now()
	p := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Failed,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	invalidID := "invalid_id"
	notExistingID := "not_existing_id"
	notFoundErr := &apperrors.NotFoundErr{ID: invalidID, ResourceName: "pipeline"}
	storageErr := errors.New("some storage error")
	uuidGen := mock.NewMockUUIDGenerator(ctrl)

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetPipeline(p.ID).
		Return(p, nil).
		Times(1)
	storage.
		EXPECT().
		GetPipeline(invalidID).
		Return(p, nil).
		Times(1)
	storage.
		EXPECT().
		GetPipeline(notExistingID).
		Return(nil, notFoundErr).
		Times(1)
	storage.
		EXPECT().
		DeletePipeline(p.ID).
		Return(nil).
		Times(1)
	storage.
		EXPECT().
		DeletePipeline(invalidID).
		Return(storageErr).
		Times(1)

	taskrepo := taskrepo.New()
	service := New(storage, taskrepo, uuidGen, freezed)

	tests := []struct {
		name string
		id   string
		err  error
	}{
		{
			"ok",
			p.ID,
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
