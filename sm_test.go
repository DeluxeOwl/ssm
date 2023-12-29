package ssm

import (
	"context"
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
