package storage

import (
	"context"

	"hoppr/internal/domain"
)

// StorageEngine defines the contract for reading, modifying, and persisting configuration.
type StorageEngine interface {
	// Read loads the configuration under a shared read-lock.
	Read(ctx context.Context) (*domain.Config, error)

	// Update loads, mutates, and atomically saves the configuration under an exclusive lock.
	Update(ctx context.Context, mutate func(cfg *domain.Config) error) error

	// ConfigPath returns the absolute path to the configuration file.
	ConfigPath() string
}
