package main

import (
	"log"
	"strings"

	"github.com/gchaincl/sqlhooks"
	"github.com/go-gorp/gorp"
	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	// db tag lets you specify the column name if it differs from the struct field
	Id    int64  `db:"post_id"`
	Title string `db:",size:50"`               // Column size set to 50
	Body  string `db:"article_body,size:1024"` // Set both column name and size
}

type Hooks struct{}

func (h Hooks) BeforeQuery(ctx *sqlhooks.Context) error {
	log.Println(ctx.Query, ctx.Args)
	return nil
}

func (h Hooks) AfterQuery(ctx *sqlhooks.Context) error {
	return ctx.Error
}

// Update Post's title field Before Inserting
func (h Hooks) BeforeExec(ctx *sqlhooks.Context) error {
	if strings.HasPrefix(ctx.Query, `insert into "posts"`) {
		ctx.Args[0] = "[updated] " + ctx.Args[0].(string)
	}
	return nil
}

func (h Hooks) AfterExec(ctx *sqlhooks.Context) error {
	return ctx.Error
}

func main() {
	db, err := sqlhooks.Open("sqlite3", ":memory:", &Hooks{})
	if err != nil {
		panic(err)
	}

	dbmap := gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(Post{}, "posts").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		panic(err)
	}

	dbmap.Insert(
		&Post{Title: "Foo", Body: "Some Content"},
		&Post{Title: "Bar", Body: "More Content"},
	)
	post := Post{}
	p, err := dbmap.Get(&post, 1)
	if err != nil {
		panic(err)
	}

	log.Printf("%#v", p)
}
