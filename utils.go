package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
)

type threadSafePrintliner struct {
	l sync.Mutex
	w io.Writer
}

func newThreadSafePrintliner(w io.Writer) *threadSafePrintliner {
	return &threadSafePrintliner{w: w}
}

func (p *threadSafePrintliner) println(s string) {
	p.l.Lock()
	fmt.Fprintln(p.w, s)
	p.l.Unlock()
}

func readQuery(r io.Reader) string {
	s, _ := ioutil.ReadAll(r) // N.B. not interested in this error; might as well return an empty string
	return strings.TrimSpace(strings.Replace(string(s), "\n", " ", -1))
}

func trimEmpty(s []string) []string {
	var r = make([]string, 0)
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func awaitSignal(cancel context.CancelFunc) {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	cancel()
}

func _mustGetGroupDatabases(settings *settings, groupName string) []string {
	groupDatabases, ok := settings.DatabaseGroups[groupName]
	if !ok {
		usage("Selected group is unknown: [%v]", groupName)
	}
	if len(groupDatabases) == 0 {
		usage("Selected group has no databases: [%v]", groupName)
	}

	for _, dbName := range groupDatabases {
		if _, ok := settings.Databases[dbName]; !ok {
			usage("Database [%v] in group [%v] unknown", dbName, groupName)
		}
	}
	return groupDatabases
}

func _contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getTargetDatabases(settings *settings, databasesArgs []string, flagGroupExclusion []string, flagGroupFilter []string, flagGroupSelector []string, flagDbExclusion []string) []string {
	targetDatabases := make(map[string]bool)

	// add explicit databases from Args, or 'all' databases
	for _, arg := range databasesArgs {
		if _, ok := settings.Databases[arg]; arg != "all" && !ok {
			usage("Target database unknown: [%v]", arg)
		}
		if arg == "all" {
			for dbName := range settings.Databases {
				targetDatabases[dbName] = true
			}
		} else {
			targetDatabases[arg] = true
		}
	}

	// add Selected Groups
	for _, groupName := range flagGroupSelector {
		for _, databaseName := range _mustGetGroupDatabases(settings, groupName) {
			targetDatabases[databaseName] = true
		}
	}

	// if no explicit databases, 'all' databases or Selected Groups has been set, use Filtered Group
	if len(targetDatabases) == 0 {
		if len(flagGroupFilter) == 0 {
			usage("Must either specify 'all', group selectors, group filters or a list of databases.")
		}
		for _, databaseName := range _mustGetGroupDatabases(settings, flagGroupFilter[0]) {
			targetDatabases[databaseName] = true
		}
	}

	// remove databases that are not in filtered group
	for _, groupName := range flagGroupFilter {
		groupFilterDatabases := _mustGetGroupDatabases(settings, groupName)
		for databaseName, _ := range targetDatabases {
			if !_contains(groupFilterDatabases, databaseName) {
				delete(targetDatabases, databaseName)
			}
		}
	}

	// remove databases of groups that are explicitly removed
	for _, groupName := range flagGroupExclusion {
		groupExclusionDatabases := _mustGetGroupDatabases(settings, groupName)
		for _, databaseName := range groupExclusionDatabases {
			delete(targetDatabases, databaseName)
		}
	}

	// remove databases that are explicitly removed
	for _, databaseName := range flagDbExclusion {
		delete(targetDatabases, databaseName)
	}

	result := []string{}
	for databaseName := range targetDatabases {
		result = append(result, databaseName)
	}
	sort.Strings(result)
	return result
}
