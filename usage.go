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

	fmt.Println(`sql usage:
  anything | sql [-p] target_1
  anything | sql target_1 [target_2...]
  anything | sql all

target_1, target_2 etc must be defined in $HOME/.databases.json

By default, no column names are output.
Querying one target outputs one line per result row, as tab separated values.
Querying multiple targets outputs one line per target per result row, as TSV, prepended with target_name.
Querying "all" targets every configured database.

The 'pretty' flag, -p, modifies the output to include +--+ separators and column names for human readability.
Using -p with multiple targets is not advised.

Examples:

  cat query.sql | sql test_db
  echo "SELECT * FROM users WHERE name = 'John'" | sql all
  sed 's/2015/2016/g' query_for_2015.sql | sql db1 db2 db3
  echo "SELECT * FROM users LIMIT 1" | sql -p db1
  echo "SELECT * FROM users LIMIT 1\G" | sql -p db1

For more detailed help, please go to: https://github.com/marianogappa/sql
`)
	os.Exit(1)
}
