package main

import (
	"log"
	"time"

	"github.com/gchaincl/sqlhooks"
	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/meddler"
)

type Person struct {
	ID      int       `meddler:"id,pk"`
	Name    string    `meddler:"name"`
	Age     int       `meddler:"age"`
	Created time.Time `meddler:"created,localtime"`
}

func main() {
	db, err := sqlhooks.Open("sqlite3", ":memory:", &sqlhooks.Hooks{
		Query: func(q string, a ...interface{}) func(error) {
			log.Println("[query]", q, a)
			return nil
		},
	})
	if err != nil {
		panic(err)
	}

	p := new(Person)
	meddler.Load(db, "person", p, 1)
}
