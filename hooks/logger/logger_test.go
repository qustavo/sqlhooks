package logger

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/gchaincl/sqlhooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHook() (*hook, *bytes.Buffer) {
	buf := bytes.Buffer{}
	hook := New()
	hook.Log = log.New(&buf, "", 0)

	return hook, &buf
}

func TestLoggerQuery(t *testing.T) {
	hook, buf := newTestHook()

	ctx := sqlhooks.NewContext()
	ctx.Query = "SELECT * FROM table"
	ctx.Args = []interface{}{"1", 2}

	require.NoError(t, hook.BeforeQuery(ctx))
	assert.Contains(t, buf.String(), "[query#000000001] ")
	assert.Contains(t, buf.String(), ctx.Query)
	assert.Contains(t, buf.String(), "[1 2]")

	buf.Reset()
	require.NoError(t, hook.AfterQuery(ctx))
	assert.Contains(t, buf.String(), "[query#000000001] ")
	assert.Contains(t, buf.String(), "took")
}

func TestLoggerExec(t *testing.T) {
	hook, buf := newTestHook()

	ctx := sqlhooks.NewContext()
	ctx.Query = "INSERT INTO table (foo, bar) VALUES (?, ?)"
	ctx.Args = []interface{}{"x", "z"}

	require.NoError(t, hook.BeforeExec(ctx))
	assert.Contains(t, buf.String(), "[query#000000001] ")
	assert.Contains(t, buf.String(), ctx.Query)
	assert.Contains(t, buf.String(), "[x z]")

	buf.Reset()
	require.NoError(t, hook.AfterExec(ctx))
	assert.Contains(t, buf.String(), "[query#000000001] ")
	assert.Contains(t, buf.String(), "took")
}

func TestLoggerWithErrors(t *testing.T) {
	hook, buf := newTestHook()

	ctx := sqlhooks.NewContext()
	ctx.Query = "INSERT INTO table (foo, bar) VALUES (?, ?)"
	ctx.Args = []interface{}{"x", "z"}
	ctx.Error = errors.New("boom")

	require.NoError(t, hook.BeforeExec(ctx))

	buf.Reset()
	require.Error(t, hook.AfterExec(ctx))
	assert.Contains(t, buf.String(), "Finished with error: boom")
}

func TestLoggerIncrementsQueryCounter(t *testing.T) {
	hook, buf := newTestHook()

	ctx := sqlhooks.NewContext()
	for _ = range [9]bool{} {
		hook.BeforeQuery(ctx)
	}

	buf.Reset()
	hook.BeforeQuery(ctx)
	assert.Contains(t, buf.String(), "[query#000000010] ")

	buf.Reset()
	hook.AfterQuery(ctx)
	assert.Contains(t, buf.String(), "[query#000000010] ")
}
