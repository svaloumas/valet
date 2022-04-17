package jobhdl

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

func TestHTTPPostJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	gin.SetMode(gin.ReleaseMode)

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
	*jobWithSchedule = *job
	runAt := createdAt.Add(10 * time.Minute)
	jobWithSchedule.RunAt = &runAt

	jobServiceErr := errors.New("some job service error")
	jobValidationErr := &apperrors.ResourceValidationErr{Message: "some job validation error"}
	fullQueueErr := &apperrors.FullQueueErr{}
	parseTimeErr := &apperrors.ParseTimeErr{}

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(job, nil).
		Times(1)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, jobServiceErr).
		Times(1)
	jobService.
		EXPECT().
		Create("", job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, jobValidationErr).
		Times(1)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "invalid_timestamp_format", job.Timeout, job.TaskParams).
		Return(nil, parseTimeErr).
		Times(1)
	jobService.
		EXPECT().
		Create(jobWithSchedule.Name, jobWithSchedule.TaskName, jobWithSchedule.Description, "2006-01-02T15:04:05.999999999Z", jobWithSchedule.Timeout, jobWithSchedule.TaskParams).
		Return(jobWithSchedule, nil).
		Times(1)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(job, nil).
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
		Return(fullQueueErr).
		Times(1)

	hdl := NewJobHTTPHandler(jobService, jobQueue)

	tests := []struct {
		name    string
		payload string
		code    int
		message string
	}{
		{
			"ok",
			`{
				"name":"job_name",
				"description": "some description",
				"timeout": 10,
				"task_params": {
					"url": "some-url.com"
				},
				"task_name": "test_task"
			}`,
			http.StatusAccepted,
			"",
		},
		{
			"ok with schedule",
			`{
				"name":"job_name",
				"description": "some description",
				"timeout": 10,
				"task_params": {
					"url": "some-url.com"
				},
				"task_name": "test_task",
				"run_at": "2006-01-02T15:04:05.999999999Z"
			}`,
			http.StatusAccepted,
			"",
		},
		{
			"internal server error",
			`{
				"name":"job_name",
				"description": "some description",
				"timeout": 10,
				"task_params": {
					"url": "some-url.com"
				},
				"task_name": "test_task"
			}`,
			http.StatusInternalServerError,
			"some job service error",
		},
		{
			"job validation error",
			`{
				"description": "some description",
				"timeout": 10,
				"task_params": {
					"url": "some-url.com"
				},
				"task_name": "test_task"
			}`,
			http.StatusBadRequest,
			"some job validation error",
		},
		{
			"service unavailable error",
			`{
				"name":"job_name",
				"description": "some description",
				"timeout": 10,
				"task_params": {
					"url": "some-url.com"
				},
				"task_name": "test_task"
			}`,
			http.StatusServiceUnavailable,
			"job queue is full - try again later",
		},
		{
			"invalid timestamp format",
			`{
				"name":"job_name",
				"description": "some description",
				"timeout": 10,
				"task_params": {
					"url": "some-url.com"
				},
				"task_name": "test_task",
				"run_at": "invalid_timestamp_format"
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

			r.POST("/api/jobs", hdl.Create)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/jobs", body)

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
					"id":          job.ID,
					"name":        job.Name,
					"description": job.Description,
					"task_name":   job.TaskName,
					"timeout":     float64(job.Timeout),
					"task_params": job.TaskParams,
					"status":      job.Status.String(),
					"created_at":  testTime,
				}
				if tt.name == "ok with schedule" {
					expected["run_at"] = runAtTestTime
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

func TestHTTPGetJob(t *testing.T) {
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
		TaskName:    "test_task",
		Description: "some description",
		Timeout:     10,
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}

	jobNotFoundErr := &apperrors.NotFoundErr{ID: job.ID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		Get(job.ID).
		Return(job, nil).
		Times(1)
	jobService.
		EXPECT().
		Get("invalid_id").
		Return(nil, jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Get(job.ID).
		Return(nil, jobServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewJobHTTPHandler(jobService, jobQueue)

	tests := []struct {
		name    string
		id      string
		code    int
		message string
	}{
		{"ok", job.ID, http.StatusOK, ""},
		{"not found", "invalid_id", http.StatusNotFound, "job with ID: auuid4 not found"},
		{"internal server error", job.ID, http.StatusInternalServerError, "some job service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.GET("/api/jobs/:id", hdl.Get)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/jobs/"+tt.id, nil)

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
					"id":          job.ID,
					"name":        job.Name,
					"description": job.Description,
					"task_name":   job.TaskName,
					"timeout":     float64(job.Timeout),
					"task_params": job.TaskParams,
					"status":      job.Status.String(),
					"created_at":  testTime,
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.code),
					"message": tt.message,
				}
			}
			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPGetJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testNow := freezed.Now()
	pendingJob := &domain.Job{
		ID:        "pending_job_id",
		Name:      "pending_job",
		TaskName:  "some task",
		Status:    domain.Pending,
		CreatedAt: &testNow,
		RunAt:     &testNow,
	}
	scheduledJob := &domain.Job{
		ID:        "scheduled_job_id",
		Name:      "scheduled_job",
		TaskName:  "some other task",
		Status:    domain.Scheduled,
		CreatedAt: &testNow,
		RunAt:     &testNow,
	}
	inprogressJob := &domain.Job{
		ID:        "inprogress_job_id",
		Name:      "inprogress_job",
		TaskName:  "some task",
		Status:    domain.InProgress,
		CreatedAt: &testNow,
		RunAt:     &testNow,
	}
	completedJob := &domain.Job{
		ID:        "completed_job_id",
		Name:      "completed_job",
		TaskName:  "some task",
		Status:    domain.Completed,
		CreatedAt: &testNow,
		RunAt:     &testNow,
	}
	failedJob := &domain.Job{
		ID:        "failed_job_id",
		Name:      "failed_job",
		TaskName:  "some task",
		Status:    domain.Failed,
		CreatedAt: &testNow,
		RunAt:     &testNow,
	}

	pendingJobs := []*domain.Job{pendingJob}
	scheduledJobs := []*domain.Job{scheduledJob}
	inprogressJobs := []*domain.Job{inprogressJob}
	completedJobs := []*domain.Job{completedJob}
	failedJobs := []*domain.Job{failedJob}

	jobStatusValidationErr := &apperrors.ResourceValidationErr{Message: "invalid job status: \"NOT_STARTED\""}
	jobServiceErr := errors.New("some job service error")

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		GetJobs("pending").
		Return(pendingJobs, nil).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("scheduled").
		Return(scheduledJobs, nil).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("in_progress").
		Return(inprogressJobs, nil).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("completed").
		Return(completedJobs, nil).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("failed").
		Return(failedJobs, nil).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("failed").
		Return(nil, jobServiceErr).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("not_started").
		Return(nil, jobStatusValidationErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewJobHTTPHandler(jobService, jobQueue)

	tests := []struct {
		name    string
		status  string
		job     *domain.Job
		code    int
		message string
	}{
		{
			"pending",
			"pending",
			pendingJob,
			http.StatusOK,
			"",
		},
		{
			"scheduled",
			"scheduled",
			scheduledJob,
			http.StatusOK,
			"",
		},
		{
			"in progress",
			"in_progress",
			inprogressJob,
			http.StatusOK,
			"",
		},
		{
			"completed",
			"completed",
			completedJob,
			http.StatusOK,
			"",
		},
		{
			"failed",
			"failed",
			failedJob,
			http.StatusOK,
			"",
		},
		{
			"job status validation error",
			"not_started",
			nil,
			http.StatusBadRequest,
			"invalid job status: \"NOT_STARTED\"",
		},
		{
			"internal server error",
			"failed",
			nil,
			http.StatusInternalServerError,
			"some job service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.GET("/api/jobs", hdl.GetJobs)
			c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/jobs?status=%s", tt.status), nil)

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
							"id":         tt.job.ID,
							"name":       tt.job.Name,
							"task_name":  tt.job.TaskName,
							"status":     tt.job.Status.String(),
							"created_at": testTime,
							"run_at":     testTime,
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
				t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}
func TestHTTPPatchJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobID := "auuid4"
	jobNotFoundErr := &apperrors.NotFoundErr{ID: jobID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		Update(jobID, "updated_name", "updated description").
		Return(nil).
		Times(1)
	jobService.
		EXPECT().
		Update("invalid_id", "updated_name", "updated description").
		Return(jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Update(jobID, "updated_name", "updated description").
		Return(jobServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewJobHTTPHandler(jobService, jobQueue)

	tests := []struct {
		name    string
		id      string
		payload string
		code    int
		message string
	}{
		{
			"ok",
			jobID,
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
			"job with ID: auuid4 not found",
		},
		{
			"internal server error",
			jobID,
			`{
				"name": "updated_name",
				"description": "updated description"
			}`,
			http.StatusInternalServerError,
			"some job service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			body := strings.NewReader(tt.payload)

			r.PATCH("/api/jobs/:id", hdl.Update)
			c.Request, _ = http.NewRequest(http.MethodPatch, "/api/jobs/"+tt.id, body)

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

func TestHTTPDeleteJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobID := "auuid4"

	jobNotFoundErr := &apperrors.NotFoundErr{ID: jobID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		Delete(jobID).
		Return(nil).
		Times(1)
	jobService.
		EXPECT().
		Delete("invalid_id").
		Return(jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Delete(jobID).
		Return(jobServiceErr).
		Times(1)

	jobQueue := mock.NewMockJobQueue(ctrl)

	hdl := NewJobHTTPHandler(jobService, jobQueue)

	tests := []struct {
		name    string
		id      string
		code    int
		message string
	}{
		{"ok", jobID, http.StatusNoContent, ""},
		{"not found", "invalid_id", http.StatusNotFound, "job with ID: auuid4 not found"},
		{"internal server error", jobID, http.StatusInternalServerError, "some job service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.DELETE("/api/jobs/:id", hdl.Delete)
			c.Request, _ = http.NewRequest(http.MethodDelete, "/api/jobs/"+tt.id, nil)

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
