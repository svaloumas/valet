package jobhdl

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
	"valet/mock"
)

func TestBuildResponseBodyDTO(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	testTime := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams:  "some params",
		Status:      domain.Completed,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	expected := ResponseBodyDTO(job)
	dto := BuildResponseBodyDTO(job)

	if eq := reflect.DeepEqual(dto, expected); !eq {
		t.Errorf("jobhdl BuildResponseBodyDTO returned wrong dto: got %#v want %#v", dto, expected)
	}
}
