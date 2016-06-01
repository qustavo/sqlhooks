/*
Package sqlhooks Attach hooks to any database/sql driver.

Sqlhooks provides a mechanism to execute a callbacks around specific database/sql functions.

The purpose of sqlhooks is to provide anway to instrument your sql statements,
making really easy to log queries or measure execution time without modifying your actual code.

Example:
	package main

	import (
		"log"
		"time"

		"github.com/gchaincl/sqlhooks"
		_ "github.com/mattn/go-sqlite3"
	)

	// Hooks satisfies sqlhooks.Queryer interface
	type Hooks struct {
	}

	func (h Hooks) BeforeQuery(ctx *sqlhooks.Context) error {
		log.Printf("[query#%s] %s, args: %v", ctx.GetID(), ctx.Query, ctx.Args)
		ctx.Set("t", time.Now())
		return nil
	}

	func (h Hooks) AfterQuery(ctx *sqlhooks.Context) error {
		d := time.Since(ctx.Get("t").(time.Time))
		log.Printf("[query#%s] took %s (err: %v)", ctx.GetID(), d, ctx.Error)
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
*/
package sqlhooks

import (
	"database/sql"
	"fmt"
	"time"
)

var (
	drivers = make(map[interface{}]string)
)

// Open Register a sqlhook driver and opens a connection against it
// driverName is the driver where we're attaching to
func Open(driverName, dsn string, hooks interface{}) (*sql.DB, error) {
	if registeredName, ok := drivers[hooks]; ok {
		return sql.Open(registeredName, dsn)
	}

	registeredName := fmt.Sprintf("sqlhooks:%d", time.Now().UnixNano())
	sql.Register(registeredName, NewDriver(driverName, hooks))
	drivers[hooks] = registeredName

	return sql.Open(registeredName, dsn)
}
