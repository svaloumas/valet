package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"

	"valet/pkg/env"
)

var (
	validTimeoutUnitOptions = map[string]time.Duration{
		"second":      time.Second,
		"millisecond": time.Millisecond,
	}
	validLoggingFormatOptions = map[string]bool{
		"text": true,
		"json": true,
	}
	validRepositoryOptions = map[string]bool{
		"memory": true,
		"mysql":  true,
	}
	validJobQueueOptions = map[string]bool{
		"memory":   true,
		"rabbitmq": true,
	}
)

type MemoryJobQueue struct {
	Capacity int `yaml:"capacity"`
}

type RabbitMQ struct {
	QueueName         string `yaml:"queue_name"`
	Durable           bool   `yaml:"durable"`
	DeletedWhenUnused bool   `yaml:"deleted_when_unused"`
	Exclusive         bool   `yaml:"exclusive"`
	NoWait            bool   `yaml:"nowait"`
}

type JobQueue struct {
	Option         string         `yaml:"option"`
	MemoryJobQueue MemoryJobQueue `yaml:"memory_job_queue"`
	RabbitMQ       RabbitMQ       `yaml:"rabbitmq"`
}

type WorkerPool struct {
	Concurrency int `yaml:"concurrency"`
	Backlog     int `yaml:"backlog"`
}

type Scheduler struct {
	RepositoryPollingInterval int `yaml:"repository_polling_interval"`
}

type Consumer struct {
	JobQueuePollingInterval int `yaml:"job_queue_polling_interval"`
}

type Repository struct {
	Option string `yaml:"option"`
	MySQL  MySQL  `yaml:"mysql"`
}

type MySQL struct {
	DSN                   string
	CaPemFile             string
	ConnectionMaxLifetime int `yaml:"connection_max_lifetime"`
	MaxIdleConnections    int `yaml:"max_idle_connections"`
	MaxOpenConnections    int `yaml:"max_open_connections"`
}

type Config struct {
	Port                string     `yaml:"port"`
	JobQueue            JobQueue   `yaml:"job_queue"`
	WorkerPool          WorkerPool `yaml:"worker_pool"`
	Scheduler           Scheduler  `yaml:"scheduler"`
	Consumer            Consumer   `yaml:"consumer"`
	Repository          Repository `yaml:"repository"`
	TimeoutUnitOption   string     `yaml:"timeout_unit"`
	LoggingFormatOption string     `yaml:"logging_format"`
	TimeoutUnit         time.Duration
	LoggingFormat       string
}

func (cfg *Config) Load() error {
	filename, _ := filepath.Abs("config.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return err
	}

	// Server config
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Job queue config
	if _, ok := validJobQueueOptions[cfg.JobQueue.Option]; !ok {
		return fmt.Errorf("%s is not a valid job queue option, valid options: %v", cfg.JobQueue.Option, validJobQueueOptions)
	}
	if cfg.JobQueue.Option == "memory" {
		if cfg.JobQueue.MemoryJobQueue.Capacity == 0 {
			cfg.JobQueue.MemoryJobQueue.Capacity = 100
		}
	}

	// Workerpool config
	if cfg.WorkerPool.Concurrency == 0 {
		// Work is CPU bound so number of cores should be fine.
		cfg.WorkerPool.Concurrency = runtime.NumCPU()
	}
	if cfg.WorkerPool.Backlog == 0 {
		// By default allow a request spike double the worker capacity
		cfg.WorkerPool.Backlog = cfg.WorkerPool.Concurrency * 2
	}

	// Timeout unit config
	timeoutUnit, ok := validTimeoutUnitOptions[cfg.TimeoutUnitOption]
	if !ok {
		return fmt.Errorf("%s is not a valid timeout_unit option, valid options: %v", cfg.TimeoutUnitOption, validTimeoutUnitOptions)
	}
	cfg.TimeoutUnit = timeoutUnit

	// Scheduler config
	if cfg.Scheduler.RepositoryPollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.Scheduler.RepositoryPollingInterval = 60
		} else {
			cfg.Scheduler.RepositoryPollingInterval = 60000
		}
	}

	// Consumer config
	if cfg.Consumer.JobQueuePollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.Consumer.JobQueuePollingInterval = 1
		} else {
			cfg.Consumer.JobQueuePollingInterval = 1000
		}
	}

	// Logging format config
	if _, ok := validLoggingFormatOptions[cfg.LoggingFormatOption]; !ok {
		return fmt.Errorf("%s is not a valid logging_format option, valid options: %v", cfg.LoggingFormatOption, validLoggingFormatOptions)
	}
	cfg.LoggingFormat = cfg.LoggingFormatOption

	// Repository config
	if _, ok := validRepositoryOptions[cfg.Repository.Option]; !ok {
		return fmt.Errorf("%s is not a valid repository option, valid options: %v", cfg.Repository.Option, validRepositoryOptions)
	}
	if cfg.Repository.Option == "mysql" {
		dsn := env.LoadVar("MYSQL_DSN")
		if dsn == "" {
			return errors.New("MySQL DSN not provided")
		}
		caPemFile := env.LoadVar("MYSQL_CA_PEM_FILE")
		cfg.Repository.MySQL.DSN = dsn
		cfg.Repository.MySQL.CaPemFile = caPemFile

		if cfg.Repository.MySQL.ConnectionMaxLifetime == 0 {
			cfg.Repository.MySQL.ConnectionMaxLifetime = 3000
		}
		if cfg.Repository.MySQL.MaxIdleConnections == 0 {
			cfg.Repository.MySQL.MaxIdleConnections = 8
		}
		if cfg.Repository.MySQL.MaxOpenConnections == 0 {
			cfg.Repository.MySQL.MaxOpenConnections = 8
		}
	}
	return nil
}
