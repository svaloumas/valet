package factory

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"valet/internal/config"
	"valet/internal/core/port"
	"valet/internal/handler/jobhdl"
	jobpb "valet/internal/handler/jobhdl/protobuf"
	"valet/internal/handler/resulthdl"
	resultpb "valet/internal/handler/resulthdl/protobuf"
	"valet/internal/handler/server"
)

const (
	HTTP = "http"
	gRPC = "grpc"
)

func ServerFactory(
	cfg config.Server,
	jobService port.JobService,
	resultService port.ResultService,
	storage port.Storage, loggingFormat string,
	logger *logrus.Logger) port.Server {

	if cfg.Protocol == HTTP {
		srv := http.Server{
			Addr:    ":" + cfg.HTTP.Port,
			Handler: server.NewRouter(jobService, resultService, storage, loggingFormat),
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
	jobpb.RegisterJobServer(s, jobgRPCHandler)
	resultpb.RegisterJobResultServer(s, resultgRPCHandler)
	reflection.Register(s)

	grpcsrv := server.NewGRPCServer(s, listener, logger)
	return grpcsrv
}
