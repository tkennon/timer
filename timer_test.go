package timer

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clock is a type that is stubbed out for timeAfter so that we can easily
// test the timer package.
type clock struct {
	next time.Duration
	fire bool
}

func newClock() *clock {
	return &clock{fire: true}
}

// After is a stub of time.After. It record the requested sleep duration and
// immediately returns the current time.
func (c *clock) After(next time.Duration) <-chan time.Time {
	c.next = next
	ch := make(chan time.Time, 1)
	if c.fire {
		ch <- time.Now()
	}
	return ch
}

type prng struct {
	val float64
}

func newPRNG() *prng {
	return &prng{}
}

func (p *prng) Float64() float64 {
	return p.val
}

func TestConstant(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	constant := NewConstant(time.Second)
	for trials := 0; trials < 2; trials++ {
		for i := 0; i < 100; i++ {
			c, err := constant.Start()
			require.NoError(t, err)
			assert.NotEmpty(t, <-c)
			assert.Equal(t, time.Second, fakeClock.next)
		}
		constant.Reset()
	}
}

func TestLinear(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	initial, increment := time.Second, time.Second
	linear := NewLinear(initial, increment)
	for trials := 0; trials < 2; trials++ {
		for i := 0; i < 100; i++ {
			c, err := linear.Start()
			require.NoError(t, err)
			assert.NotEmpty(t, <-c)
			assert.Equal(t, initial+time.Duration(i)*increment, fakeClock.next)
		}
		linear.Reset()
	}
}

func TestExponential(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	// To avoid floating point errors in a large loop we must keep the exponent
	// relatively low.
	initial := time.Second
	exponent := float32(1.1)
	exponential := NewExponential(initial, exponent)
	for trials := 0; trials < 2; trials++ {
		for i := 0; i < 100; i++ {
			c, err := exponential.Start()
			require.NoError(t, err)
			assert.NotEmpty(t, <-c)
			expected := float64(initial) * math.Pow(float64(exponent), float64(i))
			actual := float64(fakeClock.next)
			tolerance := float64(initial) * 0.01
			assert.InDelta(t, expected, actual, tolerance, "iteration %d", i)
		}
		exponential.Reset()
	}
}

func TestWithJitter(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After
	fakePRNG := newPRNG()
	ma := magnitude
	defer func() { magnitude = ma }()
	magnitude = fakePRNG.Float64
	jitter := 0.1

	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Second)},
		{NewLinear(time.Second, time.Second)},
		{NewExponential(time.Second, 2.0)},
	}
	for _, tt := range tests {
		for _, val := range []float64{-1.0, 0.0, 1.0} {
			timer := tt.timer.WithJitter(jitter)
			timer.Reset()
			fakePRNG.val = val
			c, err := timer.Start()
			require.NoError(t, err)
			assert.NotEmpty(t, <-c)
			assert.LessOrEqual(t, 0.0*float64(time.Second), float64(fakeClock.next))
			assert.GreaterOrEqual(t, 2.0*float64(time.Second), float64(fakeClock.next))
		}
	}
}

func TestWithMaxInterval(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	maxInterval := time.Minute
	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Hour)},
		{NewLinear(time.Second, time.Second)},
		{NewExponential(time.Second, 3.0)},
	}
	for _, tt := range tests {
		timer := tt.timer.WithMaxInterval(maxInterval)
		for i := 0; i < 100; i++ {
			c, err := timer.Start()
			require.NoError(t, err)
			assert.NotEmpty(t, <-c)
			assert.LessOrEqual(t, fakeClock.next.Nanoseconds(), maxInterval.Nanoseconds())
		}
	}
}

func TestWithMinInterval(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	minInterval := time.Second
	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Millisecond)},
		{NewLinear(time.Minute, -time.Second)},
		{NewExponential(time.Minute, 0.1)},
	}
	for _, tt := range tests {
		timer := tt.timer.WithMinInterval(minInterval)
		for i := 0; i < 100; i++ {
			c, err := timer.Start()
			require.NoError(t, err)
			assert.NotEmpty(t, <-c)
			assert.GreaterOrEqual(t, fakeClock.next.Nanoseconds(), minInterval.Nanoseconds())
		}
	}
}

func TestWithMaxDuration(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	maxDuration := time.Minute
	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Second)},
		{NewLinear(time.Second, time.Second)},
		{NewExponential(time.Second, 2.0)},
	}
	for _, tt := range tests {
		timer := tt.timer.WithMaxDuration(maxDuration)
		for {
			c, err := timer.Start()
			if err == nil {
				assert.NotEmpty(t, <-c)
			} else {
				break
			}
		}
		c, err := timer.Start()
		require.Error(t, err)
		require.Empty(t, c)
	}
}

func TestWithContext(t *testing.T) {
	ta := timeAfter
	defer func() { timeAfter = ta }()

	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Second)},
		{NewLinear(time.Second, time.Second)},
		{NewExponential(time.Second, 2.0)},
	}
	for _, tt := range tests {
		fakeClock := newClock()
		timeAfter = fakeClock.After
		ctx, cancel := context.WithCancel(context.Background())
		timer := tt.timer.WithContext(ctx)

		c, err := timer.Start()
		require.NoError(t, err)
		assert.NotEmpty(t, <-c)

		cancel()
		fakeClock.fire = false

		c, err = timer.Start()
		assert.NoError(t, err)
		select {
		case now := <-c:
			t.Log("timer fired", now)
			t.Fail()
		case <-time.After(time.Millisecond):
		}
	}
}

func TestWithFunc(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Second)},
		{NewLinear(time.Second, time.Second)},
		{NewExponential(time.Second, 2.0)},
	}
	for _, tt := range tests {
		done := make(chan struct{})
		timer := tt.timer.WithFunc(func() { close(done) })
		c, err := timer.Start()
		require.NoError(t, err)
		assert.NotEmpty(t, <-c)
		<-done
	}
}

func TestStop(t *testing.T) {
	ta := timeAfter
	defer func() { timeAfter = ta }()

	tests := []struct {
		timer *Timer
	}{
		{NewConstant(time.Second)},
		{NewLinear(time.Second, time.Second)},
		{NewExponential(time.Second, 2.0)},
	}
	for _, tt := range tests {
		fakeClock := newClock()
		timeAfter = fakeClock.After

		_, err := tt.timer.Start()
		require.NoError(t, err)
		stopped := tt.timer.Stop()
		assert.False(t, stopped)

		fakeClock.fire = false
		time.Sleep(10 * time.Millisecond)

		_, err = tt.timer.Start()
		require.NoError(t, err)
		stopped = tt.timer.Stop()
		assert.True(t, stopped)
	}
}

func TestInvalidSettings(t *testing.T) {
	fakeClock := newClock()
	ta := timeAfter
	defer func() { timeAfter = ta }()
	timeAfter = fakeClock.After

	linear := NewConstant(time.Minute).
		WithMaxInterval(time.Second).
		WithMinInterval(time.Hour)

	_, err := linear.Start()
	require.Error(t, err)
}
