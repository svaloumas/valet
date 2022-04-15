package factory

import (
	"valet/internal/config"
	"valet/internal/core/port"
	"valet/internal/repository/jobqueue"
)

func JobQueueFactory(cfg config.JobQueue, loggingFormat string) port.JobQueue {
	if cfg.Option == "memory" {
		return jobqueue.NewFIFOQueue(cfg.MemoryJobQueue.Capacity)
	}
	return jobqueue.NewRabbitMQ(cfg.RabbitMQ, loggingFormat)
}
