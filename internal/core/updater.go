package core

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"hoppr/internal/version"
)

const (
	RepoOwner       = "HawkdotDev"
	RepoName        = "Hoppr"
	GitHubAPI       = "https://api.github.com/repos/HawkdotDev/Hoppr/releases/latest"
	GitHubLatestWeb = "https://github.com/HawkdotDev/Hoppr/releases/latest"
	UserAgent       = "Hoppr-SelfUpdater/1.1"
	HTTPTimeout     = 25 * time.Second
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

type UpdateService struct {
	client *http.Client
}

func NewUpdateService() *UpdateService {
	return &UpdateService{
		client: &http.Client{
			Timeout: HTTPTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 10 redirects for web URL resolution
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// FetchLatestRelease resolves the latest release tag via GitHub API with automatic HTML redirect fallback.
func (u *UpdateService) FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	// 1. Primary: Query GitHub REST API
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubAPI, nil)
	if err == nil {
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := u.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var rel GitHubRelease
			if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil && rel.TagName != "" {
				return &rel, nil
			}
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	// 2. Fallback: Zero-quota HTTP redirect resolution (bypasses GitHub API rate limits)
	headReq, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubLatestWeb, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build release lookup request: %w", err)
	}
	headReq.Header.Set("User-Agent", UserAgent)

	headResp, err := u.client.Do(headReq)
	if err != nil {
		return nil, fmt.Errorf("unable to reach GitHub releases: %w", err)
	}
	defer headResp.Body.Close()

	// Extract tag name from final redirected URL (e.g. .../releases/tag/v1.1.0)
	finalURL := headResp.Request.URL.String()
	if idx := strings.LastIndex(finalURL, "/tag/"); idx != -1 {
		tag := finalURL[idx+5:]
		return &GitHubRelease{
			TagName: tag,
			Name:    "Release " + tag,
		}, nil
	}

	return nil, fmt.Errorf("unable to determine latest release tag from GitHub")
}

// CheckUpdate compares the running version with the latest release tag.
func (u *UpdateService) CheckUpdate(ctx context.Context) (latestTag string, hasUpdate bool, err error) {
	rel, err := u.FetchLatestRelease(ctx)
	if err != nil {
		return "", false, err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(version.Version, "v")

	if latest == "" {
		return "", false, fmt.Errorf("invalid release tag received")
	}

	hasUpdate = latest != current && current != "dev" && current != "unknown"
	return rel.TagName, hasUpdate, nil
}

// DownloadAndApplyUpdate concurrently downloads package and checksums, validates SHA256 in-flight, and replaces the binary.
func (u *UpdateService) DownloadAndApplyUpdate(ctx context.Context, targetTag string, progress func(step string)) error {
	arch := runtime.GOARCH
	osName := runtime.GOOS

	archiveExt := "tar.gz"
	if osName == "windows" {
		archiveExt = "zip"
	}

	assetName := fmt.Sprintf("hoppr-%s-%s-%s.%s", targetTag, osName, arch, archiveExt)
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", RepoOwner, RepoName, targetTag, assetName)
	checksumURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/checksums.txt", RepoOwner, RepoName, targetTag)

	progress(fmt.Sprintf("Downloading %s (parallel fetching archive & checksums)...", assetName))

	// Parallel Download Channel
	var (
		wg           sync.WaitGroup
		archiveData  []byte
		archiveErr   error
		expectedHash string
	)

	wg.Add(2)

	// Fetch archive in parallel
	go func() {
		defer wg.Done()
		archiveData, archiveErr = u.downloadBytes(ctx, downloadURL)
	}()

	// Fetch checksums in parallel
	go func() {
		defer wg.Done()
		hash, err := u.fetchChecksum(ctx, checksumURL, assetName)
		if err == nil {
			expectedHash = hash
		}
	}()

	wg.Wait()

	if archiveErr != nil {
		return fmt.Errorf("failed to download release package: %w", archiveErr)
	}

	// In-flight SHA256 validation
	if expectedHash != "" {
		progress("Validating SHA256 cryptographic signature...")
		hasher := sha256.New()
		hasher.Write(archiveData)
		actualHash := hex.EncodeToString(hasher.Sum(nil))

		if !strings.EqualFold(expectedHash, actualHash) {
			return fmt.Errorf("security error: SHA256 checksum mismatch (expected %s, got %s)", expectedHash, actualHash)
		}
	}

	// Extract binary from stream
	progress("Extracting binary executable...")
	binName := "hop"
	if osName == "windows" {
		binName = "hop.exe"
	}

	newBinaryBytes, err := u.extractBinary(archiveData, osName, binName)
	if err != nil {
		return fmt.Errorf("failed to extract binary from archive: %w", err)
	}

	// In-place atomic binary replacement
	progress("Applying atomic binary swap...")
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to determine current executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("unable to resolve binary symlink: %w", err)
	}

	if err := u.replaceBinary(execPath, newBinaryBytes); err != nil {
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	return nil
}

func (u *UpdateService) downloadBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s for %s", resp.Status, url)
	}

	return io.ReadAll(resp.Body)
}

func (u *UpdateService) fetchChecksum(ctx context.Context, url, assetName string) (string, error) {
	data, err := u.downloadBytes(ctx, url)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.Contains(fields[1], assetName) {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("asset %s not found in checksums.txt", assetName)
}

func (u *UpdateService) extractBinary(data []byte, osName, binName string) ([]byte, error) {
	if osName == "windows" {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == binName {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
	} else {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if filepath.Base(hdr.Name) == binName {
				return io.ReadAll(tr)
			}
		}
	}
	return nil, fmt.Errorf("%s not found inside archive", binName)
}

func (u *UpdateService) replaceBinary(execPath string, newBytes []byte) error {
	dir := filepath.Dir(execPath)
	tempFile, err := os.CreateTemp(dir, "hop-update-*.tmp")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	if _, err := tempFile.Write(newBytes); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(0755); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	// On Windows, running .exe cannot be overwritten directly.
	// We rename current executable to .old, then move the new binary in.
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("failed to move running binary to .old: %w", err)
		}
		if err := os.Rename(tempName, execPath); err != nil {
			// Rollback if move fails
			_ = os.Rename(oldPath, execPath)
			return fmt.Errorf("failed to move new binary in place: %w", err)
		}
		_ = os.Remove(oldPath) // Attempt immediate cleanup
		return nil
	}

	// On Unix, atomic rename replaces inode directly
	return os.Rename(tempName, execPath)
}
