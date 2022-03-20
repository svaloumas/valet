package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"valet/internal/core/services/jobsrv"
	"valet/internal/handlers/jobhdl"
	"valet/internal/repositories/jobqueue"
	"valet/internal/repositories/jobrepo"
	"valet/internal/workerpool"
	rtime "valet/pkg/time"
	"valet/pkg/uuidgen"
)

var (
	addr             = ":8080"
	jobQueueCapacity = 100
	wpConcurrency    = runtime.NumCPU() / 2
	wpBacklog        = wpConcurrency * 2
	tickInterval     = 500 * time.Millisecond
)

func main() {
	task := workerpool.NewDummyTask()

	wp := workerpool.NewWorkerPoolImpl(wpConcurrency, wpBacklog, task)
	wp.Start()

	logger := log.New(os.Stderr, "[valet] ", log.LstdFlags)

	jobRepository := jobrepo.NewMemDB()
	jobQueue := jobqueue.NewFIFOQueue(jobQueueCapacity)
	jobService := jobsrv.New(jobRepository, jobQueue, uuidgen.New(), rtime.New())

	jobTransmitter := workerpool.NewTransmitter(jobQueue, wp, int(tickInterval))
	go jobTransmitter.Transmit()

	jobHandhler := jobhdl.NewHTTPHandler(jobService)

	router := gin.New()
	router.POST("/jobs", jobHandhler.Create)
	router.GET("/jobs/:id", jobHandhler.Get)
	router.DELETE("/jobs/:id", jobHandhler.Delete)

	srv := http.Server{
		Addr:    addr,
		Handler: router,
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
