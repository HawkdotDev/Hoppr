package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ShorthandResolver handles contextual '.' resolution.
type ShorthandResolver struct{}

// NewShorthandResolver creates a new ShorthandResolver instance.
func NewShorthandResolver() *ShorthandResolver {
	return &ShorthandResolver{}
}

// ResolveProjectArgs resolves '.', empty values, and relative paths contextually.
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

// ResolveImportArgs resolves list and folder directory for bulk-import.
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
		resolvedFolder = folder
		if !filepath.IsAbs(resolvedFolder) {
			resolvedFolder = filepath.Join(cwd, resolvedFolder)
		}
	}

	resolvedFolder = filepath.Clean(resolvedFolder)
	return resolvedList, resolvedFolder, nil
}
