package uuidgen

import (
	"github.com/google/uuid"
)

type UUIDGenerator interface {
	GenerateRandomUUIDString() (string, error)
}

type UUIDGen struct{}

func New() UUIDGenerator {
	return &UUIDGen{}
}

func (u *UUIDGen) GenerateRandomUUIDString() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", nil
	}
	return uuid.String(), nil
}
