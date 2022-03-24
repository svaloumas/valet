package apperrors

import (
	"fmt"
)

// FullQueueErr is an error to indicate that a queue is full.
type FullQueueErr struct{}

func (e *FullQueueErr) Error() string {
	return "job queue is full - try again later"

}

// NotFoundErr is an error indicating an resource is not found.
type NotFoundErr struct {
	ID           string
	ResourceName string
}

func (e *NotFoundErr) Error() string {
	return fmt.Sprintf("%s with ID: %s not found", e.ResourceName, e.ID)
}

// FullWorkerPoolBacklog is an error indicating that the
// the worker pool backlog queue is full.
type FullWorkerPoolBacklog struct{}

func (e *FullWorkerPoolBacklog) Error() string {
	return "worker pool backlog is full"
}