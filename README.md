# Valet service

Simple stateless Go server responsible for executing tasks hereinafter referred to as jobs.

## Overview

At its core, `Valet` is an asynchronous task executor.

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