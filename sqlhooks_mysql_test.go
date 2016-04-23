// +build mysql

package sqlhooks

import _ "github.com/go-sql-driver/mysql"

func init() {
	queries["mysql"] = ops{
		wipe:        "DROP TABLE IF EXISTS t",
		create:      "CREATE TABLE t(f1 varchar(32), f2 varchar(32))",
		insert:      "INSERT INTO t VALUES(?, ?)",
		selectwhere: "SELECT f1, f2 FROM t WHERE f1=? AND f2=?",
		selectall:   "SELECT f1, f2 FROM t",
	}
}
