package task

import (
	"log"
	"time"

	"valet/internal/core/domain"
)

type dummytask struct{}

func NewDummyTask() *dummytask {
	return &dummytask{}
}

func (dt *dummytask) Run(j *domain.Job) ([]byte, error) {
	log.Printf("[dummy task] hello from dummy task")
	time.Sleep(1 * time.Second)
	return make([]byte, 0), nil
}
