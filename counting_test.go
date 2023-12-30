package ssm

import (
	"context"
	"fmt"
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
