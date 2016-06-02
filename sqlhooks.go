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

// Beginner is the interface implemented by objects that wants to hook to Begin function
type Beginner interface {
	BeforeBegin(c *Context) error
	AfterBegin(c *Context) error
}

// Commiter is the interface implemented by objects that wants to hook to Commit function
type Commiter interface {
	BeforeCommit(c *Context) error
	AfterCommit(c *Context) error
}

// Rollbacker is the interface implemented by objects that wants to hook to Rollback function
type Rollbacker interface {
	BeforeRollback(c *Context) error
	AfterRollback(c *Context) error
}

// Stmter is the interface implemented by objects that wants to hook to Statement related functions
type Stmter interface {
	BeforePrepare(*Context) error
	AfterPrepare(*Context) error

	BeforeStmtQuery(*Context) error
	AfterStmtQuery(*Context) error

	BeforeStmtExec(*Context) error
	AfterStmtExec(*Context) error
}

// Queryer is the interface implemented by objects that wants to hook to Query function
type Queryer interface {
	BeforeQuery(c *Context) error
	AfterQuery(c *Context) error
}

// Execer is the interface implemented by objects that wants to hook to Exec function
type Execer interface {
	BeforeExec(c *Context) error
	AfterExec(c *Context) error
}
