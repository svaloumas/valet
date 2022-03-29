# Valet service
[![CI](https://github.com/svaloumas/valet/actions/workflows/ci.yml/badge.svg)](https://github.com/svaloumas/valet/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-95.9%25-brightgreen)

Simple stateless Go server responsible for executing tasks, referred as jobs.

## Overview

At its core, `valet` is an asynchronous task executor.<br>
The user can define callbacks to be executed by the service and assign them to jobs. Every job can be assigned with a different user defined
callback, a JSON payload with the data required for the callback to be executed, and an optional timeout interval.

The service exposes a JSON REST API providing CRUD endpoints for job the management. Configuration uses a single `yaml` file living under the root
directory of the project.

## Development

1. Clone the repo.

```bash
git clone https://github.com/svaloumas/valet.git
```

2. Build the `valet` executable.

```bash
make build
```

3. Run `Valet` service.

```bash
./valet
```

## Tests

Run the complete test suite.

```bash
make test
```