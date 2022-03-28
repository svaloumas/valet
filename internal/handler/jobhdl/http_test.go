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
	"valet/internal/core/domain"
	"valet/mock"
	"valet/pkg/apperrors"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "test_task",
		Description: "some description",
		Metadata:    "some metadata",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobServiceErr := errors.New("some job service error")
	jobValidationErr := &apperrors.ResourceValidationErr{Message: "some job validation error"}

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		Create(expectedJob.Name, expectedJob.TaskType, expectedJob.Description, expectedJob.Metadata).
		Return(expectedJob, nil).
		Times(1)
	jobService.
		EXPECT().
		Create(expectedJob.Name, expectedJob.TaskType, expectedJob.Description, expectedJob.Metadata).
		Return(nil, jobServiceErr).
		Times(1)
	jobService.
		EXPECT().
		Create("", expectedJob.TaskType, expectedJob.Description, expectedJob.Metadata).
		Return(nil, jobValidationErr).
		Times(1)

	hdl := NewJobHTTPHandler(jobService)

	tests := []struct {
		payload string
		status  int
		message string
	}{
		{
			`{
				"name":"job_name", 
				"description": "some description", 
				"metadata": "some metadata", 
				"task_type": "test_task"
			}`,
			http.StatusAccepted,
			"",
		},
		{
			`{
				"name":"job_name", 
				"description": "some description", 
				"metadata": "some metadata", 
				"task_type": "test_task"
			}`,
			http.StatusInternalServerError,
			"some job service error",
		},
		{
			`{
				"description": "some description", 
				"metadata": "some metadata", 
				"task_type": "test_task"
			}`,
			http.StatusBadRequest,
			"some job validation error",
		},
	}

	for _, tt := range tests {
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
				"id":          expectedJob.ID,
				"name":        expectedJob.Name,
				"description": expectedJob.Description,
				"task_type":   expectedJob.TaskType,
				"metadata":    expectedJob.Metadata,
				"status":      expectedJob.Status.String(),
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
	expectedJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskType:    "test_task",
		Description: "some description",
		Metadata:    "some metadata",
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}

	jobNotFoundErr := &apperrors.NotFoundErr{ID: expectedJob.ID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")

	jobService := mock.NewMockJobService(ctrl)
	jobService.
		EXPECT().
		Get(expectedJob.ID).
		Return(expectedJob, nil).
		Times(1)
	jobService.
		EXPECT().
		Get("invalid_id").
		Return(nil, jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Get(expectedJob.ID).
		Return(nil, jobServiceErr).
		Times(1)

	hdl := NewJobHTTPHandler(jobService)

	tests := []struct {
		id      string
		status  int
		message string
	}{
		{expectedJob.ID, http.StatusOK, ""},
		{"invalid_id", http.StatusNotFound, "job with ID: auuid4 not found"},
		{expectedJob.ID, http.StatusInternalServerError, "some job service error"},
	}

	for _, tt := range tests {
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
				"id":          expectedJob.ID,
				"name":        expectedJob.Name,
				"description": expectedJob.Description,
				"task_type":   expectedJob.TaskType,
				"metadata":    expectedJob.Metadata,
				"status":      expectedJob.Status.String(),
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
		id      string
		payload string
		status  int
		message string
	}{
		{
			jobID,
			`{
				"name": "updated_name", 
				"description": "updated description" 
			}`,
			http.StatusNoContent, "",
		},
		{
			"invalid_id",
			`{
				"name": "updated_name",
				"description": "updated description"
			}`,
			http.StatusNotFound,
			"job with ID: auuid4 not found",
		},
		{
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
		id      string
		status  int
		message string
	}{
		{jobID, http.StatusNoContent, ""},
		{"invalid_id", http.StatusNotFound, "job with ID: auuid4 not found"},
		{jobID, http.StatusInternalServerError, "some job service error"},
	}

	for _, tt := range tests {
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
	}
}
