package ssm

import (
	"context"
	"fmt"
	"testing"
	"time"
)

const delay = 10 * time.Millisecond

var backOffZero = backoff(0)
var backOffTen = backoff(10)

func TestBackOff(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want aggregatorFn
	}{
		{
			name: "zero",
			d:    0,
			want: backOffZero.run,
		},
		{
			name: "ten",
			d:    10,
			want: backOffTen.run,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BackOff(tt.d); ptrOf(got) != ptrOf(tt.want) {
				t.Errorf("BackOff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockMaxRetry(retries int) func(_ context.Context) error {
	for i := 0; i < retries; i++ {
		return func(_ context.Context) error {
			return fmt.Errorf("error")
		}
	}
	return func(_ context.Context) error {
		return nil
	}
}

var retryOnce = mockMaxRetry(1)
var retryNone = mockMaxRetry(0)

func TestRetry(t *testing.T) {
	type args struct {
		retries  int
		strategy aggregatorFn
		fn       func(ctx context.Context) error
	}
	tests := []struct {
		name string
		args args
		want Fn
	}{
		{
			name: "we get what we get",
			args: args{
				retries:  0,
				strategy: nil,
				fn:       retryOnce,
			},
			want: retry(0, nil, retryOnce),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Retry(tt.args.retries, tt.args.strategy, tt.args.fn); ptrOf(got) != ptrOf(tt.want) {
				t.Errorf("Retry() = %v, want %v", got, tt.want)
			}
		})
	}
}
