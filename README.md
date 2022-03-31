# Valet service
[![CI](https://github.com/svaloumas/valet/actions/workflows/ci.yml/badge.svg)](https://github.com/svaloumas/valet/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-93.7%25-brightgreen)

Simple stateless Go server responsible for executing tasks, referred as jobs.

* [Overview](#overview)
* [Installation](#installation)
* [Configuration](#configuration)
* [Usage](#usage)
* [Tests](#tests)

<a name="overview"/>

## Overview

At its core, `valet` is an asynchronous task executor.

The user can define callbacks to be executed by the service and assign them to jobs. Every job can be assigned with a different user defined
callback, a JSON payload with the data required for the callback to be executed, and an optional timeout interval.

The service exposes a JSON RestAPI providing CRUD endpoints for the job management. Configuration uses a single `yaml` file living under the root
directory of the project.

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

<a name="usage"/>

## Usage

Define your own tasks under `task/` directory. The tasks should implement the `task.TaskFunc` type.

```go
// DummyMetadata is an example of a task metadata structure.
type DummyMetadata struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy task callback.
func DummyTask(metadata interface{}) (interface{}, error) {
	taskMetadata := &DummyMetadata{}
	mapstructure.Decode(metadata, taskMetadata)

        // Do something with the metadata you injected through the API
        // ...
	taskMetadata.URL = "http://www.test-url.com"
	return taskMetadata, nil
}

```

Register your new task callback in `main` function living in `cmd/valetd/main.go` and provide a name for it.

```go
taskrepo.Register("dummytask", task.DummyTask)
```

Create a new job by making an POST HTTP call to `/jobs`. You can inject any arbitrary metadata for your task to run
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

<a name="tests"/>

## Tests

Run the complete test suite.

```bash
docker-compose run --rm valet make test
```