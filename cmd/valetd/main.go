package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"valet/internal/config"
	"valet/internal/core/domain/taskrepo"
	"valet/internal/core/service/consumersrv"
	"valet/internal/core/service/jobsrv"
	"valet/internal/core/service/resultsrv"
	"valet/internal/core/service/schedulersrv"
	"valet/internal/core/service/worksrv"
	"valet/internal/factory"
	vlog "valet/pkg/log"
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
	cfg := new(config.Config)
	filepath, _ := filepath.Abs("config.yaml")
	if err := cfg.Load(filepath); err != nil {
		log.Fatalf("could not load config: %s", err)
	}

	logger := vlog.NewLogger("valet", cfg.LoggingFormat)

	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("dummytask", task.DummyTask)

	jobQueue := factory.JobQueueFactory(cfg.JobQueue, cfg.LoggingFormat)
	logger.Infof("initialized [%s] as a job queue", cfg.JobQueue.Option)

	storage := factory.StorageFactory(cfg.Repository, cfg.LoggingFormat)
	logger.Infof("initialized [%s] as a repository", cfg.Repository.Option)

	jobService := jobsrv.New(storage, jobQueue, taskrepo, uuidgen.New(), rtime.New())
	resultService := resultsrv.New(storage)

	workpoolLogger := vlog.NewLogger("workerpool", cfg.LoggingFormat)
	workService := worksrv.New(
		storage, taskrepo, rtime.New(), cfg.TimeoutUnit,
		cfg.WorkerPool.Concurrency, cfg.WorkerPool.Backlog, workpoolLogger)
	workService.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerLogger := vlog.NewLogger("consumer", cfg.LoggingFormat)
	consumerService := consumersrv.New(jobQueue, workService, consumerLogger)
	consumerService.Consume(ctx, time.Duration(cfg.Consumer.JobQueuePollingInterval)*cfg.TimeoutUnit)

	schedulerLogger := vlog.NewLogger("scheduler", cfg.LoggingFormat)
	schedulerService := schedulersrv.New(storage, workService, rtime.New(), schedulerLogger)
	schedulerService.Schedule(ctx, time.Duration(cfg.Scheduler.RepositoryPollingInterval)*cfg.TimeoutUnit)

	srv := http.Server{
		Addr:    ":" + cfg.Server.HTTPPort,
		Handler: NewRouter(jobService, resultService, storage, cfg.LoggingFormat),
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

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("failed to properly shutdown the server:", err)
	}
	logger.Println("server exiting...")

	jobQueue.Close()
	workService.Stop()
}
