package timer

import (
	"context"
	"math/rand"
	"time"
)

var (
	// The function that does the actual sleeping should be stubbed during
	// testing.
	timeAfter = time.After
	// magnitude returns a uniformly random number between [-1.0, 1.0).
	magnitude = func() float64 { return 1.0 - 2.0*rand.Float64() }
)

type interval interface {
	next() time.Duration
	reset()
}

// Timer is an object that sleeps.
type Timer struct {
	ctx         context.Context
	interval    interval
	total       time.Duration
	jitter      float64
	minInterval time.Duration
	maxInterval time.Duration
	cumDuration time.Duration
	stop        chan struct{}
	f           func()
}

func newTimer(interval interval) *Timer {
	return &Timer{
		ctx:      context.Background(),
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// WithJitter adds a uniformly random jitter to the time the timer next fires.
// The jitter will be within `fraction` of the current timer. e.g. if
// WithJitter(0.2) is applied to an exponential timer that would otherwise fire
// on 1, 2, 4, 8, ... seconds then the first timer will fire in between 0.8-1.2
// seconds, and the second will fire between 1.6-2.4 seconds. The jitter
// fraction may be greater than one, allowing the possible jittered timers to
// fire immediately if the calculated interval with the jitter is less than
// zero.
func (t *Timer) WithJitter(fraction float64) *Timer {
	t.jitter = fraction
	return t
}

// WithMinInterval sets the minimum interval between times the timer fires.
func (t *Timer) WithMinInterval(d time.Duration) *Timer {
	t.minInterval = d
	return t
}

// WithMaxInterval sets the maximum interval between times the timer fires.
func (t *Timer) WithMaxInterval(d time.Duration) *Timer {
	t.maxInterval = d
	return t
}

// WithCumulativeDuration sets the cumulative total duration the timer will run
// for over successive calls to Start. Once the maximum duration is reached,
// calls to start will fail.
func (t *Timer) WithCumulativeDuration(d time.Duration) *Timer {
	t.cumDuration = d
	return t
}

// WithContext adds a context.Context to the timer. If the context expires then
// the timer will also expire and will not fire again.
func (t *Timer) WithContext(ctx context.Context) *Timer {
	t.ctx = ctx
	return t
}

// WithFunc will execute f in its own goroutine after the timer has expired. The
// running of f can be stopped by Stopping the timer.
func (t *Timer) WithFunc(f func()) *Timer {
	t.f = f
	return t
}

// Start starts the timer. The interval the timer runs for is determined by the
// type of the timer: e.g. linear, exponential etc. Successive calls to Start
// will return channels that fire for different intervals. If it returns true,
// it returns a channel that will return the time of timer expiry, otherwise it
// returns nil. It will return nil, false if the timer
func (t *Timer) Start() (<-chan time.Time, bool) {
	next := t.interval.next()

	// Add jitter.
	jitter := time.Duration(t.jitter*magnitude()) * next
	next = next + jitter

	// Floor a single interval.
	if next < t.minInterval {
		next = t.minInterval
	}

	// Cap a single interval.
	if t.maxInterval.Nanoseconds() > 0 && next > t.maxInterval {
		next = t.maxInterval
	}

	// Cap the sum of all intervals.
	if t.cumDuration.Nanoseconds() > 0 && t.total+next > t.cumDuration {
		return nil, false
	}
	t.total += next

	// Asynchronously wait for the timer to expire.
	ch := make(chan time.Time, 1)
	go func() {
		select {
		case <-t.ctx.Done():
			return
		case <-t.stop:
			return
		case now := <-timeAfter(next):
			ch <- now
			if t.f != nil {
				t.f()
			}
		}
	}()

	return ch, true
}

// Reset resets the timer to its initial interval.
func (t *Timer) Reset() {
	t.interval.reset()
}

// Stop stops the timer from firing. It returns true if it stopped the timer
// from fring, and false if the timer was expired or not started. Stop does not
// close the channel returned by Start.
func (t *Timer) Stop() bool {
	select {
	case t.stop <- struct{}{}:
		return true
	default:
		return false
	}
}
