package main

import (
	"log"

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

func main() {
	db, err := sqlhooks.Open("sqlite3", ":memory:", &sqlhooks.Hooks{
		Query: func(q string, a ...interface{}) func(error) {
			log.Println(q, a)
			return nil
		},
		Exec: func(q string, a ...interface{}) func(error) {
			log.Println(q, a)
			return nil
		},
	})
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
	dbmap.Get(&post, 1)
}
