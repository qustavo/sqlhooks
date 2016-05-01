package sqlhooks

import (
	"database/sql"
	"testing"
)

func init() {
	sql.Register("sqlhooks", NewDriver("test", &Hooks{
		Exec: func(string, ...interface{}) func(error) {
			return func(error) {
			}
		},
	}))
}

func newDB(b *testing.B, driver string) *sql.DB {
	db, err := sql.Open(driver, "db")
	if err != nil {
		b.Fatalf("Open: %v", err)
	}

	if _, err := db.Exec("WIPE"); err != nil {
		b.Fatalf("WIPE: %v", err)
	}

	if _, err := db.Exec("CREATE|t|f1=string"); err != nil {
		b.Fatalf("CREATE: %v", err)
	}

	return db
}

func BenchmarkExec(b *testing.B) {
	db := newDB(b, "test")
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT|t|f1=?", "xxx")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecWithSQLHooks(b *testing.B) {
	db := newDB(b, "sqlhooks")
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT|t|f1=?", "xxx")
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkPreparedExec(b *testing.B) {
	db := newDB(b, "test")
	stmt, err := db.Prepare("INSERT|t|f1=?")
	if err != nil {
		b.Fatalf("prepare: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := stmt.Exec("xxx"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPreparedExecWithSQLHooks(b *testing.B) {
	db := newDB(b, "sqlhooks")
	stmt, err := db.Prepare("INSERT|t|f1=?")
	if err != nil {
		b.Fatalf("prepare: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := stmt.Exec("xxx"); err != nil {
			b.Fatal(err)
		}
	}
}
