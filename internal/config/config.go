package config

import (
	"errors"
	"fmt"
	"io/ioutil"
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
		"redis":  true,
	}
	validJobQueueOptions = map[string]bool{
		"memory":   true,
		"rabbitmq": true,
	}
	validProtocolOptions = map[string]bool{
		"http": true,
		"grpc": true,
	}
)

type HTTP struct {
	Port string `yaml:"port"`
}

type GRPC struct {
	Port string `yaml:"port"`
}

type Server struct {
	Protocol string `yaml:"protocol"`
	HTTP     HTTP   `yaml:"http"`
	GRPC     GRPC   `yaml:"grpc"`
}

type MemoryJobQueue struct {
	Capacity int `yaml:"capacity"`
}

type QueueParams struct {
	Name              string `yaml:"name"`
	Durable           bool   `yaml:"durable"`
	DeletedWhenUnused bool   `yaml:"deleted_when_unused"`
	Exclusive         bool   `yaml:"exclusive"`
	NoWait            bool   `yaml:"no_wait"`
}

type ConsumeParams struct {
	Name      string `yaml:"name"`
	AutoACK   bool   `yaml:"auto_ack"`
	Exclusive bool   `yaml:"exclusive"`
	NoLocal   bool   `yaml:"no_local"`
	NoWait    bool   `yaml:"no_wait"`
}

type PublishParams struct {
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routing_key"`
	Mandatory  bool   `yaml:"mandatory"`
	Immediate  bool   `yaml:"immediate"`
}

type RabbitMQ struct {
	QueueParams   QueueParams   `yaml:"queue_params"`
	ConsumeParams ConsumeParams `yaml:"consume_params"`
	PublishParams PublishParams `yaml:"publish_params"`
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
	Redis  Redis  `yaml:"redis"`
}

type MySQL struct {
	DSN                   string
	CaPemFile             string
	ConnectionMaxLifetime int `yaml:"connection_max_lifetime"`
	MaxIdleConnections    int `yaml:"max_idle_connections"`
	MaxOpenConnections    int `yaml:"max_open_connections"`
}

type Redis struct {
	URL          string
	KeyPrefix    string `yaml:"key_prefix"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type Config struct {
	Server            Server     `yaml:"server"`
	JobQueue          JobQueue   `yaml:"job_queue"`
	WorkerPool        WorkerPool `yaml:"worker_pool"`
	Scheduler         Scheduler  `yaml:"scheduler"`
	Consumer          Consumer   `yaml:"consumer"`
	Repository        Repository `yaml:"repository"`
	TimeoutUnitOption string     `yaml:"timeout_unit"`
	LoggingFormat     string     `yaml:"logging_format"`
	TimeoutUnit       time.Duration
}

func (cfg *Config) Load(filepath string) error {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return err
	}
	err = cfg.setServerConfig()
	if err != nil {
		return err
	}
	err = cfg.setJobQueueConfig()
	if err != nil {
		return err
	}
	cfg.setWorkerPoolConfig()
	err = cfg.setTimeoutUnitConfig()
	if err != nil {
		return err
	}
	cfg.setSchedulerConfig()
	cfg.setConsumerConfig()
	err = cfg.setLoggingFormatConfig()
	if err != nil {
		return err
	}
	err = cfg.setRepositoryConfig()
	if err != nil {
		return err
	}
	return nil
}

func (cfg *Config) setServerConfig() error {
	if _, ok := validProtocolOptions[cfg.Server.Protocol]; !ok {
		return fmt.Errorf("%s is not a valid protocol option, valid options: %v", cfg.Server.Protocol, validProtocolOptions)
	}
	if cfg.Server.HTTP.Port == "" {
		cfg.Server.HTTP.Port = "8080"
	}
	if cfg.Server.GRPC.Port == "" {
		cfg.Server.GRPC.Port = "50051"
	}
	return nil
}

func (cfg *Config) setJobQueueConfig() error {
	if _, ok := validJobQueueOptions[cfg.JobQueue.Option]; !ok {
		return fmt.Errorf("%s is not a valid job queue option, valid options: %v", cfg.JobQueue.Option, validJobQueueOptions)
	}
	if cfg.JobQueue.Option == "memory" {
		if cfg.JobQueue.MemoryJobQueue.Capacity == 0 {
			cfg.JobQueue.MemoryJobQueue.Capacity = 100
		}
	}
	return nil
}

func (cfg *Config) setWorkerPoolConfig() {
	if cfg.WorkerPool.Concurrency == 0 {
		// Work is CPU bound so number of cores should be fine.
		cfg.WorkerPool.Concurrency = runtime.NumCPU()
	}
	if cfg.WorkerPool.Backlog == 0 {
		// By default allow a request spike double the worker capacity
		cfg.WorkerPool.Backlog = cfg.WorkerPool.Concurrency * 2
	}
}

func (cfg *Config) setTimeoutUnitConfig() error {
	timeoutUnit, ok := validTimeoutUnitOptions[cfg.TimeoutUnitOption]
	if !ok {
		return fmt.Errorf("%s is not a valid timeout_unit option, valid options: %v", cfg.TimeoutUnitOption, validTimeoutUnitOptions)
	}
	cfg.TimeoutUnit = timeoutUnit
	return nil
}

func (cfg *Config) setSchedulerConfig() {
	if cfg.Scheduler.RepositoryPollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.Scheduler.RepositoryPollingInterval = 60
		} else {
			cfg.Scheduler.RepositoryPollingInterval = 60000
		}
	}
}

func (cfg *Config) setConsumerConfig() {
	if cfg.Consumer.JobQueuePollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.Consumer.JobQueuePollingInterval = 1
		} else {
			cfg.Consumer.JobQueuePollingInterval = 1000
		}
	}
}

func (cfg *Config) setLoggingFormatConfig() error {
	if _, ok := validLoggingFormatOptions[cfg.LoggingFormat]; !ok {
		return fmt.Errorf("%s is not a valid logging_format option, valid options: %v", cfg.LoggingFormat, validLoggingFormatOptions)
	}
	return nil
}

func (cfg *Config) setRepositoryConfig() error {
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
	if cfg.Repository.Option == "redis" {
		url := env.LoadVar("REDIS_URL")
		if url == "" {
			return errors.New("Redis URL not provided")
		}
		cfg.Repository.Redis.URL = url

		if cfg.Repository.Redis.PoolSize == 0 {
			cfg.Repository.Redis.PoolSize = 10
		}
		if cfg.Repository.Redis.MinIdleConns == 0 {
			cfg.Repository.Redis.MinIdleConns = 10
		}
	}
	return nil
}
