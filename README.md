# Valet
[![CI](https://github.com/svaloumas/valet/actions/workflows/ci.yml/badge.svg)](https://github.com/svaloumas/valet/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-90.3%25-brightgreen)

Stateless Go server responsible for executing tasks asynchronously and concurrently.

* [Overview](#overview)
* [Architecture](#architecture)
* [Installation](#installation)
* [Configuration](#configuration)
* [Usage](#usage)
* [Tests](#tests)

<a name="overview"/>

## Overview

At its core, `valet` is a simple asynchronous task executor and scheduler. A task is a user-defined `func` that is executed as a callback by the service.
The implementation uses the notion of `job`, which describes the work that needs to be done and carries information about the task that will run for the specific job.
User-defined tasks are assigned to jobs. Every job can be assigned with a different task, a JSON payload with the data required for the task to be executed,
and an optional timeout interval. Jobs can be scheduled to run at a specified time or immediately.

After the tasks have been executed, their results along with the errors (if any) are stored into a repository.

The service exposes a JSON RestAPI providing CRUD endpoints for the job resource management. Configuration uses a single `yaml` file living under the root
directory of the project.

<a name="architecture"/>

## Architecture

The project strives to follow the hexagonal architecture design pattern and to support modularity and extendability.
Currently, it provides the following interfaces and can be configured accordingly:

#### API

* HTTP
* gRPC (to be implemented)

#### Repository

* In memory key-value storage.
* MySQL (to be implemented)

#### Message queue

* In memory job queue.
* RabbitMQ (to be implemented)

<a name="installation"/>

## Installation

1. Clone the repo.

```bash
git clone https://github.com/svaloumas/valet.git
```

2. Download and install [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/).

3. Build and run the `valet` container.

```bash
docker-compose up --build -d
```

<a name="configuration"/>

## Configuration

All configuration is set through `config.yaml`, which lives in the project's root directory.

Available configuration options:

| Parameter                  | Type     | Default                     | Description                               |
| -------------------------- | -------- | --------------------------- | ----------------------------------------- |
| port                       | string   | 8080                        | The port that the server should listen to |
| job_queue_capacity         | integer  | 100                         | The capacity of the job queue (applies only for in-memory job queue)|
| worker_pool_concurrency    | integer  | number of CPU cores         | The number of go-routines responsible for executing the jobs concurrently |
| worker_pool_backlog        | integer  | worker_pool_concurrency * 2 | The capacity of the worker pool work queue |
| timeout_unit               | string   | -                           | The unit of time that will be used for the timeout interval specified for each job |
| scheduler_polling_interval | integer  | 60 seconds                  | The time interval in which the scheduler will poll for new events |
| job_queue_polling_timeout  | integer  | 1 second                    | The time interval in which the consumer will poll the queue for new jobs |
| logging_format             | string   | -                           | The logging format of the service, text and JSON are supported |

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
    "metadata": {
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
    "metadata": {
        "url": "www.some-url.com"
    }
}
```

<a name="tests"/>

## Tests

Run the complete test suite.

```bash
docker-compose run --rm valet make test
```