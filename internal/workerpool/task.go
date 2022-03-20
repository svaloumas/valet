package workerpool

import (
	"log"
	"valet/internal/core/domain"
)

type dummytask struct{}

func NewDummyTask() *dummytask {
	return &dummytask{}
}

func (dt *dummytask) Run(j *domain.Job) ([]byte, error) {
	log.Printf("[dummy task] hello from dummy task")
	return make([]byte, 0), nil
}
