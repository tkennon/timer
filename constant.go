package timer

import "time"

type constant time.Duration

func (c constant) next() time.Duration {
	return time.Duration(c)
}

func (constant) reset() {
}

// NewConstant returns a constant timer, functionally equivalent to the standard
// library time.Timer.
func NewConstant(interval time.Duration) *Timer {
	return newTimer(constant(interval))
}
