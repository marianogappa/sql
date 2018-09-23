package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type database struct {
	AppServer string
	DbServer  string
	DbName    string
	User      string
	Pass      string
}

var help = flag.Bool("help", false, "shows usage")
var listDBs = flag.Bool("list-dbs", false, "List all available DBs (used for auto-completion)")

var printLock sync.Mutex

func init() {
	flag.BoolVar(help, "h", false, "shows usage")
}

func main() {
	flag.Parse()
	if *help {
		usage("")
	}
	if *listDBs { // for auto-completion
		for dbName := range mustReadDatabasesConfigFile() {
			fmt.Print(dbName, " ")
		}
		fmt.Println()
		os.Exit(0)
	}

	databases := mustReadDatabasesConfigFile()

	if len(os.Args[1:]) == 0 {
		usage("Target database unspecified; where should I run the query?")
	}

	var query string
	var databasesArgs []string

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Couldn't os.Stdin.Stat(): %v", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Stdin is a terminal. The last argument is the SQL.
		if len(os.Args) < 3 {
			usage("No SQL to run. Exiting.")
		}
		query = os.Args[len(os.Args)-1]
		databasesArgs = os.Args[1 : len(os.Args)-1]
	} else {
		query = readQuery(os.Stdin)
		databasesArgs = os.Args[1:]
	}

	if len(query) <= 3 {
		usage("No SQL to run. Exiting.")
	}

	os.Exit(_main(databases, databasesArgs, query, newThreadSafePrintliner(os.Stdout).println))
}

func _main(databases map[string]database, databasesArgs []string, query string, println func(string)) int {
	targetDatabases := []string{}
	for _, k := range databasesArgs {
		if _, ok := databases[k]; k != "all" && !ok {
			usage("Target database unknown: [%v]", k)
		}
		if k == "all" {
			targetDatabases = nil
			for k := range databases {
				targetDatabases = append(targetDatabases, k)
			}
			break
		}
		targetDatabases = append(targetDatabases, k)
	}

	quitContext, cancel := context.WithCancel(context.Background())
	go awaitSignal(cancel)

	var wg sync.WaitGroup
	wg.Add(len(targetDatabases))

	returnCode := 0
	for _, k := range targetDatabases {
		go func(db database, k string) {
			defer wg.Done()
			if r := runSQL(quitContext, db, query, k, len(targetDatabases) > 1, println); !r {
				returnCode = 1
			}
		}(databases[k], k)
	}

	wg.Wait()
	return returnCode
}

func runSQL(quitContext context.Context, db database, sql string, key string, prependKey bool, println func(string)) bool {
	userOption := ""
	if db.User != "" {
		userOption = fmt.Sprintf("-u %v ", db.User)
	}

	passOption := ""
	if db.Pass != "" {
		passOption = fmt.Sprintf("-p%v ", db.Pass)
	}

	hostOption := ""
	if db.DbServer != "" {
		hostOption = fmt.Sprintf("-h %v ", db.DbServer)
	}

	prepend := ""
	if prependKey {
		prepend = key + "\t"
	}

	mysql := "mysql"
	options := fmt.Sprintf(" -Nsr %v%v%v%v -e ", userOption, passOption, hostOption, db.DbName)

	var cmd *exec.Cmd
	if db.AppServer != "" {
		query := fmt.Sprintf(`'%v'`, strings.Replace(sql, `'`, `'"'"'`, -1))
		cmd = exec.CommandContext(quitContext, "ssh", db.AppServer, mysql+options+query)
	} else {
		args := append(trimEmpty(strings.Split(options, " ")), sql)
		cmd = exec.CommandContext(quitContext, "mysql", args...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Cannot create pipe for STDOUT of running command on %v; not running. err=%v\n", key, err)
		return false
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Cannot create pipe for STDERR of running command on %v; not running. err=%v\n", key, err)
		return false
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Cannot start command on %v; not running. err=%v\n", key, err)
		return false
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		println(prepend + scanner.Text())
	}

	stderrLines := []string{}
	scanner = bufio.NewScanner(stderr)
	for scanner.Scan() {
		stderrLines = append(stderrLines, scanner.Text())
	}

	cmd.Wait()

	result := true
	if len(stderrLines) > 0 {
		result = false
		log.Println(key + " had errors:")
		for _, v := range stderrLines {
			log.Println(key + " [ERROR] " + v)
		}
	}

	return result
}
