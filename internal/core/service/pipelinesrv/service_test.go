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
