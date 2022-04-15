package factory

import (
	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/repository/jobqueue"
)

func JobQueueFactory(cfg config.JobQueue, loggingFormat string) port.JobQueue {
	if cfg.Option == "memory" {
		return jobqueue.NewFIFOQueue(cfg.MemoryJobQueue.Capacity)
	}
	return jobqueue.NewRabbitMQ(cfg.RabbitMQ, loggingFormat)
}
