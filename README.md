# sqlhooks [![Build Status](https://travis-ci.org/gchaincl/sqlhooks.svg)](https://travis-ci.org/gchaincl/sqlhooks) [![Coverage Status](https://coveralls.io/repos/github/gchaincl/sqlhooks/badge.svg?branch=master)](https://coveralls.io/github/gchaincl/sqlhooks?branch=master)

Attach hooks to any database/sql driver.

The purpose of sqlhooks is to provide a way to instrument your sql statements, making really easy to log queries or measure execution time without modifying your actual code.

# Install
```bash
go get github.com/gchaincl/sqlhooks
```

## Breaking changes
`V1` isn't backward compatible with previous versions, if you want to fetch old versions, you can get them from [gopkg.in](http://gopkg.in/)
```bash
go get gopkg.in/gchaincl/sqlhooks.v0
```

# Usage [![GoDoc](https://godoc.org/github.com/gchaincl/dotsql?status.svg)](https://godoc.org/github.com/gchaincl/sqlhooks)

```go
// This example shows how to instrument sql queries in order to display the time that they consume
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gchaincl/sqlhooks"
	"github.com/mattn/go-sqlite3"
)

// Hooks satisfies the sqlhook.Hooks interface
type Hooks struct {}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	fmt.Printf("> %s %q", query, args)
	return context.WithValue(ctx, "begin", time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	begin := ctx.Value("begin").(time.Time)
	fmt.Printf(". took: %s\n", time.Since(begin))
	return ctx, nil
}

func main() {
	// First, register the wrapper
	sql.Register("sqlite3WithHooks", sqlhooks.Wrap(&sqlite3.SQLiteDriver{}, &Hooks{}))

	// Connect to the registered wrapped driver
	db, _ := sql.Open("sqlite3WithHooks", ":memory:")

	// Do you're stuff
	db.Exec("CREATE TABLE t (id INTEGER, text VARCHAR(16))")
	db.Exec("INSERT into t (text) VALUES(?), (?)", "foo", "bar")
	db.Query("SELECT id, text FROM t")
}

/*
Output should look like:
> CREATE TABLE t (id INTEGER, text VARCHAR(16)) [[]]. took: 121.238µs
> INSERT into t (text) VALUES(?), (?) [[{"" '\x01' "foo"} {"" '\x02' "bar"}]]. took: 36.364µs
> SELECT id, text FROM t []. took: 4.653µs
*/
```
