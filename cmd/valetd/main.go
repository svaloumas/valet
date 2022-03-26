package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"valet/internal/core/service/jobsrv"
	"valet/internal/core/service/resultsrv"
	"valet/internal/repository/jobqueue"
	"valet/internal/repository/jobrepo"
	"valet/internal/repository/resultrepo"
	"valet/internal/repository/workerpool"
	"valet/internal/repository/workerpool/task"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"

	_ "valet/doc/swagger"
)

var (
	addr             = ":8080"
	jobQueueCapacity = 100
	wpConcurrency    = runtime.NumCPU() / 2
	wpBacklog        = wpConcurrency * 2
	tickInterval     = 500 * time.Millisecond
	taskType         = "dummytask"
)

func main() {
	taskFunc := task.TaskTypes[taskType]
	jobQueue := jobqueue.NewFIFOQueue(jobQueueCapacity)

	jobRepository := jobrepo.NewJobDB()
	jobService := jobsrv.New(jobRepository, jobQueue, uuidgen.New(), rtime.New())

	resultRepository := resultrepo.NewResultDB()
	resultService := resultsrv.New(resultRepository)

	wp := workerpool.NewWorkerPoolImpl(
		jobService, resultService, wpConcurrency, wpBacklog, taskFunc)
	wp.Start()

	jobTransmitter := NewTransmitter(jobQueue, wp, int(tickInterval))
	go jobTransmitter.Transmit()

	srv := http.Server{
		Addr:    addr,
		Handler: NewRouter(jobService, resultService),
	}

	logger := log.New(os.Stderr, "[valet] ", log.LstdFlags)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Printf("%s", err)
		}
	}()

	gracefulTerm := make(chan os.Signal, 1)
	signal.Notify(gracefulTerm, syscall.SIGINT, syscall.SIGTERM)
	sig := <-gracefulTerm
	logger.Printf("server notified %+v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("failed to properly shutdown the server:", err)
	}
	logger.Println("server exiting...")

	jobTransmitter.Stop()
	jobQueue.Close()
	wp.Stop()
}
