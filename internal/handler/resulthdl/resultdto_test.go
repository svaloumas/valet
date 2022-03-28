package resulthdl

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
)

func TestBuildResponseBodyDTO(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	expected := ResponseBodyDTO(result)
	dto := BuildResponseBodyDTO(result)

	if eq := reflect.DeepEqual(dto, expected); !eq {
		t.Errorf("resulthdl BuildResponseBodyDTO returned wrong dto: got %#v want %#v", dto, expected)
	}
}
