package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hoppr/internal/domain"
	"hoppr/internal/storage"
)

// ProjectService coordinates domain workflows and storage interactions.
type ProjectService struct {
	storage   storage.StorageEngine
	shorthand *ShorthandResolver
}

// NewProjectService creates a new ProjectService instance.
func NewProjectService(storage storage.StorageEngine) *ProjectService {
	return &ProjectService{
		storage:   storage,
		shorthand: NewShorthandResolver(),
	}
}

// AddProject saves a project to a specific or default list.
func (s *ProjectService) AddProject(ctx context.Context, name, list string) (finalName, finalList, path string, err error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return "", "", "", err
	}

	finalName, finalList, path, err = s.shorthand.ResolveProjectArgs(name, list, cfg.DefaultList)
	if err != nil {
		return "", "", "", err
	}

	if finalName == "" {
		return "", "", "", domain.ErrEmptyName
	}

	err = s.storage.Update(ctx, func(c *domain.Config) error {
		if c.Lists[finalList] == nil {
			c.Lists[finalList] = make(map[string]string)
		}
		c.Lists[finalList][finalName] = path
		return nil
	})

	if err != nil {
		return "", "", "", err
	}

	return finalName, finalList, path, nil
}

// RemoveProject deletes a project from a specified or default list.
func (s *ProjectService) RemoveProject(ctx context.Context, name, list string) (finalName, finalList string, err error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return "", "", err
	}

	finalName, finalList, _, err = s.shorthand.ResolveProjectArgs(name, list, cfg.DefaultList)
	if err != nil {
		return "", "", err
	}

	err = s.storage.Update(ctx, func(c *domain.Config) error {
		projects, exists := c.Lists[finalList]
		if !exists {
			return domain.ErrListNotFound
		}
		if _, found := projects[finalName]; !found {
			return domain.ErrProjectNotFound
		}
		delete(c.Lists[finalList], finalName)
		return nil
	})

	if err != nil {
		return "", "", err
	}

	return finalName, finalList, nil
}

// GetPath looks up the directory path for a project name.
func (s *ProjectService) GetPath(ctx context.Context, name, list string) (string, error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return "", err
	}

	targetList := list
	if targetList == "" || targetList == "." {
		targetList = cfg.DefaultList
	}

	// 1. Try exact match in target list
	if projects, ok := cfg.Lists[targetList]; ok {
		if path, found := projects[name]; found {
			return ExpandHome(path), nil
		}
	}

	// 2. Try exact match across all lists
	if list == "" || list == "." {
		for _, projects := range cfg.Lists {
			if path, found := projects[name]; found {
				return ExpandHome(path), nil
			}
		}
	}

	// 3. Case-insensitive fallback in target list
	nameLower := strings.ToLower(name)
	if projects, ok := cfg.Lists[targetList]; ok {
		for projName, path := range projects {
			if strings.ToLower(projName) == nameLower {
				return ExpandHome(path), nil
			}
		}
	}

	// 4. Case-insensitive fallback across all lists
	if list == "" || list == "." {
		for _, projects := range cfg.Lists {
			for projName, path := range projects {
				if strings.ToLower(projName) == nameLower {
					return ExpandHome(path), nil
				}
			}
		}
	}

	return "", domain.ErrProjectNotFound
}

// GetEditor returns the configured editor executable.
func (s *ProjectService) GetEditor(ctx context.Context) (string, error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return "", err
	}
	return cfg.Editor, nil
}

// GetAllLists returns all lists and the default list name.
func (s *ProjectService) GetAllLists(ctx context.Context) (map[string]map[string]string, string, error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return nil, "", err
	}
	return cfg.Lists, cfg.DefaultList, nil
}

// GetList returns the projects of a specific list.
func (s *ProjectService) GetList(ctx context.Context, list string) (map[string]string, bool, error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return nil, false, err
	}

	targetList := list
	if targetList == "" || targetList == "." {
		targetList = cfg.DefaultList
	}

	projects, exists := cfg.Lists[targetList]
	if !exists {
		return nil, false, domain.ErrListNotFound
	}

	return projects, targetList == cfg.DefaultList, nil
}

