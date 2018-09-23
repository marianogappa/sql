package main

import (
	"bytes"
	"log"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestSQL(t *testing.T) {
	var err error
	for i := 1; i <= 5; i++ { // Try up to 10 times, because MySQL takes a while to become online
		var c = exec.Command("mysql", "-h", "test-mysql", "-u", "root", "-e", "SELECT * FROM db1.table1")
		if err = c.Run(); err == nil {
			break
		}
		log.Printf("Retrying (%v/10) in 1 sec because MySQL is not yet ready", i)
		time.Sleep(1 * time.Second)
	}
	for err != nil {
		t.Errorf("bailing because couldn't connect to MySQL after 10 tries: %v", err)
		t.FailNow()
	}

	var (
		testConfig = map[string]database{
			"db1": database{DbServer: "test-mysql", DbName: "db1", User: "root", Pass: ""},
			"db2": database{DbServer: "test-mysql", DbName: "db2", User: "root", Pass: ""},
			"db3": database{DbServer: "test-mysql", DbName: "db3", User: "root", Pass: ""},
		}
		ts = []struct {
			name      string
			targetDBs []string
			query     string
			expected  []string
		}{
			{
				name:      "happy case",
				targetDBs: []string{"db1", "db2", "db3"},
				query:     "SELECT id FROM table1",
				expected: []string{
					"",
					"db1	1",
					"db1	2",
					"db1	3",
					"db2	1",
					"db2	2",
					"db2	3",
					"db3	1",
					"db3	2",
					"db3	3",
				},
			},
		}
	)
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			var buf = bytes.Buffer{}
			_main(testConfig, tc.targetDBs, tc.query, newThreadSafePrintliner(&buf).println)
			var actual = strings.Split(buf.String(), "\n")
			sort.Strings(actual)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected %v but got %v", tc.expected, actual)
			}
		})
	}
}
