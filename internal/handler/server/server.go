package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// HTTPServer represents an HTTP server.
type HTTPServer struct {
	srv    http.Server
	logger *logrus.Logger
}

// GRPCServer represents a gRPC server.
type GRPCServer struct {
	srv      *grpc.Server
	listener net.Listener
	logger   *logrus.Logger
}

// NewHTTPServer creates and returns a new HTTPServer instance.
func NewHTTPServer(srv http.Server, logger *logrus.Logger) *HTTPServer {
	return &HTTPServer{
		srv:    srv,
		logger: logger,
	}
}

// NewGRPCServer creates and returns a new GRPCServer instance.
func NewGRPCServer(srv *grpc.Server, listener net.Listener, logger *logrus.Logger) *GRPCServer {
	return &GRPCServer{
		srv:      srv,
		listener: listener,
		logger:   logger,
	}
}

// Serve starts the server.
func (s *HTTPServer) Serve() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			s.logger.Printf("failed to serve HTTP: %s", err)
		}
	}()
}

// GracefullyStop gracefully stops the server.
func (s *HTTPServer) GracefullyStop() {
	s.logger.Println("server exiting...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Fatal("failed to properly shutdown the server:", err)
	}
}

// Serve starts the server.
func (s *GRPCServer) Serve() {
	go func() {
		go func() {
			s.logger.Printf("listening gRPC at: %s", s.listener.Addr().String())
			if err := s.srv.Serve(s.listener); err != nil {
				s.logger.Printf("failed to serve gRPC: %s", err)
			}
		}()
	}()
}

// GracefullyStop gracefully stops the server.
func (s *GRPCServer) GracefullyStop() {
	s.logger.Println("server exiting...")
	s.srv.GracefulStop()
}
