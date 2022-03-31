# sqlhooks
![Build Status](https://github.com/qustavo/sqlhooks/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/qustavo/sqlhooks)](https://goreportcard.com/report/github.com/qustavo/sqlhooks)
[![Coverage Status](https://coveralls.io/repos/github/qustavo/sqlhooks/badge.svg?branch=master)](https://coveralls.io/github/qustavo/sqlhooks?branch=master)

Attach hooks to any database/sql driver.

The purpose of sqlhooks is to provide a way to instrument your sql statements, making really easy to log queries or measure execution time without modifying your actual code.

# Install
```bash
go get github.com/qustavo/sqlhooks/v2
```
Requires Go >= 1.14.x

## Breaking changes
`V2` isn't backward compatible with previous versions, if you want to fetch old versions, you can use go modules or get them from [gopkg.in](http://gopkg.in/)
```bash
go get github.com/qustavo/sqlhooks
go get gopkg.in/qustavo/sqlhooks.v1
```

# Usage [![GoDoc](https://godoc.org/github.com/qustavo/dotsql?status.svg)](https://godoc.org/github.com/qustavo/sqlhooks)

```go
// This example shows how to instrument sql queries in order to display the time that they consume
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/qustavo/sqlhooks/v2"
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
> CREATE TABLE t (id INTEGER, text VARCHAR(16)) []. took: 121.238µs
> INSERT into t (text) VALUES(?), (?) ["foo" "bar"]. took: 36.364µs
> SELECT id, text FROM t []. took: 4.653µs
*/
```

# Benchmarks
```
 go test -bench=. -benchmem
 goos: linux
 goarch: amd64
 pkg: github.com/qustavo/sqlhooks/v2
 cpu: Intel(R) Xeon(R) W-10885M CPU @ 2.40GHz
 BenchmarkSQLite3/Without_Hooks-16                 191196              6163 ns/op             456 B/op         14 allocs/op
 BenchmarkSQLite3/With_Hooks-16                    189997              6329 ns/op             456 B/op         14 allocs/op
 BenchmarkMySQL/Without_Hooks-16                    13278             83462 ns/op             309 B/op          7 allocs/op
 BenchmarkMySQL/With_Hooks-16                       13460             87331 ns/op             309 B/op          7 allocs/op
 BenchmarkPostgres/Without_Hooks-16                 13016             91421 ns/op             401 B/op         10 allocs/op
 BenchmarkPostgres/With_Hooks-16                    12339             94033 ns/op             401 B/op         10 allocs/op
 PASS
 ok      github.com/qustavo/sqlhooks/v2  10.294s
```
