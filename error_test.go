package ssm

import (
	"context"
	"fmt"
	"testing"
)

func mockEmpty(_ context.Context) Fn {
	return End
}

func TestStartState(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want Fn
	}{
		{
			name: "empty",
			ctx:  context.Background(),
			want: nil,
		},
		{
			name: "some func",
			ctx:  context.WithValue(context.Background(), __start, mockEmpty),
			want: mockEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartState(tt.ctx); !sameFns(got, tt.want) {
				t.Errorf("StartState() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Fn
	}{
		{
			name: "empty",
			err:  nil,
			want: errState{}.stop,
		},
		{
			name: "random err",
			err:  fmt.Errorf("test"),
			want: errState{fmt.Errorf("test")}.stop,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorEnd(tt.err)
			if !sameFns(got, tt.want) {
				t.Errorf("ErrorEnd() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
			if st := got(context.Background()); !sameFns(st, End) {
				t.Errorf("Post run state for ErrorEnd() = %v, want %v", nameOf(st), nameOf(End))
			}
		})
	}
}

func TestErrorRestart(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		ctx      context.Context
		want     Fn
		endState Fn
	}{
		{
			name: "empty",
			err:  nil,
			want: errState{}.restart,
		},
		{
			name: "random err, background context",
			err:  fmt.Errorf("test"),
			want: errState{fmt.Errorf("test")}.restart,
		},
		{
			name:     "random err, start state in context",
			err:      fmt.Errorf("test"),
			ctx:      context.WithValue(context.Background(), __start, mockEmpty),
			want:     errState{fmt.Errorf("test")}.restart,
			endState: mockEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorRestart(tt.err)
			if !sameFns(got, tt.want) {
				t.Errorf("ErrorRestart() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
			if st := got(context.Background()); !sameFns(st, End) {
				t.Errorf("Post run state for ErrorRestart() = %v, want %v", nameOf(st), nameOf(End))
			}
		})
	}
}

func TestIsError(t *testing.T) {
	tests := []struct {
		name string
		f    Fn
		want bool
	}{
		{
			name: "nil",
			f:    nil,
			want: false,
		},
		{
			name: "End",
			f:    End,
			want: false,
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
			want: true,
		},
		{
			name: "ErrorRestart",
			f:    ErrorRestart(fmt.Errorf("test")),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsError(tt.f); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}
