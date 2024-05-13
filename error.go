package ssm

import (
	"context"
	"reflect"
)

// ErrorEnd represents an error state which returns an End state.
func ErrorEnd(err error) Fn {
	return errState{err}.stop
}

// ErrorRestart represents an error state which returns the first iteration passed.
// This iteration is loaded from the context, and is saved there by the Run and RunParallel functions.
func ErrorRestart(err error) Fn {
	return errState{err}.restart
}

// StartState retrieves the initial state from ctx context.Context.
// If nothing is found it returns the End state.
func StartState(ctx context.Context) Fn {
	start := ctx.Value(__start)
	if start == nil {
		return End
	}
	if state, ok := start.(func(context.Context) Fn); ok {
		return state
	}
	return End
}

func Cancel(ctx context.Context) context.CancelCauseFunc {
	cancel := ctx.Value(__cancel)
	if cancel == nil {
		return nil
	}
	if cancelFn, ok := cancel.(context.CancelCauseFunc); ok {
		return cancelFn
	}
	return nil
}

type smKeys string

const __start smKeys = "__start"
const __cancel smKeys = "__cancel"

// Error this state
func (e errState) Error() string {
	return e.error.Error()
}

type errState struct {
	error
}

type ErrorFn Fn

func (f ErrorFn) Error() string {
	return "error"
}

func ptrOf(t Fn) uintptr {
	return reflect.ValueOf(t).Pointer()
}

var (
	_ptrEndStop    = ptrOf(errState{}.stop)
	_ptrEndRestart = ptrOf(errState{}.restart)
)

// IsError ver grubby API to check if a state Fn is an error state
func IsError(f Fn) bool {
	p := ptrOf(f)
	return p == _ptrEndStop || p == _ptrEndRestart
}

func (e errState) stop(ctx context.Context) Fn {
	cancelFn := Cancel(ctx)
	if cancelFn != nil {
		defer cancelFn(e.error)
	}
	return End
}

func (e errState) restart(ctx context.Context) Fn {
	cancelFn := Cancel(ctx)
	if cancelFn != nil {
		defer cancelFn(e.error)
	}
	return StartState(ctx)
}
