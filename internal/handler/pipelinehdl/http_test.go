package pipelinehdl

import (
	"encoding/json"
	"errors"
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
