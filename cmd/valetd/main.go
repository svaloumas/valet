package main

import (
	"github.com/gin-gonic/gin"

	"valet/internal/core/services/jobsrv"
	"valet/internal/handlers/jobhdl"
	"valet/internal/repositories/jobrepo"
	"valet/pkg/time"
	"valet/pkg/uuidgen"
)

var (
	addr = ":8080"
)

func main() {
	jobRepository := jobrepo.NewMemDB()
	jobService := jobsrv.New(jobRepository, uuidgen.New(), time.New())
	jobHandhler := jobhdl.NewHTTPHandler(jobService)

	router := gin.New()
	router.POST("/jobs", jobHandhler.Create)
	router.GET("/jobs/:id", jobHandhler.Get)
	router.DELETE("/jobs/:id", jobHandhler.Delete)

	router.Run(addr)
}
