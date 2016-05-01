package sqlhooks

import (
	"database/sql"
	"log"
	"time"
)

func ExampleNewDriver() {
	hooks := Hooks{
		// This hook will log the query
		Query: func(query string, args ...interface{}) func(error) {
			// Log the query
			log.Println("Query: ", query)

			// Query is done
			return func(err error) {
				if err != nil {
					log.Printf("%v", err)
				} else {
					log.Printf("query ok!")
				}
			}
		},
		// This hook will measure exec time
		Exec: func(query string, args ...interface{}) func(error) {
			t := time.Now()
			return func(err error) {
				log.Printf("Exec took: %s (%v)\n", time.Since(t), err)
			}
		},
	}

	// mysql is the driver we're going to attach to
	NewDriver("mysql", &hooks)
}

func ExampleRegister() {
	// Register the driver under `hooked-mysql` name
	sql.Register("hooked-mysql", NewDriver("mysql", &Hooks{}))

	// Open a connection
	sql.Open("hooked-mysql", "/db")
}

func ExampleOpen() {
	db, err := Open("mysql", "user:pass@/db", &Hooks{
		Query: func(query string, args ...interface{}) func(error) {
			log.Println(query)
			return nil
		},
	})

	if err != nil {
		panic(err)
	}

	db.Query("SELECT 1+1")
}
