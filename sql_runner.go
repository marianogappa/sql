package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type sqlType int

const (
	mySQL sqlType = iota
	postgreSQL
)

type exists struct{}

type sqlOptions struct {
	cmd   string
	user  string
	host  string
	pass  string
	db    string
	flags string
}

var validSQLTypes = map[sqlType]exists{
	mySQL:      exists{},
	postgreSQL: exists{},
}

var sqlTypeToOptions = map[sqlType]sqlOptions{
	mySQL: {
		"mysql",
		"-u%v",
		"-h%v",
		"-p%v",
		"%v",
		"-Nsre",
	},
	postgreSQL: {
		"psql",
		"-U %v",
		"-h%v",
		"PGPASSWORD=%v",
		"-d %v",
		"-tc",
	},
}

type sqlRunner struct {
	typ         sqlType
	printer     func(string)
	query       string
	quitContext context.Context
	multi       bool
}

func mustNewSQLRunner(quitContext context.Context, typ sqlType, printer func(string), query string, multi bool) *sqlRunner {
	return &sqlRunner{
		typ,
		printer,
		query,
		quitContext,
		multi,
	}
}

func (sr *sqlRunner) runSQL(db database, key string) bool {
	sqlOptions := sqlTypeToOptions[sr.typ]

	userOption := ""
	if db.User != "" {
		userOption = fmt.Sprintf(sqlOptions.user, db.User)
	}

	passOption := ""
	if db.Pass != "" {
		passOption = fmt.Sprintf(sqlOptions.pass, db.Pass)
	}

	hostOption := ""
	if db.DbServer != "" {
		hostOption = fmt.Sprintf(sqlOptions.host, db.DbServer)
	}

	dbOption := ""
	if db.DbName != "" {
		dbOption = fmt.Sprintf(sqlOptions.db, db.DbName)
	}

	prepend := ""
	if sr.multi {
		prepend = key + "\t"
	}

	options := ""
	if sr.typ == postgreSQL {
		options = fmt.Sprintf("%v %v %v %v %v", sqlOptions.cmd, userOption, hostOption, dbOption, sqlOptions.flags)
	} else {
		options = fmt.Sprintf("%v %v %v %v %v %v", sqlOptions.cmd, dbOption, userOption, passOption, hostOption, sqlOptions.flags)
	}

	var cmd *exec.Cmd
	if db.AppServer != "" {
		escapedQuery := fmt.Sprintf(`'%v'`, strings.Replace(sr.query, `'`, `'"'"'`, -1))
		cmd = exec.CommandContext(sr.quitContext, "ssh", db.AppServer, options+escapedQuery)

	} else {
		args := append(trimEmpty(strings.Split(options, " ")), sr.query)
		cmd = exec.CommandContext(sr.quitContext, args[0], args[1:]...)
	}

	if sr.typ == postgreSQL {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, passOption)
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
		line := scanner.Text()
		if line != "" {
			sr.printer(prepend + strings.TrimSpace(line))
		}
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
