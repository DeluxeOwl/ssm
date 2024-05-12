package ssm_test

import (
	"context"
	"fmt"
	"time"

	sm "git.sr.ht/~mariusor/ssm"
)

type maxKey string

const _max maxKey = "__max"
const delay = 10 * time.Millisecond

func start(ctx context.Context) sm.Fn {
	fmt.Print("start ")
	i := iter(0)
	return i.next
}

type iter int

func (i *iter) next(ctx context.Context) sm.Fn {
	fmt.Printf("%d ", *i)
	if m, ok := ctx.Value(_max).(int); ok {
		if int(*i) == m {
			fmt.Print("end")
			return sm.End
		}
	}
	*i = *i + 1
	return i.next
}

func ExampleRun() {
	ctx := context.WithValue(context.Background(), _max, 10)

	sm.Run(ctx, start)

	// Output: start 0 1 2 3 4 5 6 7 8 9 10 end
}

func ExampleRetry() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries\n")
	retry := sm.Retry(10, func(_ context.Context) sm.Fn {
		cnt++
		run := time.Now()
		fmt.Printf("%d:%s\n", cnt, run.Sub(st).Truncate(time.Millisecond))
		st = run
		return sm.ErrorEnd(fmt.Errorf("retrying"))
	})

	sm.Run(context.Background(), retry)

	// Output: Retries
	// 1:0s
	// 2:0s
	// 3:0s
	// 4:0s
	// 5:0s
	// 6:0s
	// 7:0s
	// 8:0s
	// 9:0s
	// 10:0s
}

func ExampleConstant() {
	st := time.Now()
	cnt := 0

	fmt.Printf("8 retries with constant backoff of 10ms\n")
	cnst := sm.Retry(8, sm.BackOff(sm.Constant(delay), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		fmt.Printf("%d:%s\n", cnt, run.Sub(st).Truncate(delay))
		st = run
		return sm.ErrorEnd(fmt.Errorf("keep going"))
	}))

	sm.Run(context.Background(), cnst)

	// Output: 8 retries with constant backoff of 10ms
	// 1:10ms
	// 2:10ms
	// 3:10ms
	// 4:10ms
	// 5:10ms
	// 6:10ms
	// 7:10ms
	// 8:10ms
}

func ExampleLinear() {
	st := time.Now()
	cnt := 0

	fmt.Printf("5 retries with 2x linear backoff 10ms delay\n")
	linear := sm.Retry(5, sm.BackOff(sm.Linear(delay, 2), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		fmt.Printf("%d:%s\n", cnt, run.Sub(st).Truncate(delay))
		st = run
		return sm.ErrorEnd(fmt.Errorf("don't stop"))
	}))

	sm.Run(context.Background(), linear)

	// Output: 5 retries with 2x linear backoff 10ms delay
	// 1:10ms
	// 2:20ms
	// 3:40ms
	// 4:80ms
	// 5:160ms
}

func ExampleJitter() {
	st := time.Now()
	cnt := 0

	fmt.Printf("2 retries with 1ms jitter over 10ms delay\n")
	jitter := sm.Retry(2, sm.BackOff(sm.Jitter(time.Millisecond, sm.Constant(delay)), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		// NOTE(marius): The jitter adds a maximum of 1 ms, so with truncation of 2 ms the output will be correct
		// This is not a very good example test, though.
		fmt.Printf("%d:%s\n", cnt, run.Sub(st).Truncate(2*time.Millisecond))
		st = run
		return sm.ErrorEnd(fmt.Errorf("never right"))
	}))

	sm.Run(context.Background(), jitter)

	// Output: 2 retries with 1ms jitter over 10ms delay
	// 1:10ms
	// 2:10ms
}
