package jobqueue

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/mock"
)

var (
	jobqueue *redisqueue
)

func TestMain(m *testing.M) {
	redisURL := os.Getenv("REDIS_URL")
	cfg := config.Redis{
		URL:          redisURL,
		KeyPrefix:    "",
		PoolSize:     1,
		MinIdleConns: 5,
	}
	jobqueue = NewRedisQueue(cfg, "text")
	jobqueue.logger = &logrus.Logger{Out: ioutil.Discard}
	defer jobqueue.Close()

	m.Run()
}

func TestRedisQueuePush(t *testing.T) {
	defer jobqueue.FlushDB(ctx)

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
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.Name = "second_job"

	if err := jobqueue.Push(job); err != nil {
		t.Errorf("redisqueue could not push job to queue: got %#v want nil", err)
	}
	if err := jobqueue.Push(secondJob); err != nil {
		t.Errorf("redisqueue could not push job to queue: got %#v want nil", err)
	}

	queueJobBytes, _ := jobqueue.Client.RPop(ctx, "job-queue").Bytes()
	queueSecondJobBytes, _ := jobqueue.Client.RPop(ctx, "job-queue").Bytes()

	queueJob := &domain.Job{}
	queueSecondJob := &domain.Job{}

	json.Unmarshal(queueJobBytes, queueJob)
	json.Unmarshal(queueSecondJobBytes, queueSecondJob)

	if eq := reflect.DeepEqual(queueJob, job); !eq {
		t.Errorf("redisqueue returned wrong job: got %v want %v", queueJob, job)
	}
	if eq := reflect.DeepEqual(queueSecondJob, secondJob); !eq {
		t.Errorf("redisqueue returned wrong job: got %v want %v", queueJob, job)
	}
}

func TestRedisQueuePop(t *testing.T) {
	defer jobqueue.FlushDB(ctx)

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
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.Name = "second_job"

	if err := jobqueue.Push(job); err != nil {
		t.Errorf("redisqueue could not push job to queue: got %#v want nil", err)
	}
	if err := jobqueue.Push(secondJob); err != nil {
		t.Errorf("redisqueue could not push job to queue: got %#v want nil", err)
	}

	// give some time for the AMQP call
	time.Sleep(300 * time.Millisecond)
	queueJob := jobqueue.Pop()
	if queueJob == nil {
		t.Errorf("redisqueue pop did not return job: got nil want %#v", queueJob)
	} else {
		if eq := reflect.DeepEqual(queueJob, job); !eq {
			t.Errorf("redisqueue pop returned wrong job: got %#v want %#v", queueJob, job)
		}
	}

	queueSecondJob := jobqueue.Pop()
	if queueSecondJob == nil {
		t.Errorf("redisqueue pop did not return job: got nil want %#v", queueSecondJob)
	} else {
		if eq := reflect.DeepEqual(queueSecondJob, secondJob); !eq {
			t.Errorf("redisqueue pop returned wrong job: got %#v want %#v", queueSecondJob, secondJob)
		}
	}
}

func TestRedisQueueClose(t *testing.T) {
	expected := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Status: domain.Pending,
	}

	redisURL := os.Getenv("REDIS_URL")
	cfg := config.Redis{
		URL:          redisURL,
		KeyPrefix:    "",
		PoolSize:     1,
		MinIdleConns: 5,
	}
	redisTestClose := NewRedisQueue(cfg, "text")
	redisTestClose.logger = &logrus.Logger{Out: ioutil.Discard}
	redisTestClose.Close()

	if err := redisTestClose.Push(expected); err == nil {
		t.Errorf("redisqueue pushed on closed queue: got %#v want some err", err)
	}
}
