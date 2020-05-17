package timer

import "time"

type linear struct {
	initial   time.Duration
	current   time.Duration
	increment time.Duration
}

func (l *linear) next() time.Duration {
	current := l.current
	l.current += l.increment
	return current
}

func (l *linear) reset() {
	l.current = l.initial
}

// NewLinear returns a linear backoff timer.
func NewLinear(initial, increment time.Duration) *Timer {
	return newTimer(&linear{
		initial:   initial,
		current:   initial,
		increment: increment,
	})
}
