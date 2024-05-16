package ssm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestRunParallel(t *testing.T) {
	iter := func(_ context.Context) Fn {
		return End
	}

	stopAfter1Ms, _ := context.WithTimeout(context.Background(), time.Millisecond)

	type args struct {
		ctx    context.Context
		states []Fn
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "just start",
			args: args{
				ctx:    context.Background(),
				states: []Fn{iter},
			},
		},
		{
			name: "just start",
			args: args{
				ctx:    context.Background(),
				states: []Fn{iter},
			},
		},
		{
			name: "with error",
			args: args{
				ctx:    context.Background(),
				states: []Fn{ErrorEnd(errors.New("test"))},
			},
			want: errors.New("test"),
		},
		{
			name: "with deadline of 1ms",
			args: args{
				ctx:    stopAfter1Ms,
				states: []Fn{After(10*time.Millisecond, End)},
			},
			want: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunParallel(tt.args.ctx, tt.args.states...)
			if (err != nil) && !reflect.DeepEqual(tt.want, err) {
				t.Errorf("RunParallel() error = %v, wanted %v", err, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	iter := func(_ context.Context) Fn {
		return End
	}

	stopAfter1Ms, _ := context.WithTimeout(context.Background(), time.Millisecond)

	type args struct {
		ctx    context.Context
		states []Fn
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "just start",
			args: args{
				ctx:    context.Background(),
				states: []Fn{iter},
			},
		},
		{
			name: "just start",
			args: args{
				ctx:    context.Background(),
				states: []Fn{iter},
			},
		},
		{
			name: "with error",
			args: args{
				ctx:    context.Background(),
				states: []Fn{ErrorEnd(errors.New("test"))},
			},
			want: errors.New("test"),
		},
		{
			name: "with deadline of 1ms",
			args: args{
				ctx:    stopAfter1Ms,
				states: []Fn{After(10*time.Millisecond, End)},
			},
			want: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.args.ctx, tt.args.states...)
			if (err != nil) && !reflect.DeepEqual(tt.want, err) {
				t.Errorf("Run() error = %v, wanted %v", err, tt.want)
			}
		})
	}
}

func TestIsEnd(t *testing.T) {
	tests := []struct {
		name string
		f    Fn
		want bool
	}{
		{
			name: "nil",
			f:    nil,
			want: true,
		},
		{
			name: "End",
			f:    End,
			want: true,
		},
		{
			name: "mockEmpty",
			f:    mockEmpty,
			want: false,
		},
		{
			name: "state func literal",
			f: func(ctx context.Context) Fn {
				return End
			},
			want: false,
		},
		{
			name: "ErrorEnd",
			f:    ErrorEnd(fmt.Errorf("test")),
			want: false,
		},
		{
			name: "ErrorRestart",
			f:    ErrorRestart(fmt.Errorf("test")),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEnd(tt.f); got != tt.want {
				t.Errorf("IsEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}
