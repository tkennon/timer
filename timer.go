package timer

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// Functions stubbed out during testing.
var (
	// timeAfter does the actual sleeping.
	timeAfter = time.After
	// magnitude returns a uniformly random number in the range [-1.0, 1.0).
	magnitude = func() float64 { return 1.0 - 2.0*rand.Float64() }
)

// Errors returned by the timer package.
var (
	// ErrMaxDurationElapsed is returned from (*Timer).Start() when the maximum
	// cumulative timer duration has elapsed.
	ErrMaxDurationElapsed = errors.New("maximum timer duration elapsed")
	// ErrInvalidSettings is returned from (*Timer).Start() when invalid timer
	// settings are detected; such as if the minimum duration had been set to
	// greater than the maximum duration.
	ErrInvalidSettings = errors.New("invalid timer settings")
)

type interval interface {
	next() time.Duration
	reset()
}

// Timer is an object that sleeps.
type Timer struct {
	ctx            context.Context
	interval       interval
	total          time.Duration
	jitter         float64
	minIntervalSet bool
	minInterval    time.Duration
	maxIntervalSet bool
	maxInterval    time.Duration
	maxDurationSet bool
	maxDuration    time.Duration
	stop           chan struct{}
	f              func()
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
	t.minIntervalSet = true
	return t
}

// WithMaxInterval sets the maximum interval between times the timer fires.
func (t *Timer) WithMaxInterval(d time.Duration) *Timer {
	t.maxInterval = d
	t.maxIntervalSet = true
	return t
}

// WithMaxDuration sets the maxiumum cumulative total duration the timer will
// run for over successive calls to Start. Once the maximum duration is reached,
// calls to start will fail.
func (t *Timer) WithMaxDuration(d time.Duration) *Timer {
	t.maxDuration = d
	t.maxDurationSet = true
	return t
}

// WithContext adds a context.Context to the timer. If the context expires then
// the timer will also expire and will not fire again.
func (t *Timer) WithContext(ctx context.Context) *Timer {
	t.ctx = ctx
	return t
}

// WithFunc will execute f in its own goroutine after the timer has expired. To
// prevent running f, the timer must be stopped before f is invoked.
func (t *Timer) WithFunc(f func()) *Timer {
	t.f = f
	return t
}

// Start starts the timer. The returned channel will send the current time after
// an interval has elapsed. It returns an error if the timer could not be
// started due to restrictions imposed in the timer config (e.g. maximum
// duration reached). Successive calls to Start will return channels that fire
// for different intervals. The difference in the intervals is determiend by the
// type of time: e.g. linear or exponential etc.
func (t *Timer) Start() (<-chan time.Time, error) {
	// Sanity check the min/max intervals.
	if t.minIntervalSet && t.maxIntervalSet {
		if t.minInterval > t.maxInterval {
			return nil, ErrInvalidSettings
		}
	}

	// Get the next interval.
	next := t.interval.next()

	// Add jitter.
	next += time.Duration(t.jitter*magnitude()) * next

	// Floor a single interval.
	if t.minIntervalSet && next < t.minInterval {
		next = t.minInterval
	}

	// Cap a single interval.
	if t.maxIntervalSet && next > t.maxInterval {
		next = t.maxInterval
	}

	// Cap the sum of all intervals.
	if t.maxDurationSet && t.total+next > t.maxDuration {
		return nil, ErrMaxDurationElapsed
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

	return ch, nil
}

// Reset resets the timer to its initial interval, but retains all timer
// configuration (such as jitter, max/min intervals etc).
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
