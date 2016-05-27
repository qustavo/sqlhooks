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

type MyQueyer struct {
}

func (mq MyQueyer) BeforeQuery(ctx *sqlhooks.Context) error {
	log.Printf("[query#%s] %s %q", ctx.GetID(), ctx.Query, ctx.Args)
	return nil
}

func (mq MyQueyer) AfterQuery(ctx *sqlhooks.Context) error {
	log.Printf("[query#%s] done (err = %v)", ctx.GetID(), ctx.Error)
	return ctx.Error
}

func main() {
	db, err := sqlhooks.Open("sqlite3", ":memory:", &MyQueyer{})
	if err != nil {
		panic(err)
	}

	p := new(Person)
	meddler.Load(db, "person", p, 1)
}
