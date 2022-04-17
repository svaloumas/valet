package pipelinehdl

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/mock"
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
	p := &domain.Pipeline{
		ID:          "auuid4",
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Completed,
		Jobs: []*domain.Job{
			{Name: "a_job"},
		},
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	expected := ResponseBodyDTO(p)
	dto := BuildResponseBodyDTO(p)

	if eq := reflect.DeepEqual(dto, expected); !eq {
		t.Errorf("jobhdl BuildResponseBodyDTO returned wrong dto: got %#v want %#v", dto, expected)
	}
}
