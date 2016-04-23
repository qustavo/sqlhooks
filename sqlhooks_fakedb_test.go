package sqlhooks

func init() {
	queries["test"] = ops{
		wipe:        "WIPE",
		create:      "CREATE|t|f1=string,f2=string",
		insert:      "INSERT|t|f1=?,f2=?",
		selectwhere: "SELECT|t|f1,f2|f1=?,f2=?",
		selectall:   "SELECT|t|f1,f2|",
	}
}
