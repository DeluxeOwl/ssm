package ssm

import "context"

type errState struct {
	error
}

func (e errState) Error() string {
	return e.error.Error()
}

type smKeys string

const __start smKeys = "__start"

func StartState(ctx context.Context) Fn {
	start := ctx.Value(__start)
	if start == nil {
		return End
	}
	if state, ok := start.(func(context.Context) Fn); ok {
		return Fn(state)
	}
	return End
}

func (e errState) runStop(_ context.Context) Fn {
	return End
}

func (e errState) runRestart(ctx context.Context) Fn {
	return StartState(ctx)
}

func EndError(err error) Fn {
	return errState{err}.runStop
}

func RestartError(err error) Fn {
	return errState{err}.runRestart
}
