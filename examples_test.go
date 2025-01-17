package ssm_test

import (
	"context"
	"errors"
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

func ExampleRunWithContextValue() {
	ctx := context.WithValue(context.Background(), _max, 10)

	sm.Run(ctx, start)

	// Output: start 0 1 2 3 4 5 6 7 8 9 10 end
}

func loop(_ context.Context) sm.Fn {
	return sm.After(time.Millisecond, loop)
}

func ExampleRunWithTimeout() {
	ctx, _ := context.WithTimeout(context.Background(), 10*delay)

	err := sm.Run(ctx, loop)

	fmt.Printf("%s", err)

	// Output: context deadline exceeded
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
	con := sm.Retry(8, sm.BackOff(sm.Constant(delay), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		fmt.Printf("%d:%s\n", cnt, run.Sub(st).Truncate(delay))
		st = run
		return sm.ErrorEnd(fmt.Errorf("keep going"))
	}))

	sm.Run(context.Background(), con)

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

func ExampleBreakerWithCount() {
	cnt := 0

	fmt.Printf("Stop after 2 successes and 2 failures\n")
	brk := sm.Breaker(sm.MaxTriesTrip(2), func(ctx context.Context) sm.Fn {
		defer func() { cnt++ }()
		if cnt < 2 {
			fmt.Printf("success[%d]\n", cnt)
			return sm.End
		}
		fmt.Printf("failure[%d]\n", cnt)
		return sm.ErrorRestart(errors.New("fail"))
	})

	sm.Run(context.Background(), brk)
	// Output: Stop after 2 successes and 2 failures
	// success[0]
	// success[1]
	// failure[2]
	// failure[3]
}

func roundRound(d time.Duration) sm.Fn {
	it := 0
	errCnt := 0
	var st time.Time

	return func(ctx context.Context) sm.Fn {
		// we simulate work by sleeping for duration "d"
		time.Sleep(d)
		defer func() { it++ }()

		// we return failures only when iteration count is divisible by 3
		if it != 0 && it%3 == 0 {
			if it == 3 {
				st = time.Now()
			}
			errCnt++
			fmt.Printf("failure[%d] after %s\n", errCnt, time.Now().Sub(st).Truncate(2*time.Millisecond).String())
			return sm.ErrorRestart(errors.New("fail"))
		}

		return roundRound(d)
	}
}

func ExampleBreakerWithTimer() {
	fmt.Printf("Stop after 3 failures in 10ms\n")

	brk := sm.Breaker(sm.TimedTrip(10*time.Millisecond, sm.MaxTriesTrip(3)), roundRound(time.Millisecond))

	sm.Run(context.Background(), brk)

	// Output: Stop after 3 failures in 10ms
	// failure[1] after 0s
	// failure[2] after 2ms
	// failure[3] after 6ms
}

func ExampleContextDone() {
	ctx, stopFn := context.WithCancelCause(context.Background())
	err := sm.Run(ctx, func(ctx context.Context) sm.Fn {
		stopFn(errors.New("hahahaha"))
		return sm.End
	})

	fmt.Printf("%s", err)                // hahahaha
	fmt.Printf("%s", context.Cause(ctx)) // hahahaha, because the same error cause can be found in our external context

	// Output: hahahahahahahaha
}

func ExampleContextCancel() {
	ctx := context.Background()
	err := sm.Run(ctx, func(ctx context.Context) sm.Fn {
		stopFn := sm.Cancel(ctx)
		stopFn(errors.New("hahahaha"))
		return sm.End
	})

	fmt.Printf("%s", err)

	// Output: hahahaha
}

func ExampleTimeoutExceeded() {
	fmt.Printf("Timeout\n")
	st := time.Now()
	err := sm.Run(context.Background(), sm.Timeout(400*time.Millisecond, func(ctx context.Context) sm.Fn {
		fmt.Println("Sleeping for 1s")
		time.Sleep(time.Second)
		return sm.End
	}))
	fmt.Printf("Error: %s after %s", err, time.Now().Sub(st).Truncate(100*time.Millisecond))

	// Output: Timeout
	// Sleeping for 1s
	// Error: context deadline exceeded after 400ms
}

func ExampleTimeout() {
	fmt.Printf("Timeout\n")
	st := time.Now()
	err := sm.Run(context.Background(), sm.Timeout(400*time.Millisecond, func(ctx context.Context) sm.Fn {
		fmt.Println("Sleeping for 200ms")
		time.Sleep(200 * time.Millisecond)
		return sm.ErrorEnd(fmt.Errorf("we reached the end"))
	}))
	fmt.Printf("Error: %s after %s", err, time.Now().Sub(st).Truncate(100*time.Millisecond))

	// Output: Timeout
	// Sleeping for 200ms
	// Error: we reached the end after 200ms
}
