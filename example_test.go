package sqlhooks

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"time"
)

func ExampleNewDriver() {
	hooks := Hooks{
		// This hook will log the query
		Query: func(fn QueryFn, query string, args ...interface{}) (driver.Rows, error) {
			// Log the query
			log.Println("Query: ", query)

			// Run Query
			r, err := fn()

			// Query is done
			log.Println("Query done!")
			return r, err
		},
		// This hook will measure exec time
		Exec: func(fn ExecFn, query string, args ...interface{}) (driver.Result, error) {
			t := time.Now()
			r, err := fn()
			log.Println("Exec took: %s\n", time.Since(t))
			return r, err
		},
	}

	// mysql is the driver we're going to attach to
	NewDriver("mysql", &hooks)
}

func ExampleRegister() {
	// Register the driver under `hooked-mysql` name
	Register("hooked-mysql", NewDriver("mysql", &Hooks{}))

	// Open a connection
	sql.Open("hooked-mysql", "/db")
}
