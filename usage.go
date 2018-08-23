package main

import (
	"fmt"
	"os"
)

func usage(error string, args ...interface{}) {
	if len(error) > 0 {
		if len(args) > 0 {
			fmt.Printf(fmt.Sprintf(error, args...))
		} else {
			fmt.Println(error)
		}
		fmt.Println()
	}

	fmt.Println(`usage: ... | sql target_1 [target_2 ...]

e.g.

  cat query.sql | sql test_db

  sed 's/2015/2016/g' query_for_2015.sql | sql db1 db2 db3

  sql all "SELECT * FROM users WHERE name = 'John'"

For more detailed help, please go to: https://github.com/marianogappa/sql
`)
	os.Exit(1)
}
