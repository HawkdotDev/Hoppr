package ui

import (
	"testing"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"list", "lsit", 2},
		{"doctor", "doctro", 2},
		{"doctor", "doctor", 0},
		{"hoppr", "hop", 2},
		{"remove", "rm", 4},
	}

	for _, tt := range tests {
		dist := Levenshtein(tt.a, tt.b)
		if dist != tt.expected {
			t.Errorf("Levenshtein(%q, %q) = %d; expected %d", tt.a, tt.b, dist, tt.expected)
		}
	}
}

func TestFindClosestMatch(t *testing.T) {
	candidates := []string{"add", "remove", "list", "create", "drop", "rename", "setdefault", "import", "doctor"}

	tests := []struct {
		input       string
		maxDistance int
		expected    string
	}{
		{"lsit", 2, "list"},
		{"doctro", 2, "doctor"},
		{"creat", 2, "create"},
		{"completelyunknown", 2, ""},
	}

	for _, tt := range tests {
		match := FindClosestMatch(tt.input, candidates, tt.maxDistance)
		if match != tt.expected {
			t.Errorf("FindClosestMatch(%q) = %q; expected %q", tt.input, match, tt.expected)
		}
	}
}
