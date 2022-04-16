package valet

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/core/service/consumersrv"
	"github.com/svaloumas/valet/internal/core/service/jobsrv"
	"github.com/svaloumas/valet/internal/core/service/resultsrv"
	"github.com/svaloumas/valet/internal/core/service/schedulersrv"
	"github.com/svaloumas/valet/internal/core/service/tasksrv"
	"github.com/svaloumas/valet/internal/core/service/worksrv"
	"github.com/svaloumas/valet/internal/factory"
	vlog "github.com/svaloumas/valet/pkg/log"
	rtime "github.com/svaloumas/valet/pkg/time"
	"github.com/svaloumas/valet/pkg/uuidgen"

	_ "github.com/svaloumas/valet/doc/swagger"
)

type valet struct {
	configPath  string
	taskService port.TaskService
}

// New initializes and returns a new valet instance.
func New(configPath string) *valet {
	taskService := tasksrv.New()
	return &valet{
		configPath:  configPath,
		taskService: taskService,
	}
}

// Run runs the service.
func (v *valet) Run() {
	cfg := new(config.Config)
	filepath, _ := filepath.Abs(v.configPath)
	if err := cfg.Load(filepath); err != nil {
		log.Fatalf("could not load config: %s", err)
	}

	logger := vlog.NewLogger("valet", cfg.LoggingFormat)

	taskrepo := v.taskService.GetTaskRepository()

	for _, name := range taskrepo.GetTaskNames() {
		logger.Infof("registered task with name: %s", name)
	}

	jobQueue := factory.JobQueueFactory(cfg.JobQueue, cfg.LoggingFormat)
	logger.Infof("initialized [%s] as a job queue", cfg.JobQueue.Option)

	storage := factory.StorageFactory(cfg.Repository)
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

	server := factory.ServerFactory(
		cfg.Server, jobService, resultService, v.taskService, storage, cfg.LoggingFormat, logger)
	server.Serve()
	logger.Infof("initialized [%s] server", cfg.Server.Protocol)

	gracefulTerm := make(chan os.Signal, 1)
	signal.Notify(gracefulTerm, syscall.SIGINT, syscall.SIGTERM)
	sig := <-gracefulTerm
	logger.Printf("server notified %+v", sig)
	server.GracefullyStop()

	jobQueue.Close()
	workService.Stop()
	storage.Close()
}

// RegisterTask registers a tack callback to the task repository under the specified name.
func (v *valet) RegisterTask(name string, callback func(interface{}) (interface{}, error)) {
	v.taskService.Register(name, callback)
}
