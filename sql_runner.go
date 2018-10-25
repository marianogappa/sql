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
	mySQL sqlType = iota + 1
	postgreSQL
)

func (t sqlType) String() string {
	return [...]string{"", "MySQL", "PostgreSQL"}[t]
}

type exists struct{}

type sqlOptions struct {
	cmd   string
	user  string
	host  string
	pass  string
	db    string
	flags string
}

var validSQLTypes = map[string]sqlType{
	"":         mySQL,
	"mysql":    mySQL,
	"postgres": postgreSQL,
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
		"-tAc",
	},
}

type sqlRunner struct {
	printer     func(string)
	query       string
	quitContext context.Context
	multi       bool
}

func mustNewSQLRunner(quitContext context.Context, printer func(string), query string, multi bool) *sqlRunner {
	return &sqlRunner{
		printer,
		query,
		quitContext,
		multi,
	}
}

func (sr *sqlRunner) runSQL(db database, key string) bool {
	typ, ok := validSQLTypes[db.SQLType]
	if !ok {
		return maybeErrorResult(key, fmt.Sprintf("Unknown sql type %v for %v", db.SQLType, key))
	}

	sqlOptions := sqlTypeToOptions[typ]

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
	if typ == postgreSQL {
		options = fmt.Sprintf("%v %v %v %v %v", sqlOptions.cmd, userOption, hostOption, dbOption, sqlOptions.flags)
	} else {
		options = fmt.Sprintf("%v %v %v %v %v %v", sqlOptions.cmd, dbOption, userOption, passOption, hostOption, sqlOptions.flags)
	}

	var cmd *exec.Cmd
	if db.AppServer != "" {
		escapedQuery := fmt.Sprintf(`'%v'`, strings.Replace(sr.query, `'`, `'"'"'`, -1))
		if typ == postgreSQL {
			escapedQuery += fmt.Sprintf("-F%s", "\t")
		}

		cmd = exec.CommandContext(sr.quitContext, "ssh", db.AppServer, options+escapedQuery)

	} else {
		args := append(trimEmpty(strings.Split(options, " ")), sr.query)
		if typ == postgreSQL {
			args = append(args, fmt.Sprintf("-F%s", "\t"))
		}
		cmd = exec.CommandContext(sr.quitContext, args[0], args[1:]...)
	}

	if typ == postgreSQL {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, passOption)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return maybeErrorResult(key, fmt.Sprintf("Cannot create pipe for STDOUT of running command on %v; not running. err=%v\n", key, err))
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return maybeErrorResult(key, fmt.Sprintf("Cannot create pipe for STDERR of running command on %v; not running. err=%v\n", key, err))
	}

	if err := cmd.Start(); err != nil {
		return maybeErrorResult(key, fmt.Sprintf("Cannot start command on %v; not running. err=%v\n", key, err))
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

	return maybeErrorResult(key, stderrLines...)
}

func maybeErrorResult(key string, stderrLines ...string) bool {
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
