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
	"valet/internal/core/service/schedulersrv"
	"valet/internal/core/service/worksrv"
	"valet/internal/repository/jobqueue"
	"valet/internal/repository/jobrepo"
	"valet/internal/repository/resultrepo"
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
	cfg := new(Config)
	if err := cfg.Load(); err != nil {
		log.Fatalf("could not load config: %s", err)
	}

	logger := vlog.NewLogger("valet", cfg.LoggingFormat)

	taskrepo := taskrepo.NewTaskRepository()
	taskrepo.Register("dummytask", task.DummyTask)

	jobQueue := jobqueue.NewFIFOQueue(cfg.JobQueueCapacity)

	jobRepository := jobrepo.NewJobDB()
	resultRepository := resultrepo.NewResultDB()

	jobService := jobsrv.New(jobRepository, jobQueue, taskrepo, uuidgen.New(), rtime.New())
	resultService := resultsrv.New(resultRepository)

	workpoolLogger := vlog.NewLogger("workerpool", cfg.LoggingFormat)
	workService := worksrv.New(
		jobRepository, resultRepository,
		taskrepo, rtime.New(), cfg.TimeoutUnit,
		cfg.WorkerPoolConcurrency, cfg.WorkerPoolBacklog, workpoolLogger)
	workService.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerLogger := vlog.NewLogger("consumer", cfg.LoggingFormat)
	consumerService := consumersrv.New(jobQueue, workService, consumerLogger)
	consumerService.Consume(ctx, time.Duration(cfg.JobQueuePollingInterval)*cfg.TimeoutUnit)

	schedulerLogger := vlog.NewLogger("scheduler", cfg.LoggingFormat)
	schedulerService := schedulersrv.New(jobRepository, workService, rtime.New(), schedulerLogger)
	schedulerService.Schedule(ctx, time.Duration(cfg.SchedulerPollingInterval)*cfg.TimeoutUnit)

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: NewRouter(jobService, resultService, cfg.LoggingFormat),
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
