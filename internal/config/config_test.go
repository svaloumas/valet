package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	os.Setenv("MYSQL_DSN", "test_dsn")
	memoryJobQueue := MemoryJobQueue{
		Capacity: 100,
	}
	rabbitMQ := RabbitMQ{
		QueueName:         "test",
		Durable:           true,
		DeletedWhenUnused: true,
		Exclusive:         true,
		NoWait:            false,
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
		Port:              "8080",
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
		filepath string
		expected *Config
		err      error
	}{
		{
			"full",
			"./testdata/test_config.yaml",
			config,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
