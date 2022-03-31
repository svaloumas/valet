package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"valet/cmd/valetd/transmitter"
	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/service/jobsrv"
	"valet/internal/core/service/resultsrv"
	"valet/internal/repository/jobqueue"
	"valet/internal/repository/jobrepo"
	"valet/internal/repository/resultrepo"
	"valet/internal/repository/workerpool"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"
	"valet/task"

	_ "valet/doc/swagger"
)

func main() {
	logger := log.New(os.Stderr, "[valet] ", log.LstdFlags)

	cfg := new(Config)
	if err := cfg.Load(); err != nil {
		logger.Fatalf("could not load config: %s", err)
	}

	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("dummytask", task.DummyTask)

	jobQueue := jobqueue.NewFIFOQueue(cfg.JobQueueCapacity)

	jobRepository := jobrepo.NewJobDB()
	jobService := jobsrv.New(jobRepository, jobQueue, taskrepo, uuidgen.New(), rtime.New())

	resultRepository := resultrepo.NewResultDB()
	resultService := resultsrv.New(resultRepository)

	wp := workerpool.NewWorkerPoolImpl(jobService, resultService, cfg.WorkerPoolConcurrency, cfg.WorkerPoolBacklog)
	wp.Start()

	transmitterLogger := log.New(os.Stderr, "[transmitter] ", log.LstdFlags)
	jobTransmitter := transmitter.NewTransmitter(jobQueue, wp, transmitterLogger)
	go jobTransmitter.Transmit()

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: NewRouter(jobService, resultService),
	}

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
