package time

import (
	"time"
)

// Time represents time.
type Time interface {
	Now() time.Time
}

// RealTime is a concrete implementation of Time interface.
type RealTime struct{}

// New initializes and returns a new Time instance.
func New() Time {
	return &RealTime{}
}

// Now returns a timestamp of the current datetime in UTC.
func (rt *RealTime) Now() time.Time {
	loc, _ := time.LoadLocation("UTC")
	return time.Now().In(loc)
}
