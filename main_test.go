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

type tests []struct {
	name      string
	targetDBs []string
	query     string
	expected  []string
}
type testConfig map[string]database

var baseTests = tests{
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
		name:      "reads two fields from all databases",
		targetDBs: []string{"db1", "db2", "db3"},
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

var mysqlTests = tests{{
	name:      `mysql outputs vertical with \G`,
	targetDBs: []string{"db1"},
	query:     `SELECT id, name FROM table1 order by id asc limit 1\G`,
	expected: []string{
		"",
		"*************************** 1. row ***************************",
		"id: 1",
		"name: John",
	},
},
}

func Test_MySQL(t *testing.T) {
	awaitDB(mySQL, t)

	var (
		testConfig = testConfig{
			"db1": database{DbServer: "test-mysql", DbName: "db1", User: "root", Pass: "", SQLType: "mysql"},
			"db2": database{DbServer: "test-mysql", DbName: "db2", User: "root", Pass: "", SQLType: ""},
			"db3": database{DbServer: "test-mysql", DbName: "db3", User: "root", Pass: "", SQLType: ""},
		}
	)
	runTests(baseTests, testConfig, t)
	runTests(mysqlTests, testConfig, t)
}

func Test_PostgreSQL(t *testing.T) {
	awaitDB(postgreSQL, t)

	var (
		testConfig = testConfig{
			"db1": database{DbServer: "test-postgres", DbName: "db1", User: "root", Pass: "", SQLType: "postgres"},
			"db2": database{DbServer: "test-postgres", DbName: "db2", User: "root", Pass: "", SQLType: "postgres"},
			"db3": database{DbServer: "test-postgres", DbName: "db3", User: "root", Pass: "", SQLType: "postgres"},
		}
	)
	runTests(baseTests, testConfig, t)
}

func Test_Mix_Mysql_PostgreSQL(t *testing.T) {
	awaitDB(mySQL, t)
	awaitDB(postgreSQL, t)

	var (
		testConfig = testConfig{
			"db1": database{DbServer: "test-postgres", DbName: "db1", User: "root", Pass: "", SQLType: "postgres"},
			"db2": database{DbServer: "test-postgres", DbName: "db2", User: "root", Pass: "", SQLType: "postgres"},
			"db3": database{DbServer: "test-mysql", DbName: "db3", User: "root", Pass: "", SQLType: ""},
			"db4": database{DbServer: "test-mysql", DbName: "db1", User: "root", Pass: "", SQLType: "mysql"},
		}
		ts = tests{
			{
				name:      "reads from two databases",
				targetDBs: []string{"db1", "db3"},
				query:     "SELECT id FROM table1",
				expected: []string{
					"",
					"db1\t1",
					"db1\t2",
					"db1\t3",
					"db3\t1",
					"db3\t2",
					"db3\t3",
				},
			},
			{
				name:      "reads two fields from all databases",
				targetDBs: []string{"db1", "db2", "db3", "db4"},
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
					"db4\t1\tJohn",
					"db4\t2\tGeorge",
					"db4\t3\tRichard",
				},
			},
		}
	)
	runTests(ts, testConfig, t)
}

func runTests(ts tests, testConfig testConfig, t *testing.T) {
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			var buf = bytes.Buffer{}
			var testSettings = settings{MaxAppServerConnections: 5, Databases: testConfig}
			_main(&testSettings, tc.targetDBs, tc.query, newThreadSafePrintliner(&buf).println)
			var actual = strings.Split(buf.String(), "\n")
			sort.Strings(actual)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected\n%#v\n but got\n%#v\n", tc.expected, actual)
			}
		})
	}
}

func awaitDB(typ sqlType, t *testing.T) {
	var pgTestCmds = []string{"-h", "test-postgres", "-U", "root", "-d", "db1", "-c", "SELECT * FROM table1"}
	var msTestCmds = []string{"-h", "test-mysql", "-u", "root", "-e", "SELECT * FROM db1.table1"}
	var err error
	var c *exec.Cmd
	for i := 1; i <= 30; i++ {
		if typ == mySQL {
			c = exec.Command("mysql", msTestCmds...)
		} else {
			c = exec.Command("psql", pgTestCmds...)
		}
		if err = c.Run(); err == nil {
			log.Printf("%s ready after %v/30 tries", typ, i)
			break
		}
		log.Printf("Retrying (%v/30) in 1 sec because %s is not yet ready", i, typ)
		time.Sleep(1 * time.Second)
	}
	for err != nil {
		t.Errorf("bailing because couldn't connect to %s after 30 tries: %v", typ, err)
		t.FailNow()
	}
}
