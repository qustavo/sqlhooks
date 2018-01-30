package loghooks

import (
	"context"
	"log"
	"os"
	"time"
)

type logger interface {
	Printf(string, ...interface{})
}

type Hook struct {
	log logger
}

func New() *Hook {
	return &Hook{
		log: log.New(os.Stderr, "", log.LstdFlags),
	}
}
func (h *Hook) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, "started", time.Now()), nil
}

func (h *Hook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	h.log.Printf("Query: `%s`, Args: `%q`. took: %s", query, args, time.Since(ctx.Value("started").(time.Time)))
	return ctx, nil
}

func (h *Hook) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	h.log.Printf("Error: %v, Query: `%s`, Args: `%q`, Took: %s",
		err, query, args, time.Since(ctx.Value("started").(time.Time)))
	return err
}
