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
	for i := 1; i <= 30; i++ { // Try up to 30 times, because MySQL takes a while to become online
		var c = exec.Command("mysql", "-h", "test-mysql", "-u", "root", "-e", "SELECT * FROM db1.table1")
		if err = c.Run(); err == nil {
			break
		}
		log.Printf("Retrying (%v/30) in 1 sec because MySQL is not yet ready", i)
		time.Sleep(1 * time.Second)
	}
	for err != nil {
		t.Errorf("bailing because couldn't connect to MySQL after 30 tries: %v", err)
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
				name:      "reads from one database",
				targetDBs: []string{"db1"},
				query:     "SELECT id FROM table1",
				expected: []string{
					"",
					"1",
					"2",
					"3",
				},
			},
			{
				name:      "reads from two databases",
				targetDBs: []string{"db1", "db2"},
				query:     "SELECT id FROM table1",
				expected: []string{
					"",
					"db1	1",
					"db1	2",
					"db1	3",
					"db2	1",
					"db2	2",
					"db2	3",
				},
			},
			{
				name:      "reads from all databases with the all keyword",
				targetDBs: []string{"all"},
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
			{
				name:      "reads two fields from all databases",
				targetDBs: []string{"all"},
				query:     "SELECT id, name FROM table1",
				expected: []string{
					"",
					"db1	1	John",
					"db1	2	George",
					"db1	3	Richard",
					"db2	1	Rob",
					"db2	2	Ken",
					"db2	3	Robert",
					"db3	1	Athos",
					"db3	2	Porthos",
					"db3	3	Aramis",
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
