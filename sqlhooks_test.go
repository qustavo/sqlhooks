package sqlhooks

import (
	"database/sql"
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	driverFlag = flag.String("driver", "test", "SQL Driver")
	dsnFlag    = flag.String("dsn", "db", "DSN")
)

type ops struct {
	wipe        string
	create      string
	insert      string
	selectwhere string
	selectall   string
}

var queries = make(map[string]ops)

func openDBWithHooks(t *testing.T, hooks interface{}, dsnArgs ...string) *sql.DB {
	q := queries[*driverFlag]

	dsn := *dsnFlag
	for _, arg := range dsnArgs {
		dsn = dsn + arg
	}

	// First, we connect directly using `test` driver
	if db, err := sql.Open(*driverFlag, dsn); err != nil {
		t.Fatalf("sql.Open: %v", err)
		return nil
	} else {
		if _, err := db.Exec(q.wipe); err != nil {
			t.Fatalf("WIPE: %v", err)
			return nil
		}

		if _, err := db.Exec(q.create); err != nil {
			t.Fatalf("CREATE: %v", err)
			return nil
		}
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close: %v", err)
		}
	}

	db, err := Open(*driverFlag, dsn, hooks)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
		return nil
	}

	return db
}

func TestBeforeAndAfterHooks(t *testing.T) {
	q := queries[*driverFlag]

	for _, hook := range []string{"Query", "Exec", "Begin", "Commit", "Rollback"} {
		beforeOk := false
		before := func(ctx *Context) error {
			beforeOk = true
			return nil
		}

		afterOk := false
		after := func(ctx *Context) error {
			afterOk = true
			return ctx.Error
		}

		hooks := NewHooksMock(before, after)
		db := openDBWithHooks(t, hooks)

		switch hook {
		case "Query":
			db.Query(q.selectall)
		case "Exec":
			db.Exec(q.insert)
		case "Begin":
			tx, _ := db.Begin()

			hooks.beforeCommit = nil
			hooks.afterCommit = nil
			tx.Commit()
		case "Commit":
			hooks.beforeBegin = nil
			hooks.afterBegin = nil

			tx, _ := db.Begin()
			tx.Commit()
		case "Rollback":
			hooks.beforeBegin = nil
			hooks.afterBegin = nil

			tx, _ := db.Begin()
			tx.Rollback()
		}

		assert.True(t, beforeOk, "'Before%s' hook didn't run", hook)
		assert.True(t, afterOk, "'After%s' hook didn't run", hook)
	}
}

func TestBeforeQueryStopsAndReturnsError(t *testing.T) {
	q := queries[*driverFlag]

	for _, hook := range []string{"Query", "Exec", "Begin", "Commit", "Rollback"} {
		someErr := fmt.Errorf("Some Error")
		before := func(ctx *Context) error {
			return someErr
		}

		// this hook should never run
		after := func(ctx *Context) error {
			assert.True(t, false, "'After%s' should not run", hook)
			return nil
		}

		hooks := NewHooksMock(before, after)
		db := openDBWithHooks(t, hooks)

		var err error
		switch hook {
		case "Query":
			_, err = db.Query(q.selectall)
		case "Exec":
			_, err = db.Exec(q.insert)
		case "Begin":
			var tx *sql.Tx
			tx, err = db.Begin()
			assert.Nil(t, tx)
		case "Commit":
			hooks.beforeBegin = nil
			hooks.afterBegin = nil
			tx, _ := db.Begin()

			err = tx.Commit()
		case "Rollback":
			hooks.beforeBegin = nil
			hooks.afterBegin = nil
			tx, _ := db.Begin()

			err = tx.Rollback()
		}

		assert.Equal(t, someErr, err, "On %s hooks", hook)
	}
}

func TestBeforeModifiesQueryAndArgs(t *testing.T) {
	if *driverFlag == "test" {
		t.SkipNow()
	}

	q := queries[*driverFlag]

	// this hook convert the select where into a select all
	before := func(ctx *Context) error {
		ctx.Args = nil
		ctx.Query = q.selectall
		return nil
	}

	after := func(ctx *Context) error {
		assert.Equal(t, q.selectall, ctx.Query)
		assert.Equal(t, []interface{}(nil), ctx.Args)
		return ctx.Error
	}

	hooks := &HooksMock{
		beforeQuery: before,
		afterQuery:  after,
	}
	db := openDBWithHooks(t, hooks)

	db.Exec(q.insert, "x", "y")
	rows, err := db.Query(q.selectwhere, "a", "b")
	assert.NoError(t, err)

	found := false
	for rows.Next() {
		found = true
	}

	assert.True(t, found)
}

func TestBeforePrepare(t *testing.T) {
	q := queries[*driverFlag]

	before := func(ctx *Context) error {
		ctx.Query = q.selectall
		return nil
	}

	db := openDBWithHooks(t, &HooksMock{beforePrepare: before})

	_, err := db.Prepare("invalid query")
	assert.NoError(t, err)
}

func TestAfterReceivesAndHideTheError(t *testing.T) {
	for _, hook := range []string{"Query", "Exec"} {
		after := func(ctx *Context) error {
			assert.Error(t, ctx.Error)
			return nil // hide the error
		}

		db := openDBWithHooks(t, &HooksMock{
			afterQuery: after,
			afterExec:  after,
		})

		var err error
		switch hook {
		case "Query":
			_, err = db.Query("invalid query")
		case "Exec":
			_, err = db.Exec("invalid query")
		}
		assert.NoError(t, err)
	}
}

func TestDriverItWorksWithNilHooks(t *testing.T) {
	q := queries[*driverFlag]

	db := openDBWithHooks(t, nil)

	for _ = range [10]bool{} {
		_, err := db.Exec(q.insert, "foo", "bar")
		assert.NoError(t, err)
	}

	rows, err := db.Query(q.selectall)
	assert.NoError(t, err)

	items := 0
	for rows.Next() {
		items++
	}

	assert.Equal(t, 10, items)
}

func TestValuesAreSavedAndRetrievedFromCtx(t *testing.T) {
	q := queries[*driverFlag]

	before := func(ctx *Context) error {
		ctx.Set("foo", 123)
		ctx.Set("bar", "sqlhooks")
		return nil
	}

	after := func(ctx *Context) error {
		assert.Equal(t, 123, ctx.Get("foo").(int))
		assert.Equal(t, "sqlhooks", ctx.Get("bar").(string))
		return ctx.Error
	}

	hooks := NewHooksMock(before, after)
	db := openDBWithHooks(t, hooks)

	_, err := db.Query(q.selectall)
	assert.NoError(t, err)
}

func TestDriverIsNotRegisteredTwice(t *testing.T) {
	registeredDrivers := sql.Drivers()

	for i := 0; i < 100; i++ {
		_, err := Open("test", "db", nil)
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
	}

	registeredAfterOpen := len(sql.Drivers()) - len(registeredDrivers)
	if registeredAfterOpen > 1 {
		t.Errorf("Driver registered %d times more than expected", registeredAfterOpen-1)
	}
}
