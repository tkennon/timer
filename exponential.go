package timer

import (
	"time"
)

type exponential struct {
	initial    time.Duration
	current    time.Duration
	multiplier float32
}

func (e *exponential) next() time.Duration {
	current := e.current
	e.current = time.Duration(float32(e.current) * e.multiplier)
	return current
}

func (e *exponential) reset() {
	e.current = e.initial
}

// NewExponential returns an exponential backoff timer.
func NewExponential(initial time.Duration, multiplier float32) *Timer {
	return newTimer(&exponential{
		initial:    initial,
		current:    initial,
		multiplier: multiplier,
	})
}
