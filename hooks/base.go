package hooks

import "context"

type Base struct {
}

func (b *Base) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return ctx, nil
}

func (b *Base) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return ctx, nil
}
