.PHONY: build test report fmt clean

SHELL=/bin/bash -o pipefail

BUILDCMD=env GOOS=darwin GOARCH=amd64 go build -v
BUILD_TIME=`TZ=UTC date +%FT%T%z`
COMMIT=`git rev-parse --short HEAD`
VERSION=0.1.0

build: generate
	$(BUILDCMD) $(LDFLAGS) -o valetd cmd/valetd/*.go

test: generate
	go test `go list ./...` -v -cover -count=1

report: generate
	go test -v ./... -covermode=count -coverprofile=coverage.out
	go tool cover -func=coverage.out -o=coverage.out

fmt:
	! go fmt ./... 2>&1 | tee /dev/tty | read

generate:
	go generate ./...

clean:
	go clean ./...
