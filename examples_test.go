package ssm_test

import (
	"context"
	"fmt"
	"time"

	sm "git.sr.ht/~mariusor/ssm"
)

type maxKey string

const max maxKey = "__max"
const delay = 10 * time.Millisecond

func start(ctx context.Context) sm.Fn {
	fmt.Print("start ")
	i := iter(0)
	return i.next
}

type iter int

func (i *iter) next(ctx context.Context) sm.Fn {
	fmt.Printf("%d ", *i)
	if m, ok := ctx.Value(max).(int); ok {
		if int(*i) == m {
			fmt.Print("end")
			return sm.End
		}
	}
	*i = *i + 1
	return i.next
}

func ExampleRun() {
	ctx := context.WithValue(context.Background(), max, 10)

	sm.Run(ctx, start)

	// Output: start 0 1 2 3 4 5 6 7 8 9 10 end
}

func ExampleBackOff() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries: ")
	start := sm.Retry(5, sm.BackOff(sm.Linear(delay, 2), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		fmt.Printf("%d:%s ", cnt, run.Sub(st).Truncate(10*time.Millisecond))
		st = run
		return sm.ErrorEnd(fmt.Errorf("err"))
	}))

	sm.Run(context.Background(), start)

	// Output: Retries: 1:10ms 2:20ms 3:40ms 4:80ms 5:160ms
}

func ExampleBackOff2() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries: ")
	start := sm.Retry(5, sm.BackOff(sm.Linear(delay, 2), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		fmt.Printf("%d:%s ", cnt, run.Sub(st).Truncate(10*time.Millisecond))
		st = run
		if cnt < 4 {
			return sm.ErrorEnd(fmt.Errorf("err"))
		}
		return sm.End
	}))

	sm.Run(context.Background(), start)

	// Output: Retries: 1:10ms 2:20ms 3:40ms 4:80ms
}

func ExampleRetry() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries: ")
	start := sm.Retry(10, func(_ context.Context) sm.Fn {
		cnt++
		run := time.Now()
		fmt.Printf("%d:%s ", cnt, run.Sub(st).Truncate(time.Millisecond))
		st = run
		return sm.ErrorEnd(fmt.Errorf("err"))
	})

	sm.Run(context.Background(), start)

	// Output: Retries: 1:0s 2:0s 3:0s 4:0s 5:0s 6:0s 7:0s 8:0s 9:0s 10:0s
}

func ExampleRetry2() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries: ")
	start := sm.Retry(10, sm.BackOff(sm.Linear(delay, 2), func(_ context.Context) sm.Fn {
		run := time.Now()
		cnt++
		fmt.Printf("%d:%s ", cnt, run.Sub(st).Truncate(10*time.Millisecond))
		st = run
		if cnt < 4 {
			return sm.ErrorEnd(fmt.Errorf("err"))
		}
		// We resolve to a non error state after 4 retries
		return sm.End
	}))

	sm.Run(context.Background(), start)

	// Output: Retries: 1:10ms 2:20ms 3:40ms 4:80ms
}
