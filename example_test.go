package timer_test

import (
	"context"
	"fmt"
	"time"

	"github.com/tkennon/timer"
)

func runFiveTimes(t *timer.Timer) error {
	for i := 0; i < 5; i++ {
		then := time.Now()
		c, err := t.Start()
		if err != nil {
			return err
		}
		now := <-c
		fmt.Println(now.Sub(then).Round(time.Millisecond))
	}

	return nil
}

func ExampleNewConstant() {
	con := timer.NewConstant(time.Millisecond)
	if err := runFiveTimes(con); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 1ms
	// 1ms
	// 1ms
	// 1ms
}

func ExampleNewLinear() {
	lin := timer.NewLinear(time.Millisecond, time.Millisecond)
	if err := runFiveTimes(lin); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 2ms
	// 3ms
	// 4ms
	// 5ms
}

func ExampleNewExponential() {
	exp := timer.NewExponential(time.Millisecond, 2.0)
	if err := runFiveTimes(exp); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 2ms
	// 4ms
	// 8ms
	// 16ms
}

func ExampleTimer_WithMinInterval() {
	lin := timer.NewLinear(5*time.Millisecond, -time.Millisecond).WithMinInterval(3 * time.Millisecond)
	if err := runFiveTimes(lin); err != nil {
		panic(err)
	}

	// Output:
	// 5ms
	// 4ms
	// 3ms
	// 3ms
	// 3ms
}

func ExampleTimer_WithMaxInterval() {
	lin := timer.NewLinear(time.Millisecond, time.Millisecond).WithMaxInterval(3 * time.Millisecond)
	if err := runFiveTimes(lin); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 2ms
	// 3ms
	// 3ms
	// 3ms
}

func ExampleTimer_WithMaxDuration() {
	exp := timer.NewExponential(time.Millisecond, 2.0).WithMaxDuration(10 * time.Millisecond)
	err := runFiveTimes(exp)
	fmt.Println(err)

	// Output:
	// 1ms
	// 2ms
	// 4ms
	// maximum timer duration elapsed
}

func Example_Timer_WithContext() {
	ctx, cancel := context.WithCancel(context.Background())
	con := timer.NewConstant(time.Millisecond).WithContext(ctx)
	cancel()
	c, err := con.Start()
	if err != nil {
		panic(err)
	}
	select {
	case <-c:
		fmt.Println("timer fired after context cancelation")
	case <-time.After(5 * time.Millisecond):
		fmt.Println("timer did not fire")
	}

	// Output:
	// timer did not fire
}

func ExampleTimer_WithFunc() {
	con := timer.NewConstant(time.Millisecond).WithFunc(func() {
		fmt.Println("hello")
	})
	if err := runFiveTimes(con); err != nil {
		panic(err)
	}

	// Output:
	// hello
	// 1ms
	// hello
	// 1ms
	// hello
	// 1ms
	// hello
	// 1ms
	// hello
	// 1ms
}

func ExampleTimer_Reset() {
	exp := timer.NewExponential(time.Millisecond, 2.0)
	if err := runFiveTimes(exp); err != nil {
		panic(err)
	}

	fmt.Println("resetting timer")
	exp.Reset()
	if err := runFiveTimes(exp); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 2ms
	// 4ms
	// 8ms
	// 16ms
	// resetting timer
	// 1ms
	// 2ms
	// 4ms
	// 8ms
	// 16ms
}

func ExampleTimer_Stop() {
	con := timer.NewConstant(time.Millisecond)
	fmt.Println("timer was running:", con.Stop())
	c, err := con.Start()
	if err != nil {
		panic(err)
	}
	fmt.Println("timer was running:", con.Stop())
	select {
	case <-c:
		fmt.Println("timer fired even after it was stopped")
	case <-time.After(5 * time.Millisecond):
		fmt.Println("timer did not fire after it was stopped")
	}

	// Output:
	// timer was running: false
	// timer was running: true
	// timer did not fire after it was stopped
}
