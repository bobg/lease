package lease

import "time"

// Clock is an interface used by [Provider] for time-based operations.
// It can be replaced with a mock implementation for testing purposes.
// Note that one good mock implementation is in github.com/benbjohnson/clock, qv.
type Clock interface {
	Now() time.Time
	After(time.Duration) <-chan time.Time
}

// DefaultClock implements the [Clock] interface in terms of the stdlib [time].
type DefaultClock struct{}

func (DefaultClock) Now() time.Time                         { return time.Now() }
func (DefaultClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
