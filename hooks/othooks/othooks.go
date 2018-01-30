package othooks

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type Hook struct {
	tracer opentracing.Tracer
}

func New(tracer opentracing.Tracer) *Hook {
	return &Hook{tracer: tracer}
}

func (h *Hook) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	parent := opentracing.SpanFromContext(ctx)
	if parent == nil {
		return ctx, nil
	}

	span := h.tracer.StartSpan("sql", opentracing.ChildOf(parent.Context()))
	span.LogFields(
		log.String("query", query),
		log.Object("args", args),
	)

	return opentracing.ContextWithSpan(ctx, span), nil
}

func (h *Hook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		defer span.Finish()
	}

	return ctx, nil
}

func (h *Hook) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		defer span.Finish()
		span.SetTag("error", true)
		span.LogFields(
			log.Error(err),
		)
	}

	return err
}
