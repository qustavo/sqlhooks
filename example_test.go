package sqlhooks

import "database/sql"

type MyQueryer struct{}

func (q MyQueryer) BeforeQuery(ctx *Context) error {
	return nil
}

func (q MyQueryer) AfterQuery(ctx *Context) error {
	return nil
}

func ExampleNewDriver() {
	// MyQueryer satisfies Queryer interface
	hooks := MyQueryer{}

	// mysql is the driver we're going to attach to
	driver := NewDriver("mysql", &hooks)
	sql.Register("sqlhooks-mysql", driver)
}

func ExampleOpen() {
	db, err := Open("mysql", "user:pass@/db", nil)
	if err != nil {
		panic(err)
	}

	db.Query("SELECT 1+1")
}
