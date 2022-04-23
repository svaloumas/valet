package pipelinehdl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
)

var testTime = "1985-05-04T04:32:53.651387234Z"
var runAtTestTime = "1985-05-04T04:42:53.651387234Z"

func TestHTTPPostPipelines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	gin.SetMode(gin.ReleaseMode)

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	job := &domain.Job{
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: false,
	}
	secondJob := &domain.Job{
		Name:        "second_job",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: true,
	}

	jobs := []*domain.Job{
		job, secondJob,
	}

	createdAt := freezed.Now()
	pipeline := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		Jobs:        jobs,
		CreatedAt:   &createdAt,
	}

	createdJob := &domain.Job{}
	*createdJob = *job
	createdJob.ID = "job_id"
	createdJob.Status = domain.Pending
	createdJob.CreatedAt = &createdAt

	createdSecondJob := &domain.Job{}
	*createdSecondJob = *secondJob
	createdSecondJob.ID = "second_job_id"
	createdSecondJob.Status = domain.Pending
	createdSecondJob.CreatedAt = &createdAt

	createdJobsWithoutSchedule := []*domain.Job{
		createdJob, createdSecondJob,
	}

	createdPipelineWithoutSchedule := &domain.Pipeline{}
	*createdPipelineWithoutSchedule = *pipeline
	createdPipelineWithoutSchedule.Jobs = createdJobsWithoutSchedule

	pipelineServiceErr := errors.New("some pipeline service error")
	queueErr := errors.New("some internal queue error")
	pipelineValidationErr := &apperrors.ResourceValidationErr{Message: "some pipeline validation error"}
	fullQueueErr := &apperrors.FullQueueErr{}
	parseTimeErr := &apperrors.ParseTimeErr{}

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "", pipeline.Jobs).
		Return(createdPipelineWithoutSchedule, nil).
		Times(3)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "", pipeline.Jobs).
		Return(nil, pipelineServiceErr).
		Times(1)
	pipelineService.
		EXPECT().
		Create("", pipeline.Description, "", pipeline.Jobs).
		Return(nil, pipelineValidationErr).
		Times(1)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "invalid_timestamp_format", pipeline.Jobs).
		Return(nil, parseTimeErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Push(createdJob).
		Return(nil).
		Times(1)
	jobQueue.
		EXPECT().
		Push(createdJob).
		Return(fullQueueErr).
		Times(1)
	jobQueue.
		EXPECT().
		Push(createdJob).
		Return(queueErr).
		Times(1)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name     string
		pipeline *domain.Pipeline
		payload  string
		code     int
		message  string
	}{
		{
			"ok",
			createdPipelineWithoutSchedule,
			`{
				"name": "pipeline_name",
				"description": "some description",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusAccepted,
			"",
		},
		{
			"full queue error",
			nil,
			`{
				"name": "pipeline_name",
				"description": "some description",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusServiceUnavailable,
			"job queue is full - try again later",
		},
		{
			"internal queue error",
			nil,
			`{
				"name": "pipeline_name",
				"description": "some description",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusInternalServerError,
			"some internal queue error",
		},
		{
			"storage error",
			nil,
			`{
				"name": "pipeline_name",
				"description": "some description",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusInternalServerError,
			"some pipeline service error",
		},
		{
			"resource validation error",
			nil,
			`{
				"description": "some description",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusBadRequest,
			"some pipeline validation error",
		},
		{
			"invalid timestamp format",
			nil,
			`{
				"name": "pipeline_name",
				"description": "some description",
				"run_at": "invalid_timestamp_format",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusBadRequest,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			body := strings.NewReader(tt.payload)

			r.POST("/api/pipelines", hdl.Create)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/pipelines", body)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code == http.StatusAccepted {
				expected = map[string]interface{}{
					"id":          tt.pipeline.ID,
					"name":        tt.pipeline.Name,
					"description": tt.pipeline.Description,
					"status":      tt.pipeline.Status.String(),
					"created_at":  testTime,
					"jobs": []interface{}{
						map[string]interface{}{
							"id":          tt.pipeline.Jobs[0].ID,
							"name":        tt.pipeline.Jobs[0].Name,
							"description": tt.pipeline.Jobs[0].Description,
							"status":      tt.pipeline.Jobs[0].Status.String(),
							"task_name":   tt.pipeline.Jobs[0].TaskName,
							"task_params": tt.pipeline.Jobs[0].TaskParams,
							"timeout":     float64(tt.pipeline.Jobs[0].Timeout),
							"created_at":  testTime,
						},
						map[string]interface{}{
							"id":                   tt.pipeline.Jobs[1].ID,
							"name":                 tt.pipeline.Jobs[1].Name,
							"description":          tt.pipeline.Jobs[1].Description,
							"status":               tt.pipeline.Jobs[1].Status.String(),
							"task_name":            tt.pipeline.Jobs[1].TaskName,
							"task_params":          tt.pipeline.Jobs[1].TaskParams,
							"timeout":              float64(tt.pipeline.Jobs[1].Timeout),
							"use_previous_results": true,
							"created_at":           testTime,
						},
					},
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}
			}

			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("jobhdl create returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPPostPipelinesWithSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	gin.SetMode(gin.ReleaseMode)

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	job := &domain.Job{
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: false,
	}
	secondJob := &domain.Job{
		Name:        "second_job",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: true,
	}

	runAt := "1985-05-04T04:42:53.651387234Z"
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAt)

	jobs := []*domain.Job{
		job, secondJob,
	}

	createdAt := freezed.Now()
	pipeline := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		Jobs:        jobs,
		CreatedAt:   &createdAt,
	}

	createdJobWithSchedule := &domain.Job{}
	*createdJobWithSchedule = *job
	createdJobWithSchedule.ID = "job_id"
	createdJobWithSchedule.Status = domain.Pending
	createdJobWithSchedule.CreatedAt = &createdAt
	createdJobWithSchedule.RunAt = &runAtTime

	createdSecondJob := &domain.Job{}
	*createdSecondJob = *secondJob
	createdSecondJob.ID = "second_job_id"
	createdSecondJob.Status = domain.Pending
	createdSecondJob.CreatedAt = &createdAt

	createdJobsWithSchedule := []*domain.Job{
		createdJobWithSchedule, createdSecondJob,
	}

	createdPipelineWithSchedule := &domain.Pipeline{}
	*createdPipelineWithSchedule = *pipeline
	createdPipelineWithSchedule.Jobs = createdJobsWithSchedule
	createdPipelineWithSchedule.RunAt = &runAtTime

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, runAt, pipeline.Jobs).
		Return(createdPipelineWithSchedule, nil).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name     string
		pipeline *domain.Pipeline
		payload  string
		code     int
		message  string
	}{
		{
			"ok with schedule",
			createdPipelineWithSchedule,
			`{
				"name": "pipeline_name",
				"description": "some description",
				"run_at": "1985-05-04T04:42:53.651387234Z",
				"jobs": [
					{
						"name": "job_name",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						}
					},
					{
						"name": "second_job",
						"description": "some description",
						"task_name": "test_task",
						"timeout": 10,
						"task_params": {
							"url": "some-url.com"
						},
						"use_previous_results": true
					}
				]
			}`,
			http.StatusAccepted,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			body := strings.NewReader(tt.payload)

			r.POST("/api/pipelines", hdl.Create)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/pipelines", body)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code == http.StatusAccepted {
				expected = map[string]interface{}{
					"id":          tt.pipeline.ID,
					"name":        tt.pipeline.Name,
					"description": tt.pipeline.Description,
					"status":      tt.pipeline.Status.String(),
					"run_at":      runAtTestTime,
					"created_at":  testTime,
					"jobs": []interface{}{
						map[string]interface{}{
							"id":          tt.pipeline.Jobs[0].ID,
							"name":        tt.pipeline.Jobs[0].Name,
							"description": tt.pipeline.Jobs[0].Description,
							"status":      tt.pipeline.Jobs[0].Status.String(),
							"task_name":   tt.pipeline.Jobs[0].TaskName,
							"task_params": tt.pipeline.Jobs[0].TaskParams,
							"timeout":     float64(tt.pipeline.Jobs[0].Timeout),
							"run_at":      runAtTestTime,
							"created_at":  testTime,
						},
						map[string]interface{}{
							"id":                   tt.pipeline.Jobs[1].ID,
							"name":                 tt.pipeline.Jobs[1].Name,
							"description":          tt.pipeline.Jobs[1].Description,
							"status":               tt.pipeline.Jobs[1].Status.String(),
							"task_name":            tt.pipeline.Jobs[1].TaskName,
							"task_params":          tt.pipeline.Jobs[1].TaskParams,
							"timeout":              float64(tt.pipeline.Jobs[1].Timeout),
							"use_previous_results": true,
							"created_at":           testTime,
						},
					},
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}
			}

			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("jobhdl create returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPGetPipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(3)

	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := freezed.Now()
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAtTestTime)
	p := &domain.Pipeline{
		ID:          "auuid4",
		Name:        "a_pipeline",
		Description: "some description",
		Status:      domain.Completed,
		RunAt:       &runAtTime,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}
	p.SetDuration()

	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: "invalid_id", ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		Get(p.ID).
		Return(p, nil).
		Times(1)
	pipelineService.
		EXPECT().
		Get("invalid_id").
		Return(nil, pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		Get(p.ID).
		Return(nil, pipelineServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name    string
		id      string
		code    int
		message string
	}{
		{"ok", p.ID, http.StatusOK, ""},
		{"not found", "invalid_id", http.StatusNotFound, "pipeline with ID: invalid_id not found"},
		{"internal server error", p.ID, http.StatusInternalServerError, "some pipeline service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.GET("/api/pipelines/:id", hdl.Get)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/pipelines/"+tt.id, nil)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code == http.StatusOK {
				expected = map[string]interface{}{
					"id":           p.ID,
					"name":         p.Name,
					"description":  p.Description,
					"status":       p.Status.String(),
					"run_at":       runAtTestTime,
					"created_at":   testTime,
					"started_at":   testTime,
					"completed_at": testTime,
					"duration":     float64(0),
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}
			}
			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("pipelinehdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPGetPipelines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	pendingPipeline := &domain.Pipeline{
		ID:          "pending_pipeline_id",
		Name:        "pending_pipeline",
		Description: "some description",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	inprogressPipeline := &domain.Pipeline{
		ID:          "inprogress_pipeline_id",
		Name:        "inprogress_pipeline",
		Description: "some description",
		Status:      domain.InProgress,
		CreatedAt:   &createdAt,
	}
	failedPipeline := &domain.Pipeline{
		ID:          "failed_pipeline_id",
		Name:        "failed_pipeline",
		Description: "some description",
		Status:      domain.Failed,
		CreatedAt:   &createdAt,
	}
	completedPipeline := &domain.Pipeline{
		ID:          "completed_pipeline_id",
		Name:        "completed_pipeline",
		Description: "some description",
		Status:      domain.Completed,
		CreatedAt:   &createdAt,
	}

	pendingPipelines := []*domain.Pipeline{pendingPipeline}
	inprogressPipelines := []*domain.Pipeline{inprogressPipeline}
	completedPipelines := []*domain.Pipeline{completedPipeline}
	failedPipelines := []*domain.Pipeline{failedPipeline}

	pipelineStatusValidationErr := &apperrors.ResourceValidationErr{Message: "invalid status: \"NOT_STARTED\""}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		GetPipelines("pending").
		Return(pendingPipelines, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("in_progress").
		Return(inprogressPipelines, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("completed").
		Return(completedPipelines, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("failed").
		Return(failedPipelines, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("failed").
		Return(nil, pipelineServiceErr).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("not_started").
		Return(nil, pipelineStatusValidationErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name     string
		status   string
		pipeline *domain.Pipeline
		code     int
		message  string
	}{
		{
			"pending",
			"pending",
			pendingPipeline,
			http.StatusOK,
			"",
		},
		{
			"in progress",
			"in_progress",
			inprogressPipeline,
			http.StatusOK,
			"",
		},
		{
			"completed",
			"completed",
			completedPipeline,
			http.StatusOK,
			"",
		},
		{
			"failed",
			"failed",
			failedPipeline,
			http.StatusOK,
			"",
		},
		{
			"pipeline status validation error",
			"not_started",
			nil,
			http.StatusBadRequest,
			"invalid status: \"NOT_STARTED\"",
		},
		{
			"internal server error",
			"failed",
			nil,
			http.StatusInternalServerError,
			"some pipeline service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.GET("/api/pipelines", hdl.GetPipelines)
			c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/pipelines?status=%s", tt.status), nil)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code == http.StatusOK {
				expected = map[string]interface{}{
					"pipelines": []interface{}{
						map[string]interface{}{
							"id":          tt.pipeline.ID,
							"name":        tt.pipeline.Name,
							"description": tt.pipeline.Description,
							"status":      tt.pipeline.Status.String(),
							"created_at":  testTime,
						},
					},
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}
			}
			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("pipelinehdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPGetPipelineJobs(t *testing.T) {
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
	runAtTime, _ := time.Parse(time.RFC3339Nano, runAtTestTime)
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
		PipelineID:  pipelineID,
		Timeout:     10,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:             domain.Pending,
		UsePreviousResults: true,
		CreatedAt:          &createdAt,
	}

	jobs := []*domain.Job{
		job, secondJob,
	}

	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: "invalid_id", ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		GetPipelineJobs(pipelineID).
		Return(jobs, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelineJobs("invalid_id").
		Return(nil, pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelineJobs(pipelineID).
		Return(nil, pipelineServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name    string
		id      string
		code    int
		message string
	}{
		{"ok", pipelineID, http.StatusOK, ""},
		{"not found", "invalid_id", http.StatusNotFound, "pipeline with ID: invalid_id not found"},
		{"internal server error", pipelineID, http.StatusInternalServerError, "some pipeline service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.GET("/api/pipelines/:id/jobs", hdl.GetPipelineJobs)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/pipelines/"+tt.id+"/jobs", nil)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code == http.StatusOK {
				expected = map[string]interface{}{
					"jobs": []interface{}{
						map[string]interface{}{
							"id":          job.ID,
							"name":        job.Name,
							"description": job.Description,
							"pipeline_id": job.PipelineID,
							"next_job_id": job.NextJobID,
							"task_name":   job.TaskName,
							"task_params": job.TaskParams,
							"timeout":     float64(job.Timeout),
							"status":      job.Status.String(),
							"run_at":      runAtTestTime,
							"created_at":  testTime,
						},
						map[string]interface{}{
							"id":                   secondJob.ID,
							"name":                 secondJob.Name,
							"description":          secondJob.Description,
							"pipeline_id":          secondJob.PipelineID,
							"timeout":              float64(secondJob.Timeout),
							"task_name":            secondJob.TaskName,
							"task_params":          secondJob.TaskParams,
							"status":               secondJob.Status.String(),
							"created_at":           testTime,
							"use_previous_results": true,
						},
					},
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}
			}
			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("pipelinehdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPPatchPipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pipelineID := "auuid4"
	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: pipelineID, ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		Update(pipelineID, "updated_name", "updated description").
		Return(nil).
		Times(1)
	pipelineService.
		EXPECT().
		Update("invalid_id", "updated_name", "updated description").
		Return(pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		Update(pipelineID, "updated_name", "updated description").
		Return(pipelineServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name    string
		id      string
		payload string
		code    int
		message string
	}{
		{
			"ok",
			pipelineID,
			`{
				"name": "updated_name", 
				"description": "updated description" 
			}`,
			http.StatusNoContent, "",
		},
		{
			"not found",
			"invalid_id",
			`{
				"name": "updated_name",
				"description": "updated description"
			}`,
			http.StatusNotFound,
			"pipeline with ID: auuid4 not found",
		},
		{
			"internal server error",
			pipelineID,
			`{
				"name": "updated_name",
				"description": "updated description"
			}`,
			http.StatusInternalServerError,
			"some pipeline service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			body := strings.NewReader(tt.payload)

			r.PATCH("/api/pipelines/:id", hdl.Update)
			c.Request, _ = http.NewRequest(http.MethodPatch, "/api/pipelines/"+tt.id, body)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code != http.StatusNoContent {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}

				if eq := reflect.DeepEqual(res, expected); !eq {
					t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
				}
			}
		})
	}
}

func TestHTTPDeletePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pipelineID := "auuid4"

	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: pipelineID, ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService := mock.NewMockPipelineService(ctrl)
	pipelineService.
		EXPECT().
		Delete(pipelineID).
		Return(nil).
		Times(1)
	pipelineService.
		EXPECT().
		Delete("invalid_id").
		Return(pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		Delete(pipelineID).
		Return(pipelineServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewPipelineHTTPHandler(pipelineService, jobQueue)

	tests := []struct {
		name    string
		id      string
		code    int
		message string
	}{
		{"ok", pipelineID, http.StatusNoContent, ""},
		{"not found", "invalid_id", http.StatusNotFound, "pipeline with ID: auuid4 not found"},
		{"internal server error", pipelineID, http.StatusInternalServerError, "some pipeline service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.DELETE("/api/pipelines/:id", hdl.Delete)
			c.Request, _ = http.NewRequest(http.MethodDelete, "/api/pipelines/"+tt.id, nil)

			r.ServeHTTP(rr, c.Request)

			b, _ := ioutil.ReadAll(rr.Body)

			var res map[string]interface{}
			json.Unmarshal(b, &res)

			if rr.Code != tt.code {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.code)
			}

			var expected map[string]interface{}
			if rr.Code != http.StatusNoContent {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}

				if eq := reflect.DeepEqual(res, expected); !eq {
					t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
				}
			}
		})
	}
}
