package ssm

import (
	"context"
	"fmt"
	"time"
)

type maxKey string

const max maxKey = "__max"

func start(ctx context.Context) Fn {
	fmt.Print("start ")
	i := iter(0)
	return i.next
}

type iter int

func (i *iter) next(ctx context.Context) Fn {
	fmt.Printf("%d ", *i)
	if m, ok := ctx.Value(max).(int); ok {
		if int(*i) == m {
			fmt.Print("end")
			return End
		}
	}
	*i = *i + 1
	return i.next
}

func ExampleCountTo10() {
	ctx := context.WithValue(context.Background(), max, 10)

	Run(ctx, start)

	// Output: start 0 1 2 3 4 5 6 7 8 9 10 end
}

func ExampleRetryWithBackOff() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries: ")
	start := Retry(4, BackOff(delay), func(_ context.Context) error {
		cnt++
		fmt.Printf("%d:%s ", cnt, time.Now().Sub(st).Truncate(delay))
		return fmt.Errorf("err")
	})

	Run(context.Background(), start)

	// Output: Retries: 1:0s 2:10ms 3:60ms 4:310ms
}

func ExampleRetryWithBatch() {
	st := time.Now()
	cnt := 0

	fmt.Printf("Retries: ")
	start := Retry(10, Batch, func(_ context.Context) error {
		cnt++
		fmt.Printf("%d:%s ", cnt, time.Now().Sub(st).Truncate(delay))
		return fmt.Errorf("err")
	})

	Run(context.Background(), start)

	// Output: Retries: 1:0s 2:0s 3:0s 4:0s 5:0s 6:0s 7:0s 8:0s 9:0s 10:0s
}
