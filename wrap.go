package ssm

import "context"

func Wrap(fn func(ctx context.Context) error) Fn {
	if fn == nil {
		return End
	}
	return func(ctx context.Context) Fn {
		if err := fn(ctx); err != nil {
			return ErrorEnd(err)
		}
		return End
	}
}

func WrapRepeat(fn func(ctx context.Context) error) Fn {
	if fn == nil {
		return End
	}
	return func(ctx context.Context) Fn {
		if err := fn(ctx); err != nil {
			return ErrorEnd(err)
		}
		return WrapRepeat(fn)
	}
}
