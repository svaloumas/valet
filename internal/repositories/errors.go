package repositories

import (
	"fmt"
)

type NotFoundError struct {
	ID           string
	ResourceName string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID: %s not found", e.ID, e.ResourceName)
}
