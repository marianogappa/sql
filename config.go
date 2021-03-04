package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

type settings struct {
	MaxAppServerConnections int64
}

type database struct {
	AppServer string
	DbServer  string
	DbName    string
	User      string
	Pass      string
	SQLType   string
}

func getConfigPaths(fileName string) []string {
	var paths []string

	usr, err := user.Current()
	if err != nil {
		usage("Couldn't obtain the current user err=%v", err)
	}

	home := usr.HomeDir

	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		xdgHome = fmt.Sprintf("%v/.config/", home)
	}
	xdgHome += fmt.Sprintf("sql/%v", fileName)

	paths = append(paths, xdgHome)

	xdgConfigDirs := strings.Split(os.Getenv("XDG_CONFIG_DIRS"), ":")
	xdgConfigDirs = append(xdgConfigDirs, "/etc/xdg")
	for _, d := range xdgConfigDirs {
		if d != "" {
			paths = append(paths, fmt.Sprintf("%v/sql/%v", d, fileName))
		}
	}

	paths = append(paths, fmt.Sprintf("%v/%v", home, fileName))
	return paths
}

func readFileContent(paths []string) ([]byte, error) {
	var byts []byte
	var err error
	for _, p := range paths {
		if byts, err = ioutil.ReadFile(p); err != nil {
			continue
		}
		break
	}
	return byts, err
}

func mustReadDatabasesConfigFile() map[string]database {
	databases := map[string]database{}

	fileName := ".databases.json"
	paths := getConfigPaths(fileName)
	byts, err := readFileContent(paths)
	if err != nil {
		usage("Couldn't find .%v in the following paths [%v]. err=%v", fileName, paths, err)

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

func readSettingsFile() *settings {
	s := new(settings)
	fileName := ".settings.json"
	paths := getConfigPaths(fileName)
	byts, err := readFileContent(paths)
	if err == nil {
		err = json.Unmarshal(byts, s)
		if err != nil {
			usage("Found but couldn't JSON unmarshal %v. Looked like this:\n\n%v\n\nerr=%v", fileName, string(byts), err)
		}
	}
	if s.MaxAppServerConnections == 0 {
		s.MaxAppServerConnections = 5
	}
	return s
}
