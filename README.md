# sqlhooks [![Build Status](https://travis-ci.org/gchaincl/sqlhooks.svg)](https://travis-ci.org/gchaincl/sqlhooks) [![Coverage Status](https://coveralls.io/repos/gchaincl/sqlhooks/badge.svg?branch=coveralls&service=github)](https://coveralls.io/github/gchaincl/sqlhooks?branch=coveralls)
Attach hooks any database/sql driver

# Install
```bash
go get github.com/gchaincl/sqlhooks
```

# Usage
```go
package main

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"time"

	"github.com/gchaincl/sqlhooks"
	_ "github.com/mattn/go-sqlite3"
)


func main() {
  // Define your hooks
  hooks := sqlhooks.Hooks{
		Query: func(fn sqlhooks.QueryFn, query string, args ...interface{}) (driver.Rows, error) {
			defer func(t time.Time) {
				log.Printf("query: %s, args: %v, took: %s\n", query, args, time.Since(t))
			}(time.Now())

			return fn()
		},
		Exec: func(fn sqlhooks.ExecFn, query string, args ...interface{}) (driver.Result, error) {
			defer func(t time.Time) {
				log.Printf("exec: %s, args: %v, took: %s\n", query, args, time.Since(t))
			}(time.Now())

			return fn()
		},
	}

	// Register the driver
	// sqlite-hooked is the attached driver, and sqlite3 is where we're attaching to
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
2000/01/01 00:01:02 exec: CREATE TABLE t (id INTEGER, text VARCHAR(16)), args: [], took: 226.169µs
2000/01/01 00:01:02 exec: INSERT into t (text) VALUES(?), (?)), args: [foo bar], took: 26.822µs
2000/01/01 00:01:02 query: SELECT id, text FROM t, args: [], took: 20.229µs
```

