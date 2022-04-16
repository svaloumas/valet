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
