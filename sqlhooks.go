/*
Package Sqlhooks provides a mechanism to execute a callbacks around specific database/sql functions.

The purpose of sqlhooks is to provide a way to instrument your database operations,
making really to log queries and arguments, measure execution time,
modifies queries before the are executed or stop execution if some conditions are met.

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
func Open(driverName, dsn string, hooks HookType) (*sql.DB, error) {
	if registeredName, ok := drivers[hooks]; ok {
		return sql.Open(registeredName, dsn)
	}

	registeredName := fmt.Sprintf("sqlhooks:%d", time.Now().UnixNano())
	sql.Register(registeredName, NewDriver(driverName, hooks))
	drivers[hooks] = registeredName

	return sql.Open(registeredName, dsn)
}

/*
HookType is the type of Hook.
In order to reduce the amount boilerplate, it's organized by database operations,
so you can only implement the hooks you need for certain operation
This type is an alias for interface{}, however the hook should implement at least one of the following interfaces:
	- Beginner
	- Commiter
	- Rollbacker
	- Stmter
	- Queryer
	- Execer

Every hook can be attached Before or After the operation.
Before hooks are triggered just before execute the operation (Begin, Commit, Rollback, Prepare, Query, Exec),
if they returns an error, neither the operation nor the After hook will executed, and the error will be returned to the caller

After hooks are triggered after the operation complete, the there is an error it will be passed inside *Context.
The error returned by an After hook will override the error returned from the operation, that's why in most cases
an after hooks should:
	return ctx.Error

*/
type HookType interface{}

// Beginner is the interface implemented by objects that wants to hook to Begin function
type Beginner interface {
	BeforeBegin(*Context) error
	AfterBegin(*Context) error
}

// Commiter is the interface implemented by objects that wants to hook to Commit function
type Commiter interface {
	BeforeCommit(*Context) error
	AfterCommit(*Context) error
}

// Rollbacker is the interface implemented by objects that wants to hook to Rollback function
type Rollbacker interface {
	BeforeRollback(*Context) error
	AfterRollback(*Context) error
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
	BeforeQuery(*Context) error
	AfterQuery(*Context) error
}

// Execer is the interface implemented by objects that wants to hook to Exec function
type Execer interface {
	BeforeExec(*Context) error
	AfterExec(*Context) error
}
