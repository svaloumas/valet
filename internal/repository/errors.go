package repository

import (
	"fmt"
)

// NotFoundErr is an error indicating an resource is not found.
type NotFoundErr struct {
	ID           string
	ResourceName string
}

func (e *NotFoundErr) Error() string {
	return fmt.Sprintf("%s with ID: %s not found", e.ResourceName, e.ID)
}
