package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	http := HTTP{
		Port: "8080",
	}
	grpc := GRPC{
		Port: "50051",
	}
	server := Server{
		Protocol: "http",
		HTTP:     http,
		GRPC:     grpc,
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
	redis := Redis{
		KeyPrefix:    "somekey",
		PoolSize:     10,
		MinIdleConns: 10,
	}
	jobqueue := JobQueue{
		Option:         "rabbitmq",
		MemoryJobQueue: memoryJobQueue,
		RabbitMQ:       rabbitMQ,
		Redis:          redis,
	}
	wp := WorkerPool{
		Workers:       4,
		QueueCapacity: 8,
	}
	scheduler := Scheduler{
		RepositoryPollingInterval: 70,
		JobQueuePollingInterval:   2,
	}
	mysql := MySQL{
		DSN:                   "test_dsn",
		CaPemFile:             "",
		ConnectionMaxLifetime: 1000,
		MaxIdleConnections:    8,
		MaxOpenConnections:    8,
	}
	postgres := Postgres{
		DSN:                   "",
		ConnectionMaxLifetime: 1000,
		MaxIdleConnections:    8,
		MaxOpenConnections:    8,
	}
	repository := Repository{
		Option:   "mysql",
		MySQL:    mysql,
		Redis:    redis,
		Postgres: postgres,
	}
	config := &Config{
		Server:            server,
		JobQueue:          jobqueue,
		WorkerPool:        wp,
		Scheduler:         scheduler,
		Repository:        repository,
		LoggingFormat:     "text",
		TimeoutUnitOption: "second",
		TimeoutUnit:       time.Second,
	}
	tests := []struct {
		name        string
		mysqlDSN    string
		postgresDSN string
		redisURL    string
		filepath    string
		expected    *Config
		err         error
	}{
		{
			"full",
			"test_dsn",
			"",
			"",
			"./testdata/test_config.yaml",
			config,
			nil,
		},
		{
			"no mysql dsn",
			"",
			"test_dsn",
			"test_url",
			"./testdata/test_config.yaml",
			nil,
			errors.New("MySQL DSN not provided"),
		},
		{
			"no postgres dsn",
			"test_dsn",
			"",
			"test_url",
			"./testdata/test_config_postgres.yaml",
			nil,
			errors.New("PostgreSQL DSN not provided"),
		},
		{
			"no redis url",
			"test_dsn",
			"test_dsn",
			"",
			"./testdata/test_config_redis.yaml",
			nil,
			errors.New("Redis URL not provided"),
		},
		{
			"wrong logging format",
			"test_dsn",
			"test_dsn",
			"test_url",
			"./testdata/test_config_invalid_logging_format.yaml",
			nil,
			errors.New("binary is not a valid logging_format option, valid options: map[json:true text:true]"),
		},
		{
			"wrong repository option",
			"test_dsn",
			"test_dsn",
			"test_url",
			"./testdata/test_config_invalid_repository_option.yaml",
			nil,
			errors.New("storage is not a valid repository option, valid options: map[memory:true mysql:true postgres:true redis:true]"),
		},
		{
			"wrong timeout unit",
			"test_dsn",
			"test_dsn",
			"test_url",
			"./testdata/test_config_invalid_timeout_unit.yaml",
			nil,
			errors.New("year is not a valid timeout_unit option, valid options: map[millisecond:1ms second:1s]"),
		},
		{
			"wrong job queue option",
			"test_dsn",
			"test_dsn",
			"test_url",
			"./testdata/test_config_invalid_job_queue_option.yaml",
			nil,
			errors.New("queueX is not a valid job queue option, valid options: map[memory:true rabbitmq:true redis:true]"),
		},
		{
			"wrong protovol option",
			"test_dsn",
			"test_dsn",
			"test_url",
			"./testdata/test_config_invalid_protocol_option.yaml",
			nil,
			errors.New("websockets is not a valid protocol option, valid options: map[grpc:true http:true]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MYSQL_DSN", tt.mysqlDSN)
			os.Setenv("POSTGRES_DSN", tt.postgresDSN)
			os.Setenv("REDIS_URL", tt.redisURL)
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

func TestLoadDefaultValues(t *testing.T) {
	http := HTTP{
		Port: "8080",
	}
	grpc := GRPC{
		Port: "50051",
	}
	server := Server{
		Protocol: "http",
		HTTP:     http,
		GRPC:     grpc,
	}
	memoryJobQueue := MemoryJobQueue{
		Capacity: 100,
	}
	redis := Redis{
		KeyPrefix: "",
	}
	jobqueue := JobQueue{
		Option:         "memory",
		MemoryJobQueue: memoryJobQueue,
		Redis:          redis,
	}
	wp := WorkerPool{
		Workers:       runtime.NumCPU(),
		QueueCapacity: runtime.NumCPU() * 2,
	}
	scheduler := Scheduler{
		RepositoryPollingInterval: 60,
		JobQueuePollingInterval:   1,
	}
	mysql := MySQL{
		DSN:                   "test_dsn",
		CaPemFile:             "",
		ConnectionMaxLifetime: 3000,
		MaxIdleConnections:    8,
		MaxOpenConnections:    8,
	}
	postgres := Postgres{
		DSN:                   "test_dsn",
		ConnectionMaxLifetime: 3000,
		MaxIdleConnections:    8,
		MaxOpenConnections:    8,
	}
	repository := Repository{
		Option:   "mysql",
		MySQL:    mysql,
		Redis:    redis,
		Postgres: postgres,
	}
	config := &Config{
		Server:            server,
		JobQueue:          jobqueue,
		WorkerPool:        wp,
		Scheduler:         scheduler,
		Repository:        repository,
		LoggingFormat:     "text",
		TimeoutUnitOption: "second",
		TimeoutUnit:       time.Second,
	}
	tests := []struct {
		name        string
		mysqlDSN    string
		postgresDSN string
		redisURL    string
		filepath    string
		expected    *Config
		err         error
	}{
		{
			"http mysql second",
			"test_dsn",
			"",
			"",
			"./testdata/test_config_defaults_http_mysql_second.yaml",
			config,
			nil,
		},
		{
			"grpc postgres second",
			"",
			"test_dsn",
			"",
			"./testdata/test_config_defaults_grpc_postgres_second.yaml",
			config,
			nil,
		},
		{
			"redis millisecond",
			"",
			"",
			"test_url",
			"./testdata/test_config_defaults_grpc_redis_millisecond.yaml",
			config,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MYSQL_DSN", tt.mysqlDSN)
			os.Setenv("POSTGRES_DSN", tt.postgresDSN)
			os.Setenv("REDIS_URL", tt.redisURL)
			filepath, _ := filepath.Abs(tt.filepath)
			cfg := new(Config)
			err := cfg.Load(filepath)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("load returned wrong error: got %#v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				if strings.Contains(tt.name, "postgres") {
					tt.expected.Repository.Option = "postgres"
					tt.expected.Repository.MySQL.DSN = ""
					tt.expected.Repository.Postgres.DSN = tt.postgresDSN
				}
				if strings.Contains(tt.name, "mysql") {
					tt.expected.Repository.Option = "mysql"
					tt.expected.Repository.Postgres.DSN = ""
					tt.expected.Repository.MySQL.DSN = tt.mysqlDSN
				}
				if strings.Contains(tt.name, "grpc") {
					tt.expected.Server.Protocol = "grpc"
				}
				if strings.Contains(tt.name, "redis") {
					tt.expected.Repository.Option = "redis"
					tt.expected.Repository.MySQL.DSN = ""
					tt.expected.Repository.Postgres.DSN = ""
					tt.expected.Repository.Redis.URL = tt.redisURL
					tt.expected.JobQueue.Option = "redis"
					tt.expected.JobQueue.Redis.URL = tt.redisURL
					tt.expected.JobQueue.Redis.MinIdleConns = 10
					tt.expected.JobQueue.Redis.PoolSize = 10
					tt.expected.Repository.Redis.MinIdleConns = 10
					tt.expected.Repository.Redis.PoolSize = 10
					tt.expected.JobQueue.MemoryJobQueue.Capacity = 0
				}
				if strings.Contains(tt.name, "millisecond") {
					tt.expected.Scheduler.RepositoryPollingInterval = 60000
					tt.expected.Scheduler.JobQueuePollingInterval = 1000
					tt.expected.TimeoutUnitOption = "millisecond"
					tt.expected.TimeoutUnit = time.Millisecond
				}
				if eq := reflect.DeepEqual(cfg, tt.expected); !eq {
					t.Errorf("load set wrong config: got %#v want %#v", cfg, tt.expected)
				}
			}
		})
	}
}
