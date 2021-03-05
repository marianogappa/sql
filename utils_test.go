package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestTrimEmpty(t *testing.T) {
	var ts = []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty case",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Removes multiple empty from beginning, middle and end",
			input:    []string{"", "a", "", "b", ""},
			expected: []string{"a", "b"},
		},
		{
			name:     "Doesn't remove spaces",
			input:    []string{" ", "a", "", "b", ""},
			expected: []string{" ", "a", "b"},
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			var actual = trimEmpty(tc.input)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected %v but got %v", tc.expected, actual)
			}
		})
	}
}

func TestReadQuery(t *testing.T) {
	var ts = []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty case",
			input:    ``,
			expected: ``,
		},
		{
			name:     "One space",
			input:    ` `,
			expected: ``,
		},
		{
			name:     "Trims spaces",
			input:    ` SELECT 1 `,
			expected: `SELECT 1`,
		},
		{
			name: "Replaces newlines with spaces",
			input: ` SELECT
1 `,
			expected: `SELECT 1`,
		},
		{
			name: "Replaces newlines with spaces, trimming them on beginning and end",
			input: ` 
			
			SELECT
1 

`,
			expected: `SELECT 1`,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			var actual = readQuery(strings.NewReader(tc.input))
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected %v but got %v", tc.expected, actual)
			}
		})
	}
}

func TestGetTargetDatabases(t *testing.T) {
	testSettings := settings{
		Databases: map[string]database{
			"db1": database{},
			"db2": database{},
			"db3": database{},
			"db4": database{},
			"db5": database{},
			"db6": database{},
		},
		DatabaseGroups: map[string][]string{
			"group1": []string{"db1", "db3"},
			"group2": []string{"db3", "db4"},
			"group3": []string{"db3", "db4", "db5"},
		},
	}

	var ts = []struct {
		name           string
		databaseArgs   []string
		groupExclusion []string
		groupFilter    []string
		groupSelector  []string
		dbExclusion    []string
		expected       []string
	}{
		{
			name:         "'all' databases",
			databaseArgs: []string{"all"},
			expected:     []string{"db1", "db2", "db3", "db4", "db5", "db6"},
		},
		{
			name:          "group 1 union group 2",
			groupSelector: []string{"group1", "group2"},
			expected:      []string{"db1", "db3", "db4"},
		},
		{
			name:        "group 1 intersect group 2",
			groupFilter: []string{"group1", "group2"},
			expected:    []string{"db3"},
		},
		{
			name:        "group 2 intersect group 2",
			groupFilter: []string{"group2", "group3"},
			expected:    []string{"db3", "db4"},
		},
		{
			name:         "'all' databases excluding 'db3' and 'db5'",
			databaseArgs: []string{"all"},
			dbExclusion:  []string{"db3", "db5"},
			expected:     []string{"db1", "db2", "db4", "db6"},
		},
		{
			name:           "'all' databases excluding group 2",
			databaseArgs:   []string{"all"},
			groupExclusion: []string{"group2"},
			expected:       []string{"db1", "db2", "db5", "db6"},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			actual := getTargetDatabases(&testSettings, tc.databaseArgs, tc.groupExclusion, tc.groupFilter, tc.groupSelector, tc.dbExclusion)

			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected\n%#v\n but got\n%#v\n", tc.expected, actual)
			}
		})
	}
}
