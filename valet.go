package valet

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/core/service/jobsrv"
	"github.com/svaloumas/valet/internal/core/service/pipelinesrv"
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
	configPath       string
	taskService      port.TaskService
	gracefulTermChan chan os.Signal
}

// New initializes and returns a new valet instance.
func New(configPath string) *valet {
	taskService := tasksrv.New()
	gracefulTerm := make(chan os.Signal, 1)
	return &valet{
		configPath:       configPath,
		taskService:      taskService,
		gracefulTermChan: gracefulTerm,
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

	pipelineService := pipelinesrv.New(storage, taskrepo, uuidgen.New(), rtime.New())
	jobService := jobsrv.New(storage, taskrepo, uuidgen.New(), rtime.New())
	resultService := resultsrv.New(storage)

	workpoolLogger := vlog.NewLogger("workerpool", cfg.LoggingFormat)
	workService := worksrv.New(
		storage, taskrepo, rtime.New(), cfg.TimeoutUnit,
		cfg.WorkerPool.Concurrency, cfg.WorkerPool.Backlog, workpoolLogger)
	workService.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerLogger := vlog.NewLogger("scheduler", cfg.LoggingFormat)
	schedulerService := schedulersrv.New(jobQueue, storage, workService, rtime.New(), schedulerLogger)
	schedulerService.Schedule(ctx, time.Duration(cfg.Scheduler.RepositoryPollingInterval)*cfg.TimeoutUnit)
	schedulerService.Dispatch(ctx, time.Duration(cfg.Scheduler.JobQueuePollingInterval)*cfg.TimeoutUnit)

	server := factory.ServerFactory(
		cfg.Server, jobService, pipelineService, resultService,
		v.taskService, jobQueue, storage, cfg.LoggingFormat, logger)
	server.Serve()
	logger.Infof("initialized [%s] server", cfg.Server.Protocol)

	signal.Notify(v.gracefulTermChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-v.gracefulTermChan
	logger.Printf("server notified %+v", sig)
	server.GracefullyStop()

	jobQueue.Close()
	workService.Stop()
	storage.Close()
}

// Stop sends an INT signal to valet, which notifies the service to stop.
func (v *valet) Stop() {
	v.gracefulTermChan <- os.Interrupt
}

// RegisterTask registers a tack callback to the task repository under the specified name.
func (v *valet) RegisterTask(name string, callback func(...interface{}) (interface{}, error)) {
	v.taskService.Register(name, callback)
}

// DecodeTaskParams uses https://github.com/mitchellh/mapstructure
// to decode task params to a pointer of map or struct.
func DecodeTaskParams(args []interface{}, params interface{}) {
	mapstructure.Decode(args[0], params)
}

// DecodeTaskParams uses https://github.com/mitchellh/mapstructure
// to safely decode previous job's results metadata to a pointer of map or struct.
func DecodePreviousJobResults(args []interface{}, results interface{}) {
	if len(args) == 2 {
		mapstructure.Decode(args[1], results)
	}
}
