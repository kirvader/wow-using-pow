package clock

import "time"

var _ Clock = &SystemClock{}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (s *SystemClock) Now() time.Time {
	return time.Now()
}
