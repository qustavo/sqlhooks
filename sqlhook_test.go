package sqlhook

import (
	"database/sql"
	"testing"
)

func TestHooks(t *testing.T) {
	expectedExec := "CREATE|t|f1=string"
	expectedQuery := "SELECT|t|f1|"

	hooks := Hooks{
		Query: func(query string, args ...interface{}) error {
			if query != expectedQuery {
				t.Errorf("query = `%s`, expected `%s`", expectedQuery)
			}
			return nil
		},
		Exec: func(query string, args ...interface{}) error {
			if query != expectedExec {
				t.Errorf("query = `%s`, expected `%s`", expectedExec)
			}
			return nil
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
