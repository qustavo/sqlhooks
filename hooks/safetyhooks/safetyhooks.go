package safetyhooks

import (
	"database/sql/driver"
	"fmt"
	"runtime"

	"github.com/gchaincl/sqlhooks/v2/hooks"
)

type Hook struct {
	hooks.Base
}

func New() *Hook {
	return &Hook{}
}

// safeRows wrap a driver.Rows interface in order to implement Sharp-Edged
// Finalizers based on https://crawshaw.io/blog/sharp-edged-finalizers.
type safeRows struct {
	driver.Rows
}

func (s *safeRows) Close() {
	runtime.SetFinalizer(s, nil)
	s.Rows.Close()
}

func doPanic() {
	_, file, line, _ := runtime.Caller(1)
	panic(fmt.Sprintf("%s:%d: row not closed", file, line))
}

func (h *Hook) Rows(r driver.Rows) driver.Rows {
	s := &safeRows{r}
	runtime.SetFinalizer(s, func(*safeRows) {
		doPanic()
	})

	return r
}
