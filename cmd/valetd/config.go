package main

import (
	"io/ioutil"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                  string `yaml:"port"`
	JobQueueCapacity      int    `yaml:"job_queue_capacity"`
	WorkerPoolConcurrency int    `yaml:"worker_pool_concurrency"`
	WorkerPoolBacklog     int    `yaml:"worker_pool_backlog"`
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
		cfg.WorkerPoolConcurrency = runtime.NumCPU() / 2
	}
	if cfg.WorkerPoolBacklog == 0 {
		// By default allow a request spike double the worker capacity
		cfg.WorkerPoolBacklog = cfg.WorkerPoolConcurrency * 2
	}
	return nil
}
