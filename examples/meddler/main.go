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
	count int
}

func (mq *MyQueyer) BeforeQuery(ctx *sqlhooks.Context) error {
	mq.count++

	ctx.Set("id", mq.count)
	log.Printf("[query#%d] %s %q", ctx.Get("id").(int), ctx.Query, ctx.Args)
	return nil
}

func (mq MyQueyer) AfterQuery(ctx *sqlhooks.Context) error {
	log.Printf("[query#%d] done (err = %v)", ctx.Get("id").(int), ctx.Error)
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
