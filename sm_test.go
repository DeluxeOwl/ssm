package ssm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestRunParallel(t *testing.T) {
	iter := func(_ context.Context) Fn {
		return End
	}

	tests := []struct {
		name   string
		states []Fn
		want   error
	}{
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "with error",
			states: []Fn{ErrorEnd(errors.New("test"))},
			want:   errors.New("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunParallel(context.Background(), tt.states...)
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

func TestRun(t *testing.T) {
	iter := func(_ context.Context) Fn {
		return End
	}

	tests := []struct {
		name   string
		states []Fn
		want   error
	}{
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "with error",
			states: []Fn{ErrorEnd(errors.New("test"))},
			want:   errors.New("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(context.Background(), tt.states...)
			if (err != nil) && !reflect.DeepEqual(tt.want, err) {
				t.Errorf("Run() error = %v, wanted %v", err, tt.want)
			}
		})
	}
}
