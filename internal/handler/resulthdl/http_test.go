package resulthdl

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/mock"
	"valet/pkg/apperrors"
)

func TestHTTPGetJobResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	gin.SetMode(gin.ReleaseMode)

	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some error message",
	}

	resultNotFoundErr := &apperrors.NotFoundErr{ID: result.JobID, ResourceName: "job result"}
	resultServiceErr := errors.New("some result service error")

	resultService := mock.NewMockResultService(ctrl)
	resultService.
		EXPECT().
		Get(result.JobID).
		Return(result, nil).
		Times(1)
	resultService.
		EXPECT().
		Get("invalid_id").
		Return(nil, resultNotFoundErr).
		Times(1)
	resultService.
		EXPECT().
		Get(result.JobID).
		Return(nil, resultServiceErr).
		Times(1)

	hdl := NewResultHTTPHandler(resultService)

	tests := []struct {
		name    string
		id      string
		status  int
		message string
	}{
		{"ok", result.JobID, http.StatusOK, ""},
		{"not found", "invalid_id", http.StatusNotFound, "job result with ID: auuid4 not found"},
		{"internal server error", result.JobID, http.StatusInternalServerError, "some result service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.GET("/api/jobs/:id/results", hdl.Get)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/jobs/"+tt.id+"/results", nil)

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
					"job_id":   result.JobID,
					"metadata": result.Metadata,
					"error":    result.Error,
				}
			} else {
				expected = map[string]interface{}{
					"error":   true,
					"code":    float64(tt.status),
					"message": tt.message,
				}
			}
			if eq := reflect.DeepEqual(res, expected); !eq {
				t.Errorf("resulthdl get returned wrong body: got %v want %v", res, expected)
			}
		})
	}
}

func TestHTTPDeleteJobResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobID := "auuid4"

	resultNotFoundErr := &apperrors.NotFoundErr{ID: jobID, ResourceName: "job result"}
	resultServiceErr := errors.New("some result service error")

	resultService := mock.NewMockResultService(ctrl)
	resultService.
		EXPECT().
		Delete(jobID).
		Return(nil).
		Times(1)
	resultService.
		EXPECT().
		Delete("invalid_id").
		Return(resultNotFoundErr).
		Times(1)
	resultService.
		EXPECT().
		Delete(jobID).
		Return(resultServiceErr).
		Times(1)

	hdl := NewResultHTTPHandler(resultService)

	tests := []struct {
		name    string
		id      string
		status  int
		message string
	}{
		{"ok", jobID, http.StatusNoContent, ""},
		{"not found", "invalid_id", http.StatusNotFound, "job result with ID: auuid4 not found"},
		{"internal server error", jobID, http.StatusInternalServerError, "some result service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rr)

			r.DELETE("/api/jobs/:id/results", hdl.Delete)
			c.Request, _ = http.NewRequest(http.MethodDelete, "/api/jobs/"+tt.id+"/results", nil)

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
					t.Errorf("resulthdl get returned wrong body: got %v want %v", res, expected)
				}
			}
		})
	}
}
