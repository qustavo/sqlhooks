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

		// Register the driver
		// "sqlite-hooked" is the attached driver, and "sqlite3" is where we're attaching to
		sqlhooks.Register("sqlite-hooked", sqlhooks.NewDriver("sqlite3", &hooks))

		// Connect to attached driver
		db, _ := sql.Open("sqlite-hooked", ":memory:")

		// Do you're stuff
		db.Exec("CREATE TABLE t (id INTEGER, text VARCHAR(16))")
		db.Exec("INSERT into t (text) VALUES(?), (?))", "foo", "bar")
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

// Register will register the driver using sql.Register()
func Register(name string, driver Driver) {
	sql.Register(name, &driver)
}

// Open Register a sqlhook driver and opens a connection against it
// driverName is the driver where we're attaching to
func Open(driverName, dsn string, hooks *Hooks) (*sql.DB, error) {
	registeredName := fmt.Sprintf("sqlhooks:%d", time.Now().UnixNano())
	sql.Register(registeredName, NewDriver(driverName, hooks))

	return sql.Open(registeredName, dsn)
}
