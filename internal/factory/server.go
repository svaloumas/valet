package factory

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/handler/jobhdl"
	jobpb "github.com/svaloumas/valet/internal/handler/jobhdl/protobuf"
	"github.com/svaloumas/valet/internal/handler/resulthdl"
	resultpb "github.com/svaloumas/valet/internal/handler/resulthdl/protobuf"
	"github.com/svaloumas/valet/internal/handler/server"
	"github.com/svaloumas/valet/internal/handler/taskhdl"
	taskpb "github.com/svaloumas/valet/internal/handler/taskhdl/protobuf"
)

const (
	HTTP = "http"
	gRPC = "grpc"
)

func ServerFactory(
	cfg config.Server,
	jobService port.JobService,
	resultService port.ResultService,
	taskService port.TaskService,
	jobQueue port.JobQueue,
	storage port.Storage, loggingFormat string,
	logger *logrus.Logger) port.Server {

	if cfg.Protocol == HTTP {
		srv := http.Server{
			Addr:    ":" + cfg.HTTP.Port,
			Handler: server.NewRouter(jobService, resultService, taskService, jobQueue, storage, loggingFormat),
		}
		httpsrv := server.NewHTTPServer(srv, logger)
		return httpsrv
	}
	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	jobgRPCHandler := jobhdl.NewJobgRPCHandler(jobService)
	resultgRPCHandler := resulthdl.NewResultgRPCHandler(resultService)
	taskgRPCHandler := taskhdl.NewTaskgRPCHandler(taskService)
	jobpb.RegisterJobServer(s, jobgRPCHandler)
	resultpb.RegisterJobResultServer(s, resultgRPCHandler)
	taskpb.RegisterTaskServer(s, taskgRPCHandler)
	reflection.Register(s)

	grpcsrv := server.NewGRPCServer(s, listener, logger)
	return grpcsrv
}
