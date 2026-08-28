package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"hoppr/internal/storage"
)

// DiagnosticResult represents the result of an environment check.
type DiagnosticResult struct {
	Title   string
	Passed  bool
	Message string
}

// DoctorService runs system and configuration diagnostics.
type DoctorService struct {
	storage storage.StorageEngine
}

// NewDoctorService creates a new DoctorService.
func NewDoctorService(storage storage.StorageEngine) *DoctorService {
	return &DoctorService{storage: storage}
}

// RunDiagnostics executes all health checks.
func (d *DoctorService) RunDiagnostics(ctx context.Context) ([]DiagnosticResult, error) {
	var results []DiagnosticResult

	// 1. Check config file accessibility
	cfgPath := d.storage.ConfigPath()
	cfgDir := filepath.Dir(cfgPath)

	if _, err := os.Stat(cfgDir); err != nil {
		results = append(results, DiagnosticResult{
			Title:   "Config Directory",
			Passed:  false,
			Message: fmt.Sprintf("Directory %s does not exist or cannot be accessed: %v", cfgDir, err),
		})
	} else {
		results = append(results, DiagnosticResult{
			Title:   "Config Directory",
			Passed:  true,
			Message: fmt.Sprintf("Accessible at %s", cfgDir),
		})
	}

	cfg, err := d.storage.Read(ctx)
	if err != nil {
		results = append(results, DiagnosticResult{
			Title:   "Config File",
			Passed:  false,
			Message: fmt.Sprintf("Failed to load: %v", err),
		})
		return results, nil
	}

	results = append(results, DiagnosticResult{
		Title:   "Config File",
		Passed:  true,
		Message: fmt.Sprintf("Valid configuration found (%d lists)", len(cfg.Lists)),
	})

	// 2. Validate saved project paths
	missingProjects := 0
	totalProjects := 0
	for listName, projects := range cfg.Lists {
		for projName, path := range projects {
			totalProjects++
			if _, err := os.Stat(path); err != nil {
				missingProjects++
				results = append(results, DiagnosticResult{
					Title:   fmt.Sprintf("Project Path: [%s] %s", listName, projName),
					Passed:  false,
					Message: fmt.Sprintf("Path does not exist on disk: %s", path),
				})
			}
		}
	}

	if missingProjects == 0 {
		results = append(results, DiagnosticResult{
			Title:   "Project Paths",
			Passed:  true,
			Message: fmt.Sprintf("All %d saved project paths are valid on disk", totalProjects),
		})
	}

	// 3. Check Editor in PATH
	if cfg.Editor != "" {
		if path, err := exec.LookPath(cfg.Editor); err != nil {
			results = append(results, DiagnosticResult{
				Title:   "Editor",
				Passed:  false,
				Message: fmt.Sprintf("Configured editor '%s' was not found in system $PATH", cfg.Editor),
			})
		} else {
			results = append(results, DiagnosticResult{
				Title:   "Editor",
				Passed:  true,
				Message: fmt.Sprintf("Configured editor '%s' found at %s", cfg.Editor, path),
			})
		}
	}

	return results, nil
}
