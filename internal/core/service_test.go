package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"hoppr/internal/domain"
	"hoppr/internal/storage"
)

func setupTestService() *ProjectService {
	memStore := storage.NewMemoryStorage(domain.NewDefaultConfig())
	return NewProjectService(memStore)
}

func TestProjectService_AddAndGet(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	cwd, _ := os.Getwd()

	name, list, path, err := svc.AddProject(ctx, "test-proj", "work")
	if err != nil {
		t.Fatalf("unexpected error adding project: %v", err)
	}

	if name != "test-proj" {
		t.Errorf("expected name 'test-proj', got '%s'", name)
	}
	if list != "work" {
		t.Errorf("expected list 'work', got '%s'", list)
	}
	if path != cwd {
		t.Errorf("expected path '%s', got '%s'", cwd, path)
	}

	// Retrieve path
	foundPath, err := svc.GetPath(ctx, "test-proj", "work")
	if err != nil {
		t.Fatalf("unexpected error getting path: %v", err)
	}
	if foundPath != cwd {
		t.Errorf("expected found path '%s', got '%s'", cwd, foundPath)
	}

	// Retrieve path across lists via fallback
	foundFallback, err := svc.GetPath(ctx, "test-proj", "")
	if err != nil {
		t.Fatalf("unexpected error getting path via fallback: %v", err)
	}
	if foundFallback != cwd {
		t.Errorf("expected fallback path '%s', got '%s'", cwd, foundFallback)
	}
}

func TestProjectService_ShorthandDot(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	cwd, _ := os.Getwd()
	expectedBase := filepath.Base(cwd)

	// Add with "." name and "." list
	name, list, path, err := svc.AddProject(ctx, ".", ".")
	if err != nil {
		t.Fatalf("unexpected error adding with shorthand: %v", err)
	}

	if name != expectedBase {
		t.Errorf("expected shorthand name '%s', got '%s'", expectedBase, name)
	}
	if list != "default" {
		t.Errorf("expected default list 'default', got '%s'", list)
	}
	if path != cwd {
		t.Errorf("expected path '%s', got '%s'", cwd, path)
	}
}

func TestProjectService_RemoveProject(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	_, _, _, err := svc.AddProject(ctx, "proj1", "default")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}

	// Remove project
	_, _, err = svc.RemoveProject(ctx, "proj1", "default")
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}

	// Verify not found
	_, err = svc.GetPath(ctx, "proj1", "default")
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectService_ListManagement(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	// 1. Create list
	err := svc.CreateList(ctx, "oss")
	if err != nil {
		t.Fatalf("create list error: %v", err)
	}

	// Duplicate create error
	err = svc.CreateList(ctx, "oss")
	if err != domain.ErrListExists {
		t.Errorf("expected ErrListExists, got %v", err)
	}

	// 2. Rename list
	err = svc.RenameList(ctx, "oss", "open-source")
	if err != nil {
		t.Fatalf("rename list error: %v", err)
	}

	// 3. Set default list
	err = svc.SetDefaultList(ctx, "open-source")
	if err != nil {
		t.Fatalf("setdefault error: %v", err)
	}

	// Cannot drop active default list
	err = svc.DropList(ctx, "open-source")
	if err != domain.ErrCannotDeleteDefaultList {
		t.Errorf("expected ErrCannotDeleteDefaultList, got %v", err)
	}

	// Reset default and drop
	_ = svc.SetDefaultList(ctx, "default")
	err = svc.DropList(ctx, "open-source")
	if err != nil {
		t.Fatalf("drop list error: %v", err)
	}
}
