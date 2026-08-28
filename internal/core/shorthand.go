package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ShorthandResolver handles contextual '.', '~', and path normalization.
type ShorthandResolver struct{}

// NewShorthandResolver creates a new ShorthandResolver instance.
func NewShorthandResolver() *ShorthandResolver {
	return &ShorthandResolver{}
}

// ExpandHome resolves '~' or '~/...' to the user's home directory.
func ExpandHome(path string) string {
	if path == "" {
		return ""
	}
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// ResolveProjectArgs resolves '.', empty values, tilde '~', and relative paths contextually.
func (r *ShorthandResolver) ResolveProjectArgs(name, list string, defaultList string) (resolvedName, resolvedList, targetPath string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("unable to determine current working directory: %w", err)
	}

	// Resolve list
	if list == "" || list == "." {
		resolvedList = defaultList
	} else {
		resolvedList = strings.TrimSpace(list)
	}

	// Resolve name
	if name == "" || name == "." {
		resolvedName = filepath.Base(cwd)
	} else {
		resolvedName = strings.TrimSpace(name)
	}

	targetPath = filepath.Clean(cwd)
	return resolvedName, resolvedList, targetPath, nil
}

// ResolveImportArgs resolves list and folder directory for bulk-import with '~' support.
func (r *ShorthandResolver) ResolveImportArgs(list, folder string, defaultList string) (resolvedList, resolvedFolder string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("unable to determine current working directory: %w", err)
	}

	if list == "" || list == "." {
		resolvedList = defaultList
	} else {
		resolvedList = strings.TrimSpace(list)
	}

	if folder == "" || folder == "." {
		resolvedFolder = cwd
	} else {
		expanded := ExpandHome(strings.TrimSpace(folder))
		if filepath.IsAbs(expanded) {
			resolvedFolder = expanded
		} else {
			resolvedFolder = filepath.Join(cwd, expanded)
		}
	}

	resolvedFolder = filepath.Clean(resolvedFolder)
	return resolvedList, resolvedFolder, nil
}
