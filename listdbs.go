package main

import (
	"fmt"
	"os"
)

func listAllDBs() {
	databases := mustReadDatabasesConfigFile()

	for dbName, _ := range databases {
		fmt.Print(dbName, " ")
	}
	fmt.Println()

	os.Exit(0)
}
