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
	"database/sql"
	"log"
	"time"

	"github.com/gchaincl/sqlhooks"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Define your hooks
	// They will print execution time
	hooks := sqlhooks.Hooks{
		Exec: func(query string, args ...interface{}) func(error) {
			log.Printf("[exec] %s, args: %v", query, args)
			return nil
		},
		Query: func(query string, args ...interface{}) func(error) {
			t := time.Now()
			id := t.Nanosecond()
			log.Printf("[query#%d] %s, args: %v", id, query, args)
			// This will be executed when Query statements has completed
			return func(err error) {
				log.Printf("[query#%d] took: %s (err: %v)", id, time.Since(t), err)
			}
		},
	}

	// Connect to hooked sqlite3 driver
	db, err := sqlhooks.Open("sqlite3", ":memory:", &hooks)
	if err != nil {
		panic(err)
	}

	// Do you're stuff
	db.Exec("CREATE TABLE t (id INTEGER, text VARCHAR(16))")
	db.Exec("INSERT into t (text) VALUES(?), (?))", "foo", "bar")
	db.Query("SELECT id, text FROM t")
	db.Query("Invalid Query")
}
```

sqlhooks will intercept Query and Exec functions, run your hooks, execute que queries and finally execute the returned func(). Output will look like:
```
2016/04/23 19:43:53 [exec] CREATE TABLE t (id INTEGER, text VARCHAR(16)), args: []
2016/04/23 19:43:53 [exec] INSERT into t (text) VALUES(?), (?)), args: [foo bar]
2016/04/23 19:43:53 [query#487301557] SELECT id, text FROM t, args: []
2016/04/23 19:43:53 [query#487301557] took: 37.765µs (err: <nil>)
2016/04/23 19:43:53 [query#487405691] Invalid Query, args: []
2016/04/23 19:43:53 [query#487405691] took: 18.312µs (err: near "Invalid": syntax error)
```

# Benchmark
```
BenchmarkExec-4                    	  500000	      4335 ns/op	     566 B/op	      16 allocs/op
BenchmarkExecWithSQLHooks-4        	  500000	      4918 ns/op	     646 B/op	      19 allocs/op
BenchmarkPreparedExec-4            	 1000000	      1884 ns/op	     181 B/op	       7 allocs/op
BenchmarkPreparedExecWithSQLHooks-4	 1000000	      1919 ns/op	     197 B/op	       8 allocs/op
```

# TODO
- [ ] `Hooks{}` should be an interface instead of a struct
- [x] Exec and Query hooks should return `func(error)`
- [ ] Arguments should be pointers so queries can be modified
- [x] Implement hooks on Tx
