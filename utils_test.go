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
