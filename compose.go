package sqlhooks

import (
	"context"
	"fmt"
)

// Compose allows for composing multiple Hooks into one.
// It runs every callback on every hook in argument order,
// even if previous hooks return an error.
// If multiple hooks return errors, the error return value will be
// MultipleErrors, which allows for introspecting the errors if necessary.
func Compose(hooks ...Hooks) Hooks {
	return composed(hooks)
}

type composed []Hooks

func (c composed) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	var errors []error
	for _, hook := range c {
		c, err := hook.Before(ctx, query, args...)
		if err != nil {
			errors = append(errors, err)
		}
		if c != nil {
			ctx = c
		}
	}
	switch len(errors) {
	case 0:
		return ctx, nil
	case 1:
		return ctx, errors[0]
	default:
		return ctx, MultipleErrors(errors)
	}
}

func (c composed) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	var errors []error
	for _, hook := range c {
		var err error
		c, err := hook.After(ctx, query, args...)
		if err != nil {
			errors = append(errors, err)
		}
		if c != nil {
			ctx = c
		}
	}
	switch len(errors) {
	case 0:
		return ctx, nil
	case 1:
		return ctx, errors[0]
	default:
		return ctx, MultipleErrors(errors)
	}
}

func (c composed) OnError(ctx context.Context, cause error, query string, args ...interface{}) error {
	var errors []error
	for _, hook := range c {
		if onErrorer, ok := hook.(OnErrorer); ok {
			if err := onErrorer.OnError(ctx, cause, query, args...); err != nil && err != cause {
				errors = append(errors, err)
			}
		}
	}
	switch len(errors) {
	case 0:
		return cause
	case 1:
		return errors[0]
	default:
		return MultipleErrors(errors)
	}
}

// MultipleErrors is an error that contains multiple errors.
type MultipleErrors []error

func (m MultipleErrors) Error() string {
	return fmt.Sprint("multiple errors:", []error(m))
}
