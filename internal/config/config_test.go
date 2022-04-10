package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	server := Server{
		HTTPPort: "8080",
	}
	memoryJobQueue := MemoryJobQueue{
		Capacity: 100,
	}
	queueParams := QueueParams{
		Name:              "test",
		Durable:           false,
		DeletedWhenUnused: false,
		Exclusive:         false,
		NoWait:            false,
	}
	consumeParams := ConsumeParams{
		Name:      "rabbitmq-consumer",
		AutoACK:   true,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
	}
	publishParams := PublishParams{
		Exchange:   "",
		RoutingKey: "test",
		Mandatory:  false,
		Immediate:  false,
	}
	rabbitMQ := RabbitMQ{
		QueueParams:   queueParams,
		ConsumeParams: consumeParams,
		PublishParams: publishParams,
	}
	jobqueue := JobQueue{
		Option:         "rabbitmq",
		MemoryJobQueue: memoryJobQueue,
		RabbitMQ:       rabbitMQ,
	}
	wp := WorkerPool{
		Concurrency: 4,
		Backlog:     8,
	}
	scheduler := Scheduler{
		RepositoryPollingInterval: 70,
	}
	consumer := Consumer{
		JobQueuePollingInterval: 2,
	}
	mysql := MySQL{
		DSN:                   "test_dsn",
		CaPemFile:             "",
		ConnectionMaxLifetime: 1000,
		MaxIdleConnections:    8,
		MaxOpenConnections:    8,
	}
	repository := Repository{
		Option: "mysql",
		MySQL:  mysql,
	}
	config := &Config{
		Server:            server,
		JobQueue:          jobqueue,
		WorkerPool:        wp,
		Scheduler:         scheduler,
		Consumer:          consumer,
		Repository:        repository,
		LoggingFormat:     "text",
		TimeoutUnitOption: "second",
		TimeoutUnit:       time.Second,
	}
	tests := []struct {
		name     string
		dsn      string
		filepath string
		expected *Config
		err      error
	}{
		{
			"full",
			"test_dsn",
			"./testdata/test_config.yaml",
			config,
			nil,
		},
		{
			"no dsn",
			"",
			"./testdata/test_config.yaml",
			nil,
			errors.New("MySQL DSN not provided"),
		},
		{
			"wrong logging format",
			"test_dsn",
			"./testdata/test_config_invalid_logging_format.yaml",
			nil,
			errors.New("binary is not a valid logging_format option, valid options: map[json:true text:true]"),
		},
		{
			"wrong repository option",
			"test_dsn",
			"./testdata/test_config_invalid_repository_option.yaml",
			nil,
			errors.New("storage is not a valid repository option, valid options: map[memory:true mysql:true]"),
		},
		{
			"wrong timeout unit",
			"test_dsn",
			"./testdata/test_config_invalid_timeout_unit.yaml",
			nil,
			errors.New("year is not a valid timeout_unit option, valid options: map[millisecond:1ms second:1s]"),
		},
		{
			"wrong job queue option",
			"test_dsn",
			"./testdata/test_config_invalid_job_queue_option.yaml",
			nil,
			errors.New("queueX is not a valid job queue option, valid options: map[memory:true rabbitmq:true]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MYSQL_DSN", tt.dsn)
			filepath, _ := filepath.Abs(tt.filepath)
			cfg := new(Config)
			err := cfg.Load(filepath)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("load returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if eq := reflect.DeepEqual(cfg, tt.expected); !eq {
					t.Errorf("load set wrong config: got %#v want %#v", cfg, tt.expected)
				}
			}
		})
	}
}
