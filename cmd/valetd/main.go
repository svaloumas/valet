package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/service/consumersrv"
	"valet/internal/core/service/jobsrv"
	"valet/internal/core/service/resultsrv"
	"valet/internal/core/service/worksrv"
	"valet/internal/repository/jobqueue"
	"valet/internal/repository/jobrepo"
	"valet/internal/repository/resultrepo"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"
	"valet/task"

	_ "valet/doc/swagger"
)

var (
	buildTime = "undefined"
	commit    = "undefined"
	version   = "undefined"
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
	resultRepository := resultrepo.NewResultDB()

	jobService := jobsrv.New(jobRepository, jobQueue, taskrepo, uuidgen.New(), rtime.New())
	resultService := resultsrv.New(resultRepository)

	workService := worksrv.New(
		jobRepository, resultRepository, taskrepo, rtime.New(), cfg.WorkerPoolConcurrency, cfg.WorkerPoolBacklog)

	consumerLogger := log.New(os.Stderr, "[consumer] ", log.LstdFlags)
	consumerService := consumersrv.New(jobQueue, workService, consumerLogger)
	go consumerService.Consume()

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

	consumerService.Stop()
}
