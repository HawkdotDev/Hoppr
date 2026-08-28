package domain

import (
	"time"
)

// Project represents a saved codebase shortcut.
type Project struct {
	Name         string    `json:"name,omitempty"`
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	LastAccessed time.Time `json:"last_accessed,omitempty"`
	AccessCount  int       `json:"access_count,omitempty"`
}

// Config represents the complete Hoppr configuration schema.
type Config struct {
	SchemaVersion int                          `json:"schema_version"`
	Lists         map[string]map[string]string `json:"lists"`
	DefaultList   string                       `json:"default_list"`
	Editor        string                       `json:"editor"`
}

// NewDefaultConfig returns a freshly initialized Config.
func NewDefaultConfig() *Config {
	return &Config{
		SchemaVersion: 1,
		Lists: map[string]map[string]string{
			"default": make(map[string]string),
		},
		DefaultList: "default",
		Editor:      "code",
	}
}
