package ssm

import (
	"context"
	"fmt"
	"testing"
	"time"
)

const delay = 10 * time.Millisecond

func mockStateWithDelay(t *testing.T, delay time.Duration) Fn {
	startTime := time.Now()
	i := 0
	return func(ctx context.Context) Fn {
		diff := time.Now().Sub(startTime)
		trunc := delay
		if delay < time.Millisecond {
			trunc = 10 * time.Microsecond
		}
		if diff.Truncate(trunc) != delay {
			t.Errorf("Execution time: %s, expected %s", diff, delay)
		} else {
			t.Logf("Executed after %s", diff)
		}
		i++
		startTime = time.Now()
		return End
	}
}

func TestBackOff(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
	}{
		{
			name: "ten",
			d:    10 * time.Millisecond,
		},
		{
			name: "hundred",
			d:    100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BackOff(Constant(tt.d), mockStateWithDelay(t, tt.d))
			got(context.Background())
			got(context.Background())
		})
	}
}

func mockMaxRetry(retries int) func(_ context.Context) Fn {
	for i := 0; i < retries; i++ {
		return func(_ context.Context) Fn {
			return ErrorEnd(fmt.Errorf("error"))
		}
	}
	return func(_ context.Context) Fn {
		return End
	}
}

var retryOnce = mockMaxRetry(1)
var retryNone = mockMaxRetry(0)

func TestRetry(t *testing.T) {
	type args struct {
		retries int
		fn      Fn
	}
	tests := []struct {
		name string
		args args
		want Fn
	}{
		{
			name: "one run",
			args: args{
				retries: 0,
				fn:      retryOnce,
			},
			want: retry(0, retryOnce),
		},
		{
			name: "one retry",
			args: args{
				retries: 1,
				fn:      retryNone,
			},
			want: retry(1, retryNone),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Retry(tt.args.retries, tt.args.fn); ptrOf(got) != ptrOf(tt.want) {
				t.Errorf("Retry() = %v, want %v", got, tt.want)
			}
		})
	}
}
