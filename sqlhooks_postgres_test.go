// +build postgres

package sqlhooks

import _ "github.com/lib/pq"

func init() {
	queries["postgres"] = ops{
		wipe:        "DROP TABLE IF EXISTS t",
		create:      "CREATE TABLE t(f1 varchar(32), f2 varchar(32))",
		insert:      "INSERT INTO t VALUES($1, $2)",
		selectwhere: "SELECT f1, f2 FROM t WHERE f1=$1 AND f2=$2",
		selectall:   "SELECT f1, f2 FROM t",
	}
}
