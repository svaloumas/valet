package domain

import (
	"reflect"
	"testing"
)

func TestFutureResultWait(t *testing.T) {
	expected := JobResult{
		JobID:    "job_id",
		Metadata: "some metadata",
		Error:    nil,
	}
	resultQueue := make(chan JobResult, 1)
	futureResult := FutureJobResult{Result: resultQueue}
	go func() {
		resultQueue <- expected
	}()
	jobResult := futureResult.Wait()
	if eq := reflect.DeepEqual(jobResult, expected); !eq {
		t.Errorf("FutureJobResult Wait returned wrong job result: got %v, want %v", jobResult, expected)
	}
}
