package timer_test

import (
	"context"
	"fmt"
	"time"

	"github.com/tkennon/timer"
)

func example(t *timer.Timer, round time.Duration) error {
	for i := 0; i < 5; i++ {
		then := time.Now()
		c, err := t.Start()
		if err != nil {
			return err
		}
		now := <-c
		fmt.Println(now.Sub(then).Round(round))
	}

	return nil
}

func ExampleNewConstant() {
	con := timer.NewConstant(time.Millisecond)
	if err := example(con, time.Millisecond); err != nil {
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
	if err := example(lin, time.Millisecond); err != nil {
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
	if err := example(exp, time.Millisecond); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 2ms
	// 4ms
	// 8ms
	// 16ms
}

func ExampleWithMinInterval() {
	lin := timer.NewLinear(5*time.Millisecond, -time.Millisecond).WithMinInterval(3 * time.Millisecond)
	if err := example(lin, time.Millisecond); err != nil {
		panic(err)
	}

	// Output:
	// 5ms
	// 4ms
	// 3ms
	// 3ms
	// 3ms
}

func ExampleWithMaxInterval() {
	lin := timer.NewLinear(time.Millisecond, time.Millisecond).WithMaxInterval(3 * time.Millisecond)
	if err := example(lin, time.Millisecond); err != nil {
		panic(err)
	}

	// Output:
	// 1ms
	// 2ms
	// 3ms
	// 3ms
	// 3ms
}

func ExampleWithMaxDuration() {
	exp := timer.NewExponential(time.Millisecond, 2.0).WithMaxDuration(10 * time.Millisecond)
	err := example(exp, time.Millisecond)
	fmt.Println(err)

	// Output:
	// 1ms
	// 2ms
	// 4ms
	// maximum timer duration elapsed
}

func ExampleWithContext() {
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

func ExampleWithFunc() {
	con := timer.NewConstant(time.Millisecond).WithFunc(func() {
		fmt.Println("hello")
	})
	if err := example(con, time.Millisecond); err != nil {
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

func ExampleReset() {

}

func ExampleStop() {

}
