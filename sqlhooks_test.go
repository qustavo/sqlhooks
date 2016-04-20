package sqlhooks

import (
	"database/sql"
	"database/sql/driver"
	"sort"
	"testing"
)

func TestHooks(t *testing.T) {
	expectedExec := "CREATE|t|f1=string"
	expectedQuery := "SELECT|t|f1|"

	hooks := Hooks{
		Query: func(fn QueryFn, query string, args ...interface{}) (driver.Rows, error) {
			if query != expectedQuery {
				t.Errorf("query = `%s`, expected `%s`", query, expectedQuery)
			}
			return fn()
		},
		Exec: func(fn ExecFn, query string, args ...interface{}) (driver.Result, error) {
			if query != expectedExec {
				t.Errorf("query = `%s`, expected `%s`", query, expectedExec)
			}
			return fn()
		},
	}
	Register("test_1", NewDriver("test", &hooks))

	db, _ := sql.Open("test_1", "d1")
	db.Exec(expectedExec)
	db.Query(expectedQuery)

	execStmt, _ := db.Prepare(expectedExec)
	execStmt.Exec()

	queryStmt, _ := db.Prepare(expectedQuery)
	queryStmt.Query()
}

func TestEmptyHooks(t *testing.T) {
	Register("test_2", NewDriver("test", &Hooks{}))
	db, _ := sql.Open("test_2", "d2")

	if _, err := db.Exec("CREATE|t|f1=string"); err != nil {
		t.Fatalf("Exec: %v\n", err)
	}

	if _, err := db.Query("SELECT|t|f1|"); err != nil {
		t.Fatalf("Query: %v\n", err)
	}
}

func TestCreateInsertAndSelect(t *testing.T) {
	Register("test_3", NewDriver("test", &Hooks{}))
	db, _ := sql.Open("test_3", "d3")

	db.Exec("CREATE|t|f1=string")
	db.Exec("INSERT|t|f1=?", "a")
	db.Exec("INSERT|t|f1=?", "b")
	db.Exec("INSERT|t|f1=?", "c")

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
