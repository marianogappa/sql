package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
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

func main() {
	usr, err := user.Current()
	if err != nil {
		usage("Couldn't obtain the current user err=%v", err)
	}

	databasesFile := usr.HomeDir + "/.databases.json"
	databases := map[string]database{}

	byts, err := ioutil.ReadFile(databasesFile)
	if err != nil {
		usage("Couldn't read [%v] file. err=%v", databasesFile, err)
	}

	err = json.Unmarshal(byts, &databases)
	if err != nil {
		usage("Couldn't unmarshal [%v] file. err=%v", databasesFile, err)
	}

	if len(databases) == 0 {
		usage("Couldn't find any database configurations in [%v] file.", databasesFile)
	}

	sql := readInput(os.Stdin)
	if len(sql) <= 3 {
		usage("No SQL to run. Exiting.")
	}

	if len(os.Args[1:]) == 0 {
		usage("Target database unspecified; where should I run the query?")
	}

	targetDatabases := []string{}
	for _, k := range os.Args[1:] {
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

	var wg sync.WaitGroup

	for _, k := range targetDatabases {
		wg.Add(1)
		go func(db database, k string) {
			defer wg.Done()
			runSQL(db, sql, k, len(targetDatabases) > 1)
		}(databases[k], k)
	}

	wg.Wait()
}

func runSQL(db database, sql string, key string, prependKey bool) {
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
		cmd = exec.Command("ssh", db.AppServer, mysql+options+query)
	} else {
		args := append(trimEmpty(strings.Split(options, " ")), sql)
		cmd = exec.Command("mysql", args...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Cannot create pipe for STDOUT of running command on %v; not running.\n", key)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Cannot create pipe for STDERR of running command on %v; not running.\n", key)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Cannot start command on %v; not running.\n", key)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(prepend + scanner.Text())
	}

	stderrLines := []string{}
	scanner = bufio.NewScanner(stderr)
	for scanner.Scan() {
		stderrLines = append(stderrLines, scanner.Text())
	}

	if len(stderrLines) > 0 {
		log.Println(key + " had errors:")
		for _, v := range stderrLines {
			log.Println(key + " [ERROR] " + v)
		}
	}

	cmd.Wait()
}

func readInput(r io.Reader) string {
	ls := []string{}
	var err error
	rd := bufio.NewReader(r)

	for {
		var s string
		s, err = rd.ReadString('\n')
		if err == io.EOF {
			return strings.Join(ls, " ")
		}
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		ls = append(ls, strings.TrimSpace(s))
	}
}

func trimEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
