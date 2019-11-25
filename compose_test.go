package sqlhooks

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

var (
	oops     = errors.New("oops")
	oopsHook = &testHooks{
		before: func(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
			return ctx, oops
		},
		after: func(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
			return ctx, oops
		},
		onError: func(ctx context.Context, err error, query string, args ...interface{}) error {
			return oops
		},
	}
	okHook = &testHooks{
		before: func(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
			return ctx, nil
		},
		after: func(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
			return ctx, nil
		},
		onError: func(ctx context.Context, err error, query string, args ...interface{}) error {
			return nil
		},
	}
)

func TestCompose(t *testing.T) {
	for _, it := range []struct {
		name  string
		hooks Hooks
		want  error
	}{
		{"happy case", Compose(okHook, okHook), nil},
		{"no hooks", Compose(), nil},
		{"multiple errors", Compose(oopsHook, okHook, oopsHook), MultipleErrors([]error{oops, oops})},
		{"single error", Compose(okHook, oopsHook, okHook), oops},
	} {
		t.Run(it.name, func(t *testing.T) {
			t.Run("Before", func(t *testing.T) {
				_, got := it.hooks.Before(context.Background(), "query")
				if !reflect.DeepEqual(it.want, got) {
					t.Errorf("unexpected error. want: %q, got: %q", it.want, got)
				}
			})
			t.Run("After", func(t *testing.T) {
				_, got := it.hooks.After(context.Background(), "query")
				if !reflect.DeepEqual(it.want, got) {
					t.Errorf("unexpected error. want: %q, got: %q", it.want, got)
				}
			})
			t.Run("OnError", func(t *testing.T) {
				cause := errors.New("crikey")
				want := it.want
				if want == nil {
					want = cause
				}
				got := it.hooks.(OnErrorer).OnError(context.Background(), cause, "query")
				if !reflect.DeepEqual(want, got) {
					t.Errorf("unexpected error. want: %q, got: %q", want, got)
				}
			})
		})
	}
}

func TestWrapErrors(t *testing.T) {
	var (
		err1 = errors.New("oops")
		err2 = errors.New("oops2")
	)
	for _, it := range []struct {
		name   string
		def    error
		errors []error
		want   error
	}{
		{"no errors", err1, nil, err1},
		{"single error", nil, []error{err1}, err1},
		{"multiple errors", nil, []error{err1, err2}, MultipleErrors([]error{err1, err2})},
	} {
		t.Run(it.name, func(t *testing.T) {
			if want, got := it.want, wrapErrors(it.def, it.errors); !reflect.DeepEqual(want, got) {
				t.Errorf("unexpected wrapping. want: %q, got %q", want, got)
			}
		})
	}
}
