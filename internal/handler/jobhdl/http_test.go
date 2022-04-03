package jobhdl

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

	"valet/internal/core/domain"
	"valet/mock"
	"valet/pkg/apperrors"
)

var testTime = "1985-05-04T04:32:53.651387234Z"

func TestPostJobs(t *testing.T) {
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
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobServiceErr := errors.New("some job service error")
	jobValidationErr := &apperrors.ResourceValidationErr{Message: "some job validation error"}
	fullQueueErr := &apperrors.FullQueueErr{}

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
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, fullQueueErr).
		Times(1)
	jobService.
		EXPECT().
		Create("", job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, jobValidationErr).
		Times(1)

	hdl := NewJobHTTPHandler(jobService)

	tests := []struct {
		name    string
		payload string
		status  int
		message string
	}{
		{
			"ok",
			`{
				"name":"job_name", 
				"description": "some description", 
				"timeout": 10,
				"task_params": "some task params", 
				"task_name": "test_task"
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
				"task_params": "some task params", 
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
				"task_params": "some task params", 
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
				"task_params": "some task params", 
				"task_name": "test_task"
			}`,
			http.StatusServiceUnavailable,
			"job queue is full - try again later",
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

			if rr.Code != tt.status {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.status)
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
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.status),
					"message": tt.message,
				}
			}

			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("jobhdl create returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestGetJob(t *testing.T) {
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
		TaskParams:  "some task params",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
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

	hdl := NewJobHTTPHandler(jobService)

	tests := []struct {
		name    string
		id      string
		status  int
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

			if rr.Code != tt.status {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.status)
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
					"code":    float64(tt.status),
					"message": tt.message,
				}
			}
			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestPatchJob(t *testing.T) {
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

	hdl := NewJobHTTPHandler(jobService)

	tests := []struct {
		name    string
		id      string
		payload string
		status  int
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

			if rr.Code != tt.status {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.status)
			}

			var expected map[string]interface{}
			if rr.Code != http.StatusNoContent {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.status),
					"message": tt.message,
				}

				if eq := reflect.DeepEqual(res, expected); !eq {
					t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
				}
			}
		})
	}
}

func TestDeleteJob(t *testing.T) {
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

	hdl := NewJobHTTPHandler(jobService)

	tests := []struct {
		name    string
		id      string
		status  int
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

			if rr.Code != tt.status {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.status)
			}

			var expected map[string]interface{}
			if rr.Code != http.StatusNoContent {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.status),
					"message": tt.message,
				}

				if eq := reflect.DeepEqual(res, expected); !eq {
					t.Errorf("jobhdl get returned wrong body: got %v want %v", res, expected)
				}
			}
		})
	}
}
