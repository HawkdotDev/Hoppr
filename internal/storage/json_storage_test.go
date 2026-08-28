package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"hoppr/internal/domain"
)

func TestJSONStorage_PersistenceAndAtomicWrite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hoppr-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewJSONStorage(tempDir)
	if err != nil {
		t.Fatalf("NewJSONStorage error: %v", err)
	}

	ctx := context.Background()

	// Initial read should return default config
	cfg, err := store.Read(ctx)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if cfg.DefaultList != "default" {
		t.Errorf("expected default list 'default', got '%s'", cfg.DefaultList)
	}

	// Update config
	err = store.Update(ctx, func(c *domain.Config) error {
		if c.Lists["work"] == nil {
			c.Lists["work"] = make(map[string]string)
		}
		c.Lists["work"]["api"] = "/home/dev/api"
		return nil
	})
	if err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Reload from disk in a fresh storage instance
	freshStore, err := NewJSONStorage(tempDir)
	if err != nil {
		t.Fatalf("NewJSONStorage error: %v", err)
	}

	reloaded, err := freshStore.Read(ctx)
	if err != nil {
		t.Fatalf("fresh read error: %v", err)
	}

	if reloaded.Lists["work"]["api"] != "/home/dev/api" {
		t.Errorf("expected path '/home/dev/api', got '%s'", reloaded.Lists["work"]["api"])
	}
}

func TestJSONStorage_LegacyMigration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hoppr-legacy-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write old schema directly
	legacyJSON := `{
		"projects": {
			"old-proj": "/legacy/path"
		},
		"editor": "nvim"
	}`
	cfgPath := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(legacyJSON), 0644); err != nil {
		t.Fatalf("write legacy json error: %v", err)
	}

	store, err := NewJSONStorage(tempDir)
	if err != nil {
		t.Fatalf("NewJSONStorage error: %v", err)
	}

	ctx := context.Background()
	cfg, err := store.Read(ctx)
	if err != nil {
		t.Fatalf("read migrated error: %v", err)
	}

	if cfg.Lists["default"]["old-proj"] != "/legacy/path" {
		t.Errorf("expected migrated path '/legacy/path', got '%s'", cfg.Lists["default"]["old-proj"])
	}
	if cfg.Editor != "nvim" {
		t.Errorf("expected editor 'nvim', got '%s'", cfg.Editor)
	}
}
