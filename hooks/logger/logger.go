// Package logger provides a query logger
package logger

import (
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/gchaincl/sqlhooks"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type hook struct {
	id  uint64
	Log Logger
}

func (h *hook) next() uint64 {
	return atomic.AddUint64(&h.id, 1)
}

func New() *hook {
	return &hook{
		Log: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (h *hook) before(ctx *sqlhooks.Context) error {
	id := h.next()
	ctx.Set("start", time.Now())
	ctx.Set("id", id)

	h.Log.Printf("[query#%09d] %s %v", id, ctx.Query, ctx.Args)
	return nil

}

func (h *hook) after(ctx *sqlhooks.Context) error {
	id := ctx.Get("id")
	took := time.Since(ctx.Get("start").(time.Time))

	if err := ctx.Error; err != nil {
		h.Log.Printf("[query#%09d] Finished with error: %v", id, err)
		return err
	}

	h.Log.Printf("[query#%09d] took %s", id, took)
	return nil
}

func (h *hook) BeforeQuery(ctx *sqlhooks.Context) error {
	return h.before(ctx)
}

func (h *hook) AfterQuery(ctx *sqlhooks.Context) error {
	return h.after(ctx)
}

func (h *hook) BeforeExec(ctx *sqlhooks.Context) error {
	return h.before(ctx)
}

func (h *hook) AfterExec(ctx *sqlhooks.Context) error {
	return h.after(ctx)
}
