package core

import (
	"context"
	"testing"

	"hoppr/internal/domain"
	"hoppr/internal/storage"
)

func TestProjectService_CaseInsensitiveGetPath(t *testing.T) {
	store := storage.NewMemoryStorage()
	service := NewProjectService(store)
	ctx := context.Background()

	// Add project with CamelCase
	_, _, _, err := service.AddProject(ctx, "MyApp", "default")
	if err != nil {
		t.Fatalf("AddProject error: %v", err)
	}

	// Exact lookup
	path, err := service.GetPath(ctx, "MyApp", "default")
	if err != nil {
		t.Errorf("expected exact lookup to succeed, got %v", err)
	}
	if path == "" {
		t.Errorf("expected non-empty path")
	}

	// Lowercase lookup (case-insensitive)
	pathLower, err := service.GetPath(ctx, "myapp", "default")
	if err != nil {
		t.Errorf("expected case-insensitive lookup to succeed, got %v", err)
	}
	if pathLower != path {
		t.Errorf("expected %q, got %q", path, pathLower)
	}

	// Non-existent lookup
	_, err = service.GetPath(ctx, "unknown", "default")
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectService_ListLifecycle(t *testing.T) {
	store := storage.NewMemoryStorage()
	service := NewProjectService(store)
	ctx := context.Background()

	// Create list
	if err := service.CreateList(ctx, "work"); err != nil {
		t.Fatalf("CreateList error: %v", err)
	}

	// Duplicate list should fail
	if err := service.CreateList(ctx, "work"); err != domain.ErrListExists {
		t.Errorf("expected ErrListExists, got %v", err)
	}

	// Rename list
	if err := service.RenameList(ctx, "work", "office"); err != nil {
		t.Fatalf("RenameList error: %v", err)
	}

	// Set default list
	if err := service.SetDefaultList(ctx, "office"); err != nil {
		t.Fatalf("SetDefaultList error: %v", err)
	}

	// Drop default list should fail
	if err := service.DropList(ctx, "office"); err != domain.ErrCannotDeleteDefaultList {
		t.Errorf("expected ErrCannotDeleteDefaultList, got %v", err)
	}
}
