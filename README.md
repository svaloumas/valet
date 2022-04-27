# Valet
[![CI](https://github.com/svaloumas/valet/actions/workflows/ci.yml/badge.svg)](https://github.com/svaloumas/valet/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/svaloumas/valet/branch/main/graph/badge.svg?token=9CI4Q74JJK)](https://codecov.io/gh/svaloumas/valet)
[![Go Report Card](https://goreportcard.com/badge/github.com/svaloumas/valet)](https://goreportcard.com/report/github.com/svaloumas/valet)
[![Go Reference](https://pkg.go.dev/badge/github.com/svaloumas/valet.svg)](https://pkg.go.dev/github.com/svaloumas/valet)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/svaloumas/valet/blob/develop/LICENSE)

Stateless Go server responsible for executing tasks asynchronously and concurrently.

* [Overview](#overview)
  * [Job](#job)
  * [Pipeline](#pipeline)
* [Architecture](#architecture)
* [Installation](#installation)
* [Configuration](#configuration)
* [Secrets](#secrets)
* [Usage](#usage)
* [Tests](#tests)

<a name="overview"/>

## Overview

At its core, `valet` is a simple job queuing system and an asynchronous task executor. A task is a user-defined `func` that is run as a callback by the service.

<a name="job"/>

### Job

The implementation uses the notion of `job`, which describes the work that needs to be done and carries information about the task that will run for the
specific job. User-defined tasks are assigned to jobs. Every job can be assigned with a different task, a JSON payload with the data required for the task
to be executed, and an optional timeout interval. Jobs can be scheduled to run at a specified time or immediately.

After the tasks have been executed, their results along with the errors (if any) are stored into a repository.

<a name="pipeline"/>

### Pipeline

A `pipeline` is a sequence of jobs that need to be executed in a specified order, one by one. Every job in the pipeline can be assigned with a different task
and parameters, and each task callback can optionally use the results of the previous task in the job sequence. A pipeline can also be scheduled to be executed
sometime in the future, or run immediately.

<a name="architecture"/>

## Architecture

Your callback functions can live in any repo and should be registered to your own build of `valet`. To unlock this level of flexibility,
the service is provided as a Go `pkg` rather than a `cmd`, enabling the task registration before building the executable.

Internally, the service consists of the following components.

* Server - Exposes a RESTful API to enable communication with the outer world.
* Job queue - A FIFO queue that supports the job queuing mechanism of the service.
* Scheduler - Responsible for dispatching the jobs from the job queue to the worker pool and of course for scheduling the tasks for future execution.
* Worker pool - A number of available go-routines, responsible for the concurrent execution of the jobs.
* Repository - The storage system where jobs and their results persist.

The design strives to follow the hexagonal architecture pattern and to support modularity and extendability. 

So far, `valet` provides the following interfaces and can be configured accordingly to function with any of the listed technologies.

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

2. Define your own task functions in your repo by implementing the type `func(...interface{}) (interface{}, error)`.

```go
package somepkg

import (
	"github.com/svaloumas/valet"
)

// DummyParams is an example of a task params structure.
type DummyParams struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(args ...interface{}) (interface{}, error) {
	dummyParams := &DummyParams{}
	var previousResultsMetadata string
	valet.DecodeTaskParams(args, dummyParams)
	valet.DecodePreviousJobResults(args, &resultsMetadata) // Applies for pipelines.

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

> `args[0]` holds the task parameters as they were given through the API call for the job/pipeline creation. `args[1]` holds the results of the 
> previous task only in case of a pipeline execution. Prefer to safely access the arguments by using `valet.DecodeTaskParams` and `valet.DecodePreviousJobResults`
> to decode them into your custom task param structs.

3. Copy `config.yaml` from the repo and set a configuration according to your needs.

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
  job_queue_polling_interval: 5     # int
  repository_polling_interval: 60   # int
# Repository config section
repository:
  option: memory                    # string - options:  memory, mysql, postgres, redis
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

Currently, secrets are the MySQL DSN, PostgreSQL DSN, Redis URL and the RabbitMQ URI. Each can be provided as an environment variable. 
Alternatively, if you choose to use the provided Docker compose files, you can create the corresponding Docker secrets.

* Env var `MYSQL_DSN`, Docker secret name `valet-mysql-dsn`.
* Env var `POSTGRES_DSN`, Docker secret name `valet-postgres-dsn`.
* Env var `REDIS_DSN`, Docker secret name `valet-redis-url`.
* End var `RABBITMQ_URI`, Docker secret name `valet-rabbitmq-uri`.

<a name="usage"/>

## Usage

Create a new job by making a POST HTTP call to `/jobs` or via gRPC to `job.Job.Create` service method. You can inject arbitrary parameters for your task to run
by including them in the request body.

```json
{
    "name": "a job",
    "description": "what this job is all about, but briefly",
    "task_name": "dummytask",
    "task_params": {
        "url": "www.some-fake-url.com"
    },
    "timeout": 10
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
        "url": "www.some-fake-url.com"
    },
    "timeout": 10
}
```

Create a new pipeline by making a POST HTTP call to `/pipelines` or via gRPC to `pipeline.Pipeline.Create` service method. You can inject arbitrary parameters
for your tasks to run by including them in the request body. Optionally, you can tune your tasks to use any results of the previous task in the pipeline, creating
a `bash`-like command pipeline. Pipelines can also be scheduled for execution in some time in the future, by adding `run_at` field to the request payload
just like with the jobs.

```json
{
    "name": "a scheduled pipeline",
    "description": "what this pipeline is all about",
    "run_at": "2022-06-06T15:04:05.999",
    "jobs": [
        {
            "name": "the first job",
            "description": "some job description",
            "task_name": "dummytask",
            "task_params": {
                "url": "www.some-fake-url.com"
            }
        },
        {
            "name": "a second job",
            "description": "some job description",
            "task_name": "anothertask",
            "task_params": {
                "url": "www.some-fake-url.com"
            },
            "use_previous_results": true,
            "timeout": 10
        },
        {
            "name": "the last job",
            "description": "some job description",
            "task_name": "dummytask",
            "task_params": {
                "url": "www.some-fake-url.com"
            },
            "use_previous_results": true
        }
    ]
}
```

<a name="tests"/>

## Tests

Run the complete test suite.

```bash
docker-compose up -d
make test
```