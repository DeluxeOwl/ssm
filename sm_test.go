package ssm

import (
	"context"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	iter := func(_ context.Context) Fn {
		return End
	}

	tests := []struct {
		name   string
		states []Fn
	}{
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "just start",
			states: []Fn{iter},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run(context.Background(), tt.states...)
		})
	}
}

func TestRunParallel(t *testing.T) {
	iter := func(_ context.Context) Fn {
		return End
	}

	tests := []struct {
		name   string
		states []Fn
	}{
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "just start",
			states: []Fn{iter},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunParallel(context.Background(), tt.states...)
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
