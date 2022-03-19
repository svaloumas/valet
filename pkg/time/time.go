package time

import (
	"time"
)

type Time interface {
	Now() time.Time
}

type RealTime struct{}

func New() Time {
	return &RealTime{}
}

func (rt *RealTime) Now() time.Time {
	loc, _ := time.LoadLocation("UTC")
	return time.Now().In(loc)
}
