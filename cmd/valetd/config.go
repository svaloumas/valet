package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	validTimeoutUnitOptions = map[string]time.Duration{
		"second":      time.Second,
		"millisecond": time.Millisecond,
	}
)

type Config struct {
	Port                     string `yaml:"port"`
	JobQueueCapacity         int    `yaml:"job_queue_capacity"`
	WorkerPoolConcurrency    int    `yaml:"worker_pool_concurrency"`
	WorkerPoolBacklog        int    `yaml:"worker_pool_backlog"`
	SchedulerPollingInterval int    `yaml:"scheduler_polling_interval"`
	JobQueuePollingInterval  int    `yaml:"job_queue_polling_interval"`
	TimeoutUnitOption        string `yaml:"timeout_unit"`
	TimeoutUnit              time.Duration
	Env                      string `yaml:"env"`
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
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.JobQueueCapacity == 0 {
		cfg.JobQueueCapacity = 100
	}
	if cfg.WorkerPoolConcurrency == 0 {
		// Work is CPU bound so number of cores should be fine.
		cfg.WorkerPoolConcurrency = runtime.NumCPU()
	}
	if cfg.WorkerPoolBacklog == 0 {
		// By default allow a request spike double the worker capacity
		cfg.WorkerPoolBacklog = cfg.WorkerPoolConcurrency * 2
	}
	timeoutUnit, ok := validTimeoutUnitOptions[cfg.TimeoutUnitOption]
	if !ok {
		return fmt.Errorf("%s is not a valid timeout_unit option, valid options: %v", cfg.TimeoutUnitOption, validTimeoutUnitOptions)
	}
	cfg.TimeoutUnit = timeoutUnit
	if cfg.SchedulerPollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.SchedulerPollingInterval = 60
		} else {
			cfg.SchedulerPollingInterval = 60000
		}
	}
	if cfg.JobQueuePollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.JobQueuePollingInterval = 1
		} else {
			cfg.SchedulerPollingInterval = 1000
		}
	}
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	return nil
}
