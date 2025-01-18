package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Vipelierre Best starter",
			expected: []string{"vipelierre", "best", "starter"},
		},
		{
			input:    "HOLLOW KNIGHT is a masterpiece",
			expected: []string{"hollow", "knight", "is", "a", "masterpiece"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("got length %v, want length %v for input %q", len(actual), len(c.expected), c.input)
		}
		for i := range actual {
			word := actual[i]
			expectedword := c.expected[i]
			if word != expectedword {
				t.Errorf("got word: %v, want word:%v, for input: %q", word, expectedword, c.input)
			}

		}
	}
}
