package sqlhooks

import (
	"database/sql"
	"flag"
	"fmt"
	"sort"
	"testing"
	"time"
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

func openDBWithHooks(t *testing.T, hooks *Hooks) *sql.DB {
	q := queries[*driverFlag]

	// First, we connect directly using `test` driver
	if db, err := sql.Open(*driverFlag, *dsnFlag); err != nil {
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
	}

	// Now, return a db handler using hooked driver
	driver := NewDriver(*driverFlag, hooks)
	driverName := fmt.Sprintf("sqlhooks-%d", time.Now().UnixNano())
	Register(driverName, driver)

	db, err := sql.Open(driverName, *dsnFlag)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
		return nil
	}

	return db
}

func TestHooks(t *testing.T) {
	q := queries[*driverFlag]
	tests := []struct {
		op    string
		query string
		args  []interface{}
	}{
		{"exec", q.insert, []interface{}{"foo", "bar"}},
		{"query", q.selectwhere, []interface{}{"foo", "bar"}},
		{"query", q.selectall, []interface{}{}},
		{"stmt.query", q.selectall, nil},
		{"stmt.exec", q.insert, []interface{}{"x", "y"}},
		{"tx.query", q.selectall, nil},
		{"tx.exec", q.insert, []interface{}{"x", "y"}},
	}

	for _, test := range tests {
		done := false
		assert := func(query string, args ...interface{}) func() {
			// Query Assertions
			if query != test.query {
				t.Errorf("query = `%s`, expected `%s`", query, test.query)
			}

			if test.args == nil {
				test.args = []interface{}{}
			}

			// Arguments assertions
			if len(args) != len(test.args) {
				t.Errorf("Expected args: %d, got %d", len(test.args), len(args))
			}
			for i, _ := range test.args {
				if args[i] != test.args[i] {
					t.Errorf("%s: arg[%d] == %#v, got %#v", test.op, i, test.args[i], args[i])
				}
			}

			return func() {
				done = true
			}
		}
		db := openDBWithHooks(t, &Hooks{Query: assert, Exec: assert})

		switch test.op {
		case "query":
			if _, err := db.Query(test.query, test.args...); err != nil {
				t.Errorf("query: %v", err)
			}
		case "exec":
			if _, err := db.Exec(test.query, test.args...); err != nil {
				t.Errorf("exec: %v", err)
			}
		case "stmt.query":
			if stmt, err := db.Prepare(test.query); err != nil {
				t.Errorf("prepare: %v", err)
			} else {
				if _, err := stmt.Query(test.args...); err != nil {
					t.Errorf("prepared query: %v", err)
				}
			}
		case "stmt.exec":
			if stmt, err := db.Prepare(test.query); err != nil {
				t.Errorf("prepare: %v", err)
			} else {
				if _, err := stmt.Exec(test.args...); err != nil {
					t.Errorf("prepared exec: %v", err)
				}
			}
		case "tx.query":
			tx, err := db.Begin()
			if err != nil {
				t.Errorf("[%s] begin: %v", test.op, err)
			}
			if _, err := tx.Query(test.query, test.args...); err != nil {
				t.Errorf("[%s] query: %v", test.op, err)
			}
			if err := tx.Commit(); err != nil {
				t.Errorf("[%s] commit: %v", test.op, err)
			}
		case "tx.exec":
			tx, err := db.Begin()
			if err != nil {
				t.Errorf("[%s] begin: %v", test.op, err)
			}
			if _, err := tx.Exec(test.query, test.args...); err != nil {
				t.Errorf("[%s] exec: %v", test.op, err)
			}
			if err := tx.Commit(); err != nil {
				t.Errorf("[%s] commit: %v", test.op, err)
			}
		}

		if done == false {
			t.Errorf("Expected %s cancelation to be completed", test.op)
		}

	}
}

func TestEmptyHooks(t *testing.T) {
	q := queries[*driverFlag]
	db := openDBWithHooks(t, &Hooks{})

	if _, err := db.Exec(q.insert, "foo", "bar"); err != nil {
		t.Fatalf("Exec: %v\n", err)
	}

	if _, err := db.Query(q.selectall); err != nil {
		t.Fatalf("Query: %v\n", err)
	}
}

func TestCreateInsertAndSelect(t *testing.T) {
	q := queries[*driverFlag]
	db := openDBWithHooks(t, &Hooks{})

	db.Exec(q.insert, "a", "1")
	db.Exec(q.insert, "b", "2")
	db.Exec(q.insert, "c", "3")

	rows, _ := db.Query(q.selectall)
	var fs []string
	for rows.Next() {
		var f1 string
		var f2 string
		rows.Scan(&f1, &f2)
		fs = append(fs, f1)
	}
	sort.Strings(fs)
	if len(fs) != 3 {
		t.Fatalf("Expected 3 rows, got: %d", len(fs))
	}

	for i, e := range []string{"a", "b", "c"}[:len(fs)] {
		f := fs[i]
		if f != e {
			t.Errorf("f1 = `%s`, expected: `%s`", f, e)
		}
	}
}

func TestTXs(t *testing.T) {
	for _, op := range []string{"commit", "rollback"} {
		ids := struct {
			begin string
			end   string
		}{}

		db := openDBWithHooks(t, &Hooks{
			Begin: func(id string) {
				ids.begin = id
			},
			Commit: func(id string) {
				ids.end = id
			},
			Rollback: func(id string) {
				ids.end = id
			},
		})

		tx, err := db.Begin()
		if err != nil {
			t.Errorf("begin: %v", err)
			continue
		}

		switch op {
		case "commit":
			if err := tx.Commit(); err != nil {
				t.Errorf("commit: %v", err)
			}
		case "rollback":
			if err := tx.Rollback(); err != nil {
				t.Errorf("rollback: %v", err)
			}
		}

		if ids.begin == "" {
			t.Errorf("Expected id to be != ''")
		}

		if ids.begin != ids.end {
			t.Errorf("Expected equals ids, got '%s = %s'", ids.begin, ids.end)
		}
	}
}
