# Valet
[![CI](https://github.com/svaloumas/valet/actions/workflows/ci.yml/badge.svg)](https://github.com/svaloumas/valet/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-79.5%25-brightgreen)

Stateless Go server responsible for executing tasks asynchronously and concurrently.

* [Overview](#overview)
* [Architecture](#architecture)
* [Installation](#installation)
* [Configuration](#configuration)
* [Secrets](#secrets)
* [Usage](#usage)
* [Tests](#tests)

<a name="overview"/>

## Overview

At its core, `valet` is a simple asynchronous task executor and scheduler. A task is a user-defined `func` that is executed as a callback by the service.
The implementation uses the notion of `job`, which describes the work that needs to be done and carries information about the task that will run for the specific job.
User-defined tasks are assigned to jobs. Every job can be assigned with a different task, a JSON payload with the data required for the task to be executed,
and an optional timeout interval. Jobs can be scheduled to run at a specified time or immediately.

After the tasks have been executed, their results along with the errors (if any) are stored into a repository.

<a name="architecture"/>

## Architecture

The project strives to follow the hexagonal architecture design pattern and to support modularity and extendability.
It provides the following interfaces and can be configured accordingly:

#### API

* HTTP
* gRPC

#### Repository

* In memory key-value storage.
* MySQL
* Redis

#### Message queue

* In memory job queue.
* RabbitMQ

<a name="installation"/>

## Installation

1. Clone the repo.

```bash
git clone https://github.com/svaloumas/valet.git
```

2. Download and install [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/).

3. Build and run the containers.

```bash
docker-compose up -d --build
```

<a name="configuration"/>

## Configuration

All configuration is set through `config.yaml`, which lives under the project's root directory.

```yaml
# Server config section
server:
  protocol: grpc                    # string - options: http, grpc
  http:
    port: 8080                      # string
  grpc:
    port: 50051                     # string
# Job queue config section
job_queue:
  option: rabbitmq                  # string - options: memory, rabbitmq
  memory_job_queue:
    capacity: 100                   # int
  rabbitmq:
    queue_params:
      queue_name: job               # string
      durable: false                # boolean
      deleted_when_unused: false    # boolean
      exclusive: false              # boolean
      no_wait: false                # boolean
    consume_params:
      name: rabbitmq-consumer       # string
      auto_ack: true                # boolean
      exclusive: false              # boolean
      no_local: false               # boolean
      no_wait: false                # boolean
    publish_params:
      exchange:                     # string
      routing_key: job              # string
      mandatory: false              # boolean
      immediate: false              # boolean
# Workerpool config section
worker_pool:
  concurrency:                      # int
  backlog:                          # int
# Scheduler config section
scheduler:
  repository_polling_interval: 60   # int
# Consumer config section
consumer:
  job_queue_polling_interval: 5     # int
# Repository config section
repository:
  option: memory                    # string - options:  memory, mysql
  mysql:
    connection_max_lifetime:        # int
    max_idle_connections:           # int
    max_open_connections:           # int
  redis:
    key_prefix: valet               # string
    min_idle_conns: 10              # int
    pool_size: 10                   # int
# Global config section
timeout_unit: second                # string - options: millisecond, second
logging_format: text                # string - options: text, json
```

<a name="secrets"/>

## Secrets

Currently, the only secrets would be the MySQL DSN and the RabbitMQ URI.
MySQL DSN can be provided as an environment variable named as `MYSQL_DSN`, or a Docker secret named as `valet-mysql-dsn`.
RabbitMQ URI can be provided as an environment variable named as `RABBITMQ_URI`, or a Docker secret named as `valet-rabbitmq-uri`.

<a name="usage"/>

## Usage

Define your own tasks under `task/` directory. The tasks should implement the `task.TaskFunc` type.

```go
// DummyParams is an example of a task params structure.
type DummyParams struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(taskParams interface{}) (interface{}, error) {
	params := &DummyParams{}
	mapstructure.Decode(taskParams, params)

        // Do something with the task params you injected through the API
        // ...
	metadata, err := downloadContent(params.URL)
        if err != nil {
            return nil, err
        }
	return metadata, nil
}

```

Register your new task callback in `main` function living in `cmd/valetd/main.go` and provide a name for it.

```go
taskrepo.Register("dummytask", task.DummyTask)
```

Create a new job by making an POST HTTP call to `/jobs`. You can inject arbitrary parameters for your task to run
by including them in the request body.

```json
{
    "name": "a job",
    "description": "what this job is all about, but briefly",
    "task_name": "dummytask",
    "task_params": {
        "url": "www.some-url.com"
    }
}
```

To schedule a new job to run at a specific time in the future, add `run_at` field to the request body.

```json
{
    "name": "a scheduled job",
    "description": "what this scheduled job is all about, but briefly",
    "task_name": "dummytask",
    "run_at": "2022-06-06T15:04:05.999",
    "task_params": {
        "url": "www.some-url.com"
    }
}
```

<a name="tests"/>

## Tests

Run the complete test suite.

```bash
docker-compose up -d mysql rabbitmq
make test
```