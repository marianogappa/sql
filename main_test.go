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

func Test_MySQL(t *testing.T) {
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
			"db1": database{DbServer: "test-mysql", DbName: "db1", User: "root", Pass: "", _sqlType: mySQL},
			"db2": database{DbServer: "test-mysql", DbName: "db2", User: "root", Pass: "", _sqlType: mySQL},
			"db3": database{DbServer: "test-mysql", DbName: "db3", User: "root", Pass: "", _sqlType: mySQL},
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
					"db1\t1",
					"db1\t2",
					"db1\t3",
					"db2\t1",
					"db2\t2",
					"db2\t3",
				},
			},
			{
				name:      "reads from all databases with the all keyword",
				targetDBs: []string{"all"},
				query:     "SELECT id FROM table1",
				expected: []string{
					"",
					"db1\t1",
					"db1\t2",
					"db1\t3",
					"db2\t1",
					"db2\t2",
					"db2\t3",
					"db3\t1",
					"db3\t2",
					"db3\t3",
				},
			},
			{
				name:      "reads two fields from all databases",
				targetDBs: []string{"all"},
				query:     "SELECT id, name FROM table1",
				expected: []string{
					"",
					"db1\t1\tJohn",
					"db1\t2\tGeorge",
					"db1\t3\tRichard",
					"db2\t1\tRob",
					"db2\t2\tKen",
					"db2\t3\tRobert",
					"db3\t1\tAthos",
					"db3\t2\tPorthos",
					"db3\t3\tAramis",
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

func Test_PostgreSQL(t *testing.T) {
	var err error
	for i := 1; i <= 30; i++ {
		var c = exec.Command("psql", "-h", "test-postgres", "-U", "root", "-d", "db1", "-c", "SELECT * FROM table1")
		if err = c.Run(); err == nil {
			break
		}
		log.Printf("Retrying (%v/30) in 1 sec because PostgreSQL is not yet ready", i)
		time.Sleep(1 * time.Second)
	}
	for err != nil {
		t.Errorf("bailing because couldn't connect to PostgreSQL after 30 tries: %v", err)
		t.FailNow()
	}

	var (
		testConfig = map[string]database{
			"db1": database{DbServer: "test-postgres", DbName: "db1", User: "root", Pass: "", _sqlType: postgreSQL},
			"db2": database{DbServer: "test-postgres", DbName: "db2", User: "root", Pass: "", _sqlType: postgreSQL},
			"db3": database{DbServer: "test-postgres", DbName: "db3", User: "root", Pass: "", _sqlType: postgreSQL},
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
					"db1\t1",
					"db1\t2",
					"db1\t3",
					"db2\t1",
					"db2\t2",
					"db2\t3",
				},
			},
			{
				name:      "reads from all databases with the all keyword",
				targetDBs: []string{"all"},
				query:     "SELECT id FROM table1",
				expected: []string{
					"",
					"db1\t1",
					"db1\t2",
					"db1\t3",
					"db2\t1",
					"db2\t2",
					"db2\t3",
					"db3\t1",
					"db3\t2",
					"db3\t3",
				},
			},
			{
				name:      "reads two fields from all databases",
				targetDBs: []string{"all"},
				query:     "SELECT id, name FROM table1",
				expected: []string{
					"",
					"db1\t1 | John",
					"db1\t2 | George",
					"db1\t3 | Richard",
					"db2\t1 | Rob",
					"db2\t2 | Ken",
					"db2\t3 | Robert",
					"db3\t1 | Athos",
					"db3\t2 | Porthos",
					"db3\t3 | Aramis",
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
