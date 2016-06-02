# sqlhooks [![Build Status](https://travis-ci.org/gchaincl/sqlhooks.svg)](https://travis-ci.org/gchaincl/sqlhooks) [![Coverage Status](https://coveralls.io/repos/github/gchaincl/sqlhooks/badge.svg?branch=master)](https://coveralls.io/github/gchaincl/sqlhooks?branch=master)

Attach hooks to any database/sql driver.

The purpose of sqlhooks is to provide a way to instrument your sql statements, making really easy to log queries or measure execution time without modifying your actual code.

# Install
```bash
go get github.com/gchaincl/sqlhooks
```

# Usage [![GoDoc](https://godoc.org/github.com/gchaincl/dotsql?status.svg)](https://godoc.org/github.com/gchaincl/sqlhooks)
```go
	package main

	import (
		"log"
		"time"

		"github.com/gchaincl/sqlhooks"
		_ "github.com/mattn/go-sqlite3"
	)

	// Hooks satisfies sqlhooks.Queryer interface
	type Hooks struct {
		count int
	}

	func (h *Hooks) BeforeQuery(ctx *sqlhooks.Context) error {
		h.count++
		ctx.Set("t", time.Now())
		ctx.Set("id", h.count)
		log.Printf("[query#%d] %s, args: %v", ctx.Get("id").(int), ctx.Query, ctx.Args)
		return nil
	}

	func (h *Hooks) AfterQuery(ctx *sqlhooks.Context) error {
		d := time.Since(ctx.Get("t").(time.Time))
		log.Printf("[query#%d] took %s (err: %v)", ctx.Get("id").(int), d, ctx.Error)
		return ctx.Error
	}

	func main() {
		hooks := &Hooks{}

		// Connect to attached driver
		db, _ := sqlhooks.Open("sqlite3", ":memory:", hooks)

		// Do you're stuff
		db.Exec("CREATE TABLE t (id INTEGER, text VARCHAR(16))")
		db.Exec("INSERT into t (text) VALUES(?), (?)", "foo", "bar")
		db.Query("SELECT id, text FROM t")
		db.Query("Invalid Query")
	}

```

```
2016/06/02 14:28:24 [query#1] SELECT id, text FROM t, args: []
2016/06/02 14:28:24 [query#1] took 122.406µs (err: <nil>)
2016/06/02 14:28:24 [query#2] Invalid Query, args: []
2016/06/02 14:28:24 [query#2] took 23.148µs (err: near "Invalid": syntax error)
```

# Benchmark
```
PASS
BenchmarkExec-4                    	  500000	      4604 ns/op
BenchmarkExecWithSQLHooks-4        	  300000	      5726 ns/op
BenchmarkPreparedExec-4            	 1000000	      1820 ns/op
BenchmarkPreparedExecWithSQLHooks-4	 1000000	      2088 ns/op
```
