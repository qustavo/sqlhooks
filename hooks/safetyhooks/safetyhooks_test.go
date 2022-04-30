package safetyhooks

import (
	"database/sql"
	"testing"

	"github.com/gchaincl/sqlhooks/v2"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T, hooks sqlhooks.Hooks) *sql.DB {
	var (
		err  error
		name = "final"
	)

	sql.Register(name, sqlhooks.Wrap(&sqlite3.SQLiteDriver{}, hooks))
	db, err := sql.Open(name, ":memory:")
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE test(id int)")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO test VALUES(1)")
	require.NoError(t, err)

	return db
}

func doQuery(db *sql.DB, query string) (*sql.Rows, error) {
	return db.Query(query)
}

func TestFinalizers(t *testing.T) {
	hooks := New()
	db := setupTestDB(t, hooks)

	_, err := doQuery(db, "SELECT * from test")
	require.NoError(t, err)
}
