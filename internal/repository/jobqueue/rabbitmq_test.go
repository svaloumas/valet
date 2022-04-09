package jobqueue

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"valet/internal/config"
	"valet/internal/core/domain"
	"valet/mock"
)

func TestRabbitMQPush(t *testing.T) {
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

	cfg := config.RabbitMQ{
		QueueName: "test",
	}
	jobqueue := NewRabbitMQ(cfg, "text")
	defer jobqueue.Close()
	jobqueue.logger = &logrus.Logger{Out: ioutil.Discard}

	if ok := jobqueue.Push(job); !ok {
		t.Errorf("rabbitmq could not push job to queue: got %#v want true", ok)
	}
}

func TestRabbitMQPop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)

	createdAt := freezed.Now()
	expected := &domain.Job{
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

	cfg := config.RabbitMQ{
		QueueName: "test",
	}
	jobqueue := NewRabbitMQ(cfg, "text")
	defer jobqueue.Close()
	jobqueue.logger = &logrus.Logger{Out: ioutil.Discard}

	if ok := jobqueue.Push(expected); !ok {
		t.Errorf("rabbitmq could not push job to queue: got %#v want true", ok)
	}
	// give some time for the AMQP call
	time.Sleep(200 * time.Millisecond)
	job := jobqueue.Pop()
	if job == nil {
		t.Errorf("rabbitmq pop did not return job: got nil want %#v", job)
	} else {
		if eq := reflect.DeepEqual(job, expected); !eq {
			t.Errorf("rabbitmq pop returned wrong job: got %#v want %#v", job, expected)
		}
	}
}

func TestRabbitMQClose(t *testing.T) {
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

	cfg := config.RabbitMQ{
		QueueName: "test",
	}
	jobqueue := NewRabbitMQ(cfg, "text")
	defer jobqueue.Close()
	jobqueue.logger = &logrus.Logger{Out: ioutil.Discard}

	jobqueue.Close()
	if ok := jobqueue.Push(expected); ok {
		t.Errorf("rabbitmq push was successful on closed queue: got %#v want false", ok)
	}
}
