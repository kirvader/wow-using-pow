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

type ClockMock struct {
	Counter int64
}

func (cm *ClockMock) Now() time.Time {
	cm.Counter += 1
	return time.Unix(cm.Counter, 0)
}
