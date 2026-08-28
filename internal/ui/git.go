package ui

import (
	"os"
	"path/filepath"
	"strings"
)

// GetGitBranch detects if a directory is a Git repository and returns its active branch name.
// Reads .git/HEAD directly for sub-millisecond execution with zero external process overhead.
func GetGitBranch(dirPath string) string {
	gitPath := filepath.Join(dirPath, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return ""
	}

	headPath := ""
	if info.IsDir() {
		headPath = filepath.Join(gitPath, "HEAD")
	} else {
		// Worktree file format: "gitdir: /path/to/.git/worktrees/name"
		content, err := os.ReadFile(gitPath)
		if err != nil {
			return ""
		}
		line := strings.TrimSpace(string(content))
		if strings.HasPrefix(line, "gitdir:") {
			realGitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
			if !filepath.IsAbs(realGitDir) {
				realGitDir = filepath.Join(dirPath, realGitDir)
			}
			headPath = filepath.Join(realGitDir, "HEAD")
		}
	}

	if headPath == "" {
		return ""
	}

	headContent, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}

	headStr := strings.TrimSpace(string(headContent))
	if strings.HasPrefix(headStr, "ref: refs/heads/") {
		return strings.TrimPrefix(headStr, "ref: refs/heads/")
	}

	// Detached HEAD state: show short SHA
	if len(headStr) >= 7 {
		return headStr[:7]
	}

	return ""
}
