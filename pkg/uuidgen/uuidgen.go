package uuidgen

import (
	"github.com/google/uuid"
)

// UUIDGenerator represents a UUID generator.
type UUIDGenerator interface {
	GenerateRandomUUIDString() (string, error)
}

// UUIDGen is a concrete implementation of UUIDGenerator interface.
type UUIDGen struct{}

func New() UUIDGenerator {
	return &UUIDGen{}
}

// GenerateRandomUUIDString generates a new random UUID string .
func (u *UUIDGen) GenerateRandomUUIDString() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", nil
	}
	return uuid.String(), nil
}
