package domain

import (
	"reflect"
	"testing"
)

func TestFutureJobResultWaitOK(t *testing.T) {
	expected := JobResult{
		JobID:    "job_id",
		Metadata: "some metadata",
		Error:    "",
	}
	resultQueue := make(chan JobResult, 1)
	futureResult := FutureJobResult{Result: resultQueue}
	go func() {
		resultQueue <- expected
	}()
	jobResult, _ := futureResult.Wait()
	if eq := reflect.DeepEqual(jobResult, expected); !eq {
		t.Errorf("FutureJobResult Wait returned wrong job result: got %v, want %v", jobResult, expected)
	}
}

func TestFutureJobResultWaitNotOK(t *testing.T) {
	resultQueue := make(chan JobResult, 1)
	futureResult := FutureJobResult{Result: resultQueue}

	close(resultQueue)
	jobResult, ok := futureResult.Wait()
	if ok {
		t.Errorf("FutureJobResult Wait returned unexpected job result: got %v, want nil", jobResult)
	}
}
