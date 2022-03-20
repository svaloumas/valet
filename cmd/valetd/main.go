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
)

func main() {
	task := workerpool.NewDummyTask()

	wp := workerpool.NewWorkerPoolImpl(wpConcurrency, wpBacklog, task)
	wp.Start()

	jobRepository := jobrepo.NewMemDB()
	jobQueue := jobqueue.NewFIFOQueue(jobQueueCapacity)
	jobService := jobsrv.New(jobRepository, jobQueue, uuidgen.New(), rtime.New())

	jobTransmitter := workerpool.NewTransmitter(jobQueue, wp)
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
			log.Printf("[http] %s", err)
		}
	}()

	gracefulTerm := make(chan os.Signal, 1)
	signal.Notify(gracefulTerm, syscall.SIGINT, syscall.SIGTERM)
	sig := <-gracefulTerm
	log.Printf("[http] server notified %+v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("[http] failed to properly shutdown the server:", err)
	}
	log.Println("[http] server exiting...")
}
