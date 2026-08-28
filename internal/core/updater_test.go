package core

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUpdateService_ExtractBinary(t *testing.T) {
	service := NewUpdateService()

	// Create test zip archive containing hop.exe
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	binName := "hop"
	if runtime.GOOS == "windows" {
		binName = "hop.exe"
	}

	w, err := zw.Create(binName)
	if err != nil {
		t.Fatalf("zip create error: %v", err)
	}
	expectedContent := []byte("binary payload mock data")
	if _, err := w.Write(expectedContent); err != nil {
		t.Fatalf("zip write error: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close error: %v", err)
	}

	extracted, err := service.extractBinary(buf.Bytes(), runtime.GOOS, binName)
	if err != nil {
		t.Fatalf("extractBinary error: %v", err)
	}

	if !bytes.Equal(extracted, expectedContent) {
		t.Errorf("expected %q, got %q", string(expectedContent), string(extracted))
	}
}

func TestUpdateService_ReplaceBinary(t *testing.T) {
	service := NewUpdateService()
	tempDir, err := os.MkdirTemp("", "hop-replace-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	binName := "hop"
	if runtime.GOOS == "windows" {
		binName = "hop.exe"
	}
	execPath := filepath.Join(tempDir, binName)

	// Write initial mock binary
	initialData := []byte("version 1.0.0")
	if err := os.WriteFile(execPath, initialData, 0755); err != nil {
		t.Fatalf("write initial binary error: %v", err)
	}

	// In-place atomic replacement
	newData := []byte("version 1.2.0 upgraded")
	if err := service.replaceBinary(execPath, newData); err != nil {
		t.Fatalf("replaceBinary error: %v", err)
	}

	// Verify replacement
	content, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("read replaced binary error: %v", err)
	}

	if !bytes.Equal(content, newData) {
		t.Errorf("expected %q, got %q", string(newData), string(content))
	}
}

func TestUpdateService_CheckUpdate(t *testing.T) {
	service := NewUpdateService()
	ctx := context.Background()

	// Live network test (gracefully skips if offline)
	tag, _, err := service.CheckUpdate(ctx)
	if err != nil {
		t.Logf("network check skipped or failed: %v", err)
		return
	}
	if tag == "" {
		t.Errorf("expected non-empty tag")
	}
}
