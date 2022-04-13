.PHONY: build test report fmt clean

BUILDCMD=env GOOS=linux GOARCH=amd64 go build -v
MYSQL_TEST_DSN="root:password@tcp(localhost:3306)/test?parseTime=true"
RABBITMQ_TEST_URI="amqp://quest:password@localhost:5672/"
REDIS_TEST_URL=redis://localhost/1

build: generate
	$(BUILDCMD) -o valetd cmd/valetd/*.go

test: generate
	MYSQL_DSN=$(MYSQL_TEST_DSN)	RABBITMQ_URI=$(RABBITMQ_TEST_URI) \
	REDIS_URL=$(REDIS_TEST_URL) go test `go list ./...` -v -cover -count=1

report: generate
	MYSQL_DSN=$(MYSQL_TEST_DSN)	RABBITMQ_URI=$(RABBITMQ_TEST_URI) \
	REDIS_URL=$(REDIS_TEST_URL) go test -v ./... -covermode=count -coverprofile=coverage.out
	go tool cover -func=coverage.out -o=coverage.out

fmt:
	! go fmt ./... 2>&1 | tee /dev/tty | read

generate:
	go generate ./...

clean:
	go clean ./...
