package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("unable to get user home dir")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"~", home},
		{"~/projects", filepath.Join(home, "projects")},
		{"/var/log", "/var/log"},
	}

	for _, tt := range tests {
		actual := ExpandHome(tt.input)
		if actual != tt.expected {
			t.Errorf("ExpandHome(%q) = %q; expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestShorthandResolver_ResolveProjectArgs(t *testing.T) {
	resolver := NewShorthandResolver()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	expectedBaseName := filepath.Base(cwd)

	// Test dot resolution
	name, list, path, err := resolver.ResolveProjectArgs(".", ".", "default")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if name != expectedBaseName {
		t.Errorf("expected name %q, got %q", expectedBaseName, name)
	}
	if list != "default" {
		t.Errorf("expected list 'default', got %q", list)
	}
	if path != cwd {
		t.Errorf("expected path %q, got %q", cwd, path)
	}
}
