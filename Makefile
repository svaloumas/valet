.PHONY: build-linux build-mac build-win test report fmt clean

BUILDCMD_LINUX=env GOOS=linux GOARCH=amd64 go build -v
BUILDCMD_MAC=env GOOS=darwin GOARCH=amd64 go build -v
BUILDCMD_WIN=env GOOS=windows GOARCH=amd64 go build -v
MYSQL_TEST_DSN="root:password@tcp(localhost:3306)/test?parseTime=true"
POSTGRES_TEST_DSN="postgres://postgres:postgres@localhost:5432/test?sslmode=disable"
RABBITMQ_TEST_URI="amqp://quest:password@localhost:5672/"
REDIS_TEST_URL="redis://localhost/1"

build-linux: generate
	$(BUILDCMD_LINUX) -o valetd cmd/valetd/*.go

build-mac: generate
	$(BUILDCMD_MAC) -o valetd cmd/valetd/*.go

build-win: generate
	$(BUILDCMD_WIN) -o valetd.exe cmd\valetd\*.go

test: generate
	MYSQL_DSN=$(MYSQL_TEST_DSN)	POSTGRES_DSN=$(POSTGRES_TEST_DSN) RABBITMQ_URI=$(RABBITMQ_TEST_URI) \
	REDIS_URL=$(REDIS_TEST_URL) go test `go list ./...` -v -cover -count=1

report: generate
	MYSQL_DSN=$(MYSQL_TEST_DSN)	POSTGRES_DSN=$(POSTGRES_TEST_DSN) RABBITMQ_URI=$(RABBITMQ_TEST_URI) \
	REDIS_URL=$(REDIS_TEST_URL) go test `go list ./...` -cover -covermode=count -coverprofile=coverage.out

fmt:
	! go fmt ./... 2>&1 | tee /dev/tty | read

generate:
	go generate ./...

clean:
	go clean ./...
