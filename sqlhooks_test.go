package sqlhooks

import (
	"database/sql"
	"fmt"
	"sort"
	"testing"
	"time"
)

func openDBWithHooks(t *testing.T, hooks *Hooks) *sql.DB {
	// First, we connect directly using `test` driver
	if db, err := sql.Open("test", "db"); err != nil {
		t.Fatalf("sql.Open: %v", err)
		return nil
	} else {
		if _, err := db.Exec("WIPE"); err != nil {
			t.Fatalf("WIPE: %v", err)
			return nil
		}

		if _, err := db.Exec("CREATE|t|f1=string,f2=string"); err != nil {
			t.Fatalf("CREATE: %v", err)
			return nil
		}
	}

	// Now, return a db handler using hooked driver
	driver := NewDriver("test", hooks)
	driverName := fmt.Sprintf("test-%d", time.Now().UnixNano())
	Register(driverName, driver)

	db, err := sql.Open(driverName, "db")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
		return nil
	}

	return db
}

func TestHooks(t *testing.T) {
	tests := []struct {
		op    string
		query string
		args  []interface{}
	}{
		{"exec", "INSERT|t|f1=?,f2=?", []interface{}{"foo", "bar"}},
		{"query", "SELECT|t|f1|f1=?,f2=?", []interface{}{"foo", "bar"}},
		{"query", "SELECT|t|f1|", []interface{}{}},
		{"stmt.query", "SELECT|t|f1|", nil},
		{"stmt.exec", "INSERT|t|f1=?", []interface{}{"x"}},
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
					t.Errorf("arg[%d] == %#v, got %#v", i, test.args[i], args[i])
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
		}

		if done == false {
			t.Errorf("Expected %s cancelation to be completed", test.op)
		}

	}
}

func TestEmptyHooks(t *testing.T) {
	db := openDBWithHooks(t, &Hooks{})

	if _, err := db.Exec("INSERT|t|f1=?", "foo"); err != nil {
		t.Fatalf("Exec: %v\n", err)
	}

	if _, err := db.Query("SELECT|t|f1|"); err != nil {
		t.Fatalf("Query: %v\n", err)
	}
}

func TestCreateInsertAndSelect(t *testing.T) {
	db := openDBWithHooks(t, &Hooks{})

	db.Exec("INSERT|t|f1=?,f2=?", "a", "1")
	db.Exec("INSERT|t|f1=?,f2=?", "b", "2")
	db.Exec("INSERT|t|f1=?,f2=?", "c", "3")

	rows, _ := db.Query("SELECT|t|f1|")
	var fs []string
	for rows.Next() {
		var f string
		rows.Scan(&f)
		fs = append(fs, f)
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
