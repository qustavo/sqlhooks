// +build sqlite3

package sqlhooks

import _ "github.com/mattn/go-sqlite3"

func init() {
	queries["sqlite3"] = ops{
		wipe:        "DROP TABLE IF EXISTS t",
		create:      "CREATE TABLE t(f1, f2)",
		insert:      "INSERT INTO t VALUES(?, ?)",
		selectwhere: "SELECT f1, f2 FROM t WHERE f1=? AND f2=?",
		selectall:   "SELECT f1, f2 FROM t",
	}
}
