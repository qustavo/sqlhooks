/*
Package sqlhooks Attach hooks to any database/sql driver.

The purpose of sqlhooks is to provide anway to instrument your sql statements,
making really easy to log queries or measure execution time without modifying your actual code.

Example:
	package main

	import (
		"database/sql"
		"log"
		"time"

		"github.com/gchaincl/sqlhooks"
		_ "github.com/mattn/go-sqlite3"
	)


	func main() {
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

*/
package sqlhooks

import (
	"database/sql"
	"database/sql/driver"
)

func convertArgs(args []driver.Value) []interface{} {
	r := make([]interface{}, len(args))
	for i, arg := range args {
		r[i] = arg
	}
	return r
}

type stmt struct {
	driver.Stmt
	query string
	hooks *Hooks
}

func (s stmt) Close() error {
	return s.Stmt.Close()
}

func (s stmt) Exec(args []driver.Value) (res driver.Result, err error) {
	if hook := s.hooks.Exec; hook != nil {
		fn := hook(s.query, convertArgs(args)...)
		if fn != nil {
			defer fn()
		}
	}

	return s.Stmt.Exec(args)
}

func (s stmt) NumInput() int {
	return s.Stmt.NumInput()
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	if hook := s.hooks.Query; hook != nil {
		fn := hook(s.query, convertArgs(args)...)
		if fn != nil {
			defer fn()
		}
	}

	return s.Stmt.Query(args)
}

type conn struct {
	driver.Conn
	hooks *Hooks
}

func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		if hook := c.hooks.Query; hook != nil {
			fn := hook(query, convertArgs(args)...)
			if fn != nil {
				defer fn()
			}
		}

		return queryer.Query(query, args)
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.Conn.(driver.Execer); ok {
		if hook := c.hooks.Exec; hook != nil {
			fn := hook(query, convertArgs(args)...)
			if fn != nil {
				defer fn()
			}
		}

		return execer.Exec(query, args)
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Prepare(query string) (driver.Stmt, error) {
	_stmt, err := c.Conn.Prepare(query)
	return stmt{_stmt, query, c.hooks}, err
}

func (c conn) Close() error {
	return c.Conn.Close()
}

func (c conn) Begin() (driver.Tx, error) {
	return c.Conn.Begin()
}

// Hooks contains hook functions for instrumenting Query and Exec
// Returned func() will be executed after statements have completed
type Hooks struct {
	Exec  func(string, ...interface{}) func()
	Query func(string, ...interface{}) func()
}

// Driver it's a proxy for a specific sql driver
type Driver struct {
	driver driver.Driver
	name   string
	hooks  *Hooks
}

// NewDriver will create a Proxy Driver with defined Hooks
// name is the underlying driver name
func NewDriver(name string, hooks *Hooks) Driver {
	return Driver{name: name, hooks: hooks}
}

// Open returns a new connection to the database, using the underlying specified driver
func (d *Driver) Open(dsn string) (driver.Conn, error) {
	if d.driver == nil {
		// Get Driver by Opening a new connection
		db, err := sql.Open(d.name, dsn)
		if err != nil {
			return nil, err
		}
		if err := db.Close(); err != nil {
			return nil, err
		}
		d.driver = db.Driver()
	}

	_conn, err := d.driver.Open(dsn)
	return conn{_conn, d.hooks}, err
}

// Register will register the driver using sql.Register()
func Register(name string, driver Driver) {
	sql.Register(name, &driver)
}
