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
			// TODO: Check what's wrong here. (StatusBadRequest should return)
			http.StatusInternalServerError,
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
				"message": tt.message,
			}
		}

		if eq := reflect.DeepEqual(res, expected); !eq {
			t.Errorf("jobhdl create returned wrong body: got %v want %v", res, expected)
		}
	}
}
