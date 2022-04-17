# Valet
[![CI](https://github.com/svaloumas/valet/actions/workflows/ci.yml/badge.svg)](https://github.com/svaloumas/valet/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/svaloumas/valet/branch/main/graph/badge.svg?token=9CI4Q74JJK)](https://codecov.io/gh/svaloumas/valet)
[![Go Report Card](https://goreportcard.com/badge/github.com/svaloumas/valet)](https://goreportcard.com/report/github.com/svaloumas/valet)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/svaloumas/valet/blob/develop/LICENSE)

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

The service is provided as a `pkg`, to enable user-defined callbacks registration before building the executable.
Its design strives to follow the hexagonal architecture pattern and to support modularity and extendability.
`valet` provides the following interfaces and can be configured accordingly:

#### API

* HTTP
* gRPC

#### Repository

* In memory key-value storage.
* MySQL
* PostgreSQL
* Redis

#### Message queue

* In memory job queue.
* RabbitMQ

<a name="installation"/>

## Installation

1. Download the `pkg`.

```bash
go get github.com/svaloumas/valet
```

2. Define your own task callbacks in your repo by implementing the type `func(interface{}) (interface{}, error)`.

```go
package somepkg

import (
	"github.com/mitchellh/mapstructure"
)

// DummyParams is an example of a task params structure.
type DummyParams struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(taskParams interface{}) (interface{}, error) {
	params := &DummyParams{}
	mapstructure.Decode(taskParams, params)

	metadata, err := downloadContent(params.URL)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func downloadContent(URL string) (string, error) {
	return "some metadata", nil
}
```

3. Copy `config.yaml` from the repo and set your own configuration.

4. Initialize `valet` in a `main` function under your repo, register your tasks to the service and run it.

```bash
mkdir -p cmd/valetd/
touch cmd/valetd/main.go
```

```go
// cmd/valetd/main.go
package main

import (
	"github.com/svaloumas/valet"
	"path/to/somepkg"
)

func main() {
	v := valet.New("/path/to/config.yaml")
	v.RegisterTask("mytask", somepkg.DummyTask)
	v.Run()
}
```

5. Build and run the service.

    5.1. To run the service and its dependencies as Docker containers, use the `Dockerfile`, `docker-compose` and `Makefile` files provided.

    ```bash
    docker-compose up -d --build
    ```

    5.2. Build and run the service as a standalone binary.

    > Optionally set the corresponding environment variables depending on your configuration options.

    ```bash
    export POSTGRES_DSN=
    export RABBITMQ_URI=
    export MYSQL_DSN=
    export REDIS_URL=
    ```

    ```bash
    go build -o valetd cmd/valted/*.go
    ./valetd

    ```

<a name="configuration"/>

## Configuration

All configuration is set through `config.yaml`, which lives under the project's root directory.

```yaml
# Server config section
server:
  protocol: http                    # string - options: http, grpc
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
  postgres:
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

Currently, the only secrets would be the MySQL DSN, PostgreSQL DSN and the RabbitMQ URI.
MySQL DSN can be provided as an environment variable named as `MYSQL_DSN`, or a Docker secret named as `valet-mysql-dsn`.
PostgreSQL DSN can be provided as an environment variable named as `POSTGRES_DSN`, or a Docker secret named as `valet-postgres-dsn`.
RabbitMQ URI can be provided as an environment variable named as `RABBITMQ_URI`, or a Docker secret named as `valet-rabbitmq-uri`.

<a name="usage"/>

## Usage

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
docker-compose up -d
make test
```