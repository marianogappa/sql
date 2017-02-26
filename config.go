package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

func mustReadDatabasesConfigFile() map[string]database {
	var paths []string
	databases := map[string]database{}

	usr, err := user.Current()
	if err != nil {
		usage("Couldn't obtain the current user err=%v", err)
	}

	home := usr.HomeDir

	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		xdgHome = fmt.Sprintf("%v/.config/", home)
	}
	xdgHome += "sql/.databases.json"

	paths = append(paths, xdgHome)

	xdgConfigDirs := strings.Split(os.Getenv("XDG_CONFIG_DIRS"), ":")
	xdgConfigDirs = append(xdgConfigDirs, "/etc/xdg")
	for _, d := range xdgConfigDirs {
		if d != "" {
			paths = append(paths, fmt.Sprintf("%v/sql/.databases.json", d))
		}
	}

	paths = append(paths, fmt.Sprintf("%v/.databases.json", home))

	var byts []byte
	for _, p := range paths {
		if byts, err = ioutil.ReadFile(p); err != nil {
			continue
		}
		break
	}
	if err != nil {
		usage("Couldn't find .databases.json in the following paths [%v]. err=%v", paths, err)
	}

	err = json.Unmarshal(byts, &databases)
	if err != nil {
		usage("Found but couldn't JSON unmarshal .databases.json. Looked like this:\n\n%v\n\nerr=%v", string(byts), err)
	}

	if len(databases) == 0 {
		usage("Couldn't find any database configurations on .databases.json. Looked like this:\n\n%v\n", string(byts))
	}

	return databases
}