// CreateList creates a new named list.
func (s *ProjectService) CreateList(ctx context.Context, list string) error {
	trimmed := strings.TrimSpace(list)
	if trimmed == "" || trimmed == "." {
		return domain.ErrEmptyName
	}

	return s.storage.Update(ctx, func(c *domain.Config) error {
		if _, exists := c.Lists[trimmed]; exists {
			return domain.ErrListExists
		}
		c.Lists[trimmed] = make(map[string]string)
		return nil
	})
}

// DropList deletes an entire list.
func (s *ProjectService) DropList(ctx context.Context, list string) error {
	trimmed := strings.TrimSpace(list)
	if trimmed == "" || trimmed == "." {
		return domain.ErrCannotDeleteDefaultList
	}

	return s.storage.Update(ctx, func(c *domain.Config) error {
		if trimmed == c.DefaultList {
			return domain.ErrCannotDeleteDefaultList
		}
		if _, exists := c.Lists[trimmed]; !exists {
			return domain.ErrListNotFound
		}
		delete(c.Lists, trimmed)
		return nil
	})
}

// RenameList renames an existing list.
func (s *ProjectService) RenameList(ctx context.Context, oldName, newName string) error {
	oldTrimmed := strings.TrimSpace(oldName)
	newTrimmed := strings.TrimSpace(newName)

	if oldTrimmed == "" || newTrimmed == "" {
		return domain.ErrEmptyName
	}

	return s.storage.Update(ctx, func(c *domain.Config) error {
		projects, exists := c.Lists[oldTrimmed]
		if !exists {
			return domain.ErrListNotFound
		}
		if _, exists := c.Lists[newTrimmed]; exists {
			return domain.ErrListExists
		}

		c.Lists[newTrimmed] = projects
		delete(c.Lists, oldTrimmed)

		if c.DefaultList == oldTrimmed {
			c.DefaultList = newTrimmed
		}
		return nil
	})
}

// SetDefaultList changes the active default list.
func (s *ProjectService) SetDefaultList(ctx context.Context, list string) error {
	trimmed := strings.TrimSpace(list)
	if trimmed == "" || trimmed == "." {
		return domain.ErrEmptyName
	}

	return s.storage.Update(ctx, func(c *domain.Config) error {
		if _, exists := c.Lists[trimmed]; !exists {
			// Auto-create list if it doesn't exist
			c.Lists[trimmed] = make(map[string]string)
		}
		c.DefaultList = trimmed
		return nil
	})
}

// Import discovers and adds all immediate subdirectories of a folder as projects.
func (s *ProjectService) Import(ctx context.Context, list, folder string) (importedCount int, finalList, finalFolder string, err error) {
	cfg, err := s.storage.Read(ctx)
	if err != nil {
		return 0, "", "", err
	}

	finalList, finalFolder, err = s.shorthand.ResolveImportArgs(list, folder, cfg.DefaultList)
	if err != nil {
		return 0, "", "", err
	}

	entries, err := os.ReadDir(finalFolder)
	if err != nil {
		return 0, "", "", fmt.Errorf("unable to read folder %s: %w", finalFolder, err)
	}

	subfolders := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subfolders[entry.Name()] = filepath.Join(finalFolder, entry.Name())
		}
	}

	if len(subfolders) == 0 {
		return 0, finalList, finalFolder, nil
	}

	err = s.storage.Update(ctx, func(c *domain.Config) error {
		if c.Lists[finalList] == nil {
			c.Lists[finalList] = make(map[string]string)
		}
		for name, path := range subfolders {
			c.Lists[finalList][name] = path
		}
		return nil
	})

	if err != nil {
		return 0, "", "", err
	}

	return len(subfolders), finalList, finalFolder, nil
}
