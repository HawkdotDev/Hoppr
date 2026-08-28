package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"hoppr/internal/domain"
)

// LegacyConfig represents the older single-map schema for backward compatibility.
type LegacyConfig struct {
	Projects map[string]string `json:"projects"`
	Editor   string            `json:"editor"`
}

// JSONStorage implements StorageEngine using atomic JSON files and file locks.
type JSONStorage struct {
	configDir  string
	configPath string
	lockPath   string
}

// NewJSONStorage creates a new JSONStorage instance.
func NewJSONStorage(configDir string) (*JSONStorage, error) {
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to resolve user home directory: %w", err)
		}
		configDir = filepath.Join(home, ".hoppr")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	return &JSONStorage{
		configDir:  configDir,
		configPath: filepath.Join(configDir, "config.json"),
		lockPath:   filepath.Join(configDir, "config.lock"),
	}, nil
}

func (s *JSONStorage) ConfigPath() string {
	return s.configPath
}

// Read loads the configuration under a shared read-lock.
func (s *JSONStorage) Read(ctx context.Context) (*domain.Config, error) {
	fl, err := newFileLock(s.lockPath)
	if err == nil {
		if err := fl.RLock(); err == nil {
			defer fl.Unlock()
		} else {
			_ = fl.Unlock()
		}
	}

	return s.loadConfigUnlocked()
}

// Update loads, modifies, and atomically writes the configuration under an exclusive lock.
func (s *JSONStorage) Update(ctx context.Context, mutate func(cfg *domain.Config) error) error {
	fl, err := newFileLock(s.lockPath)
	if err != nil {
		return fmt.Errorf("failed to acquire storage lock: %w", err)
	}
	defer fl.Unlock()

	if err := fl.Lock(); err != nil {
		return fmt.Errorf("failed to obtain exclusive lock: %w", err)
	}

	cfg, err := s.loadConfigUnlocked()
	if err != nil {
		return err
	}

	if err := mutate(cfg); err != nil {
		return err
	}

	return s.saveConfigUnlocked(cfg)
}

func (s *JSONStorage) loadConfigUnlocked() (*domain.Config, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Attempt to decode current schema
	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// If current schema fails, check if backup exists or report corrupted JSON
		return nil, fmt.Errorf("config file is corrupted (invalid JSON syntax): %w", err)
	}

	// Handle automatic migration from legacy schema (where lists was empty and projects existed)
	if len(cfg.Lists) == 0 {
		var legacy LegacyConfig
		if err := json.Unmarshal(data, &legacy); err == nil && len(legacy.Projects) > 0 {
			cfg.Lists = map[string]map[string]string{
				"default": legacy.Projects,
			}
			if legacy.Editor != "" {
				cfg.Editor = legacy.Editor
			}
		}
	}

	// Enforce valid default state
	if cfg.Lists == nil {
		cfg.Lists = make(map[string]map[string]string)
	}
	if cfg.DefaultList == "" {
		cfg.DefaultList = "default"
	}
	if _, ok := cfg.Lists[cfg.DefaultList]; !ok {
		cfg.Lists[cfg.DefaultList] = make(map[string]string)
	}
	if cfg.Editor == "" {
		cfg.Editor = "code"
	}
	cfg.SchemaVersion = 1

	return &cfg, nil
}

func (s *JSONStorage) saveConfigUnlocked(cfg *domain.Config) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Create temporary file in the same directory for atomic rename
	tmpFile, err := os.CreateTemp(s.configDir, "hoppr-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpName := tmpFile.Name()
	// SEC-5: Restrict temp file permissions to owner-only
	_ = os.Chmod(tmpName, 0600)
	defer os.Remove(tmpName) // Cleanup if rename fails

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write config data to temp file: %w", err)
	}

	// Flush OS write buffer to physical storage
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file to disk: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// SEC-3: Backup the EXISTING config before replacing (not the new data)
	if existingData, readErr := os.ReadFile(s.configPath); readErr == nil {
		backupPath := s.configPath + ".bak"
		_ = os.WriteFile(backupPath, existingData, 0600)
	}

	// Atomically replace config file
	if err := os.Rename(tmpName, s.configPath); err != nil {
		return fmt.Errorf("failed to commit config file atomically: %w", err)
	}

	return nil
}
