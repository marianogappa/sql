package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/sync/semaphore"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	result := ""
	for _, v := range *i {
		result = result + " " + v
	}
	return result
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var (
		flagHelp       = flag.Bool("help", false, "shows usage")
		flagListDBs    = flag.Bool("list-dbs", false, "List all available DBs (used for auto-completion)")
		flagListGroups = flag.Bool("list-groups", false, "List all available Groups")
	)
	var flagGroupExclusion arrayFlags
	var flagGroupFilter arrayFlags
	var flagGroupSelector arrayFlags
	var flagDbExclusion arrayFlags
	flag.Var(&flagGroupExclusion, "ge", "DB group exclusion (can be repeated)")
	flag.Var(&flagGroupFilter, "gf", "DB group filter (can be repeated, intersection)")
	flag.Var(&flagGroupSelector, "gs", "DB group selector (can be repeated, union)")
	flag.Var(&flagDbExclusion, "de", "DB exclusion (can be repeated)")

	flag.BoolVar(flagHelp, "h", false, "shows usage")
	flag.Parse()
	if *flagHelp {
		usage("")
	}

	settings := mustReadSettings()

	if *flagListDBs { // for auto-completion
		for dbName := range settings.Databases {
			fmt.Print(dbName, " ")
		}
		fmt.Println()
		os.Exit(0)
	}

	if *flagListGroups {
		for groupName := range settings.DatabaseGroups {
			fmt.Print(groupName, " ")
		}
		fmt.Println()
		os.Exit(0)
	}

	nonFlagArgs := flag.Args()

	var query string
	var databasesArgs []string

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Couldn't os.Stdin.Stat(): %v", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Stdin is a terminal. The last argument is the SQL.
		if len(nonFlagArgs) == 0 {
			usage("No SQL to run. Exiting.")
		}
		query = nonFlagArgs[len(nonFlagArgs)-1]
		databasesArgs = nonFlagArgs[:len(nonFlagArgs)-1]
	} else {
		query = readQuery(os.Stdin)
		databasesArgs = nonFlagArgs
	}

	if len(query) <= 3 {
		usage("No SQL to run. Exiting.")
	}

	targetDatabases := getTargetDatabases(settings, databasesArgs, flagGroupExclusion, flagGroupFilter, flagGroupSelector, flagDbExclusion)
	if len(targetDatabases) == 0 {
		usage("No database to run. Exiting.")
	}

	os.Exit(_main(settings, targetDatabases, query, newThreadSafePrintliner(os.Stdout).println))
}

func _main(settings *settings, targetDatabases []string, query string, println func(string)) int {
	quitContext, cancel := context.WithCancel(context.Background())
	go awaitSignal(cancel)

	var wg sync.WaitGroup
	wg.Add(len(targetDatabases))

	appServerSemaphors := make(map[string]*semaphore.Weighted)
	for _, k := range targetDatabases {
		var appServer = settings.Databases[k].AppServer
		if appServer != "" && appServerSemaphors[appServer] == nil {
			appServerSemaphors[appServer] = semaphore.NewWeighted(settings.MaxAppServerConnections)
		}
	}

	sqlRunner := mustNewSQLRunner(quitContext, println, query, len(targetDatabases) > 1)

	returnCode := 0
	for _, k := range targetDatabases {
		go func(db database, k string) {
			defer wg.Done()
			if db.AppServer != "" {
				var sem = appServerSemaphors[db.AppServer]
				sem.Acquire(quitContext, 1)
				defer sem.Release(1)
			}
			if r := sqlRunner.runSQL(db, k); !r {
				returnCode = 1
			}
		}(settings.Databases[k], k)
	}

	wg.Wait()
	return returnCode
}
