package storage

import (
	"context"
	"sync"

	"hoppr/internal/domain"
)

// MemoryStorage is an in-memory implementation of StorageEngine for unit testing.
type MemoryStorage struct {
	mu     sync.RWMutex
	config *domain.Config
}

// NewMemoryStorage creates a new in-memory storage engine.
func NewMemoryStorage(initial *domain.Config) *MemoryStorage {
	if initial == nil {
		initial = domain.NewDefaultConfig()
	}
	return &MemoryStorage{
		config: initial,
	}
}

func (m *MemoryStorage) ConfigPath() string {
	return ":memory:"
}

func (m *MemoryStorage) Read(ctx context.Context) (*domain.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Deep copy to prevent race conditions during tests
	copied := deepCopyConfig(m.config)
	return copied, nil
}

func (m *MemoryStorage) Update(ctx context.Context, mutate func(cfg *domain.Config) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	copied := deepCopyConfig(m.config)
	if err := mutate(copied); err != nil {
		return err
	}

	m.config = copied
	return nil
}

func deepCopyConfig(src *domain.Config) *domain.Config {
	dst := &domain.Config{
		SchemaVersion: src.SchemaVersion,
		DefaultList:   src.DefaultList,
		Editor:        src.Editor,
		Lists:         make(map[string]map[string]string),
	}

	for listName, projects := range src.Lists {
		dst.Lists[listName] = make(map[string]string)
		for k, v := range projects {
			dst.Lists[listName][k] = v
		}
	}

	return dst
}
