# sqlhooks [![Build Status](https://travis-ci.org/gchaincl/sqlhooks.svg)](https://travis-ci.org/gchaincl/sqlhooks) [![Coverage Status](https://coveralls.io/repos/github/gchaincl/sqlhooks/badge.svg?branch=master)](https://coveralls.io/github/gchaincl/sqlhooks?branch=master)

Attach hooks to any database/sql driver.

The purpose of sqlhooks is to provide anway to instrument your sql statements, making really easy to log queries or measure execution time without modifying your actual code.

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
		Exec: func(query string, args ...interface{}) func() {
			log.Printf("[exec] %s, args: %v", query, args)
			return nil
		},
		Query: func(query string, args ...interface{}) func() {
			t := time.Now()
			id := t.Nanosecond()
			log.Printf("[query#%d] %s, args: %v", id, query, args)
				// This will be executed when Query statements has completed
			return func() {
				log.Printf("[query#%d] took: %s\n", id, time.Since(t))
			}
		},
	}

	// Register the driver
	// "sqlite-hooked" is the attached driver, and "sqlite3" is where we're attaching to
	sqlhooks.Register("sqlite-hooked", sqlhooks.NewDriver("sqlite3", &hooks))

	// Connect to attached driver
	db, _ := sql.Open("sqlite-hooked", ":memory:")

	// Do you're stuff
	db.Exec("CREATE TABLE t (id INTEGER, text VARCHAR(16))")
	db.Exec("INSERT into t (text) VALUES(?), (?))", "foo", "bar")
	db.Query("SELECT id, text FROM t")
}
```

sqlhooks will intercept Query and Exec functions and instead run your hooks, output will look like:
```
2016/04/21 18:18:06 [exec] CREATE TABLE t (id INTEGER, text VARCHAR(16)), args: []
2016/04/21 18:18:06 [exec] INSERT into t (text) VALUES(?), (?)), args: [foo bar]
2016/04/21 18:18:06 [query#912806039] SELECT id, text FROM t, args: []
2016/04/21 18:18:06 [query#912806039] took: 32.425µs
```

# TODO
- [ ] `Hooks{}` should be an interface instead of a struct
- [ ] Exec and Query hooks should return `(func(), error)`
- [ ] Arguments should be pointers so queries can be modified
- [x] Implement hooks on Tx
