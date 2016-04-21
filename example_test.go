package sqlhooks

import (
	"database/sql"
	"log"
	"time"
)

func ExampleNewDriver() {
	hooks := Hooks{
		// This hook will log the query
		Query: func(query string, args ...interface{}) func() {
			// Log the query
			log.Println("Query: ", query)

			// Query is done
			return func() { log.Println("Query done!") }
		},
		// This hook will measure exec time
		Exec: func(query string, args ...interface{}) func() {
			t := time.Now()
			return func() {
				log.Println("Exec took: %s\n", time.Since(t))
			}
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
