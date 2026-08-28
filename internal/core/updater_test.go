package core

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUpdateService_ExtractBinary_Zip(t *testing.T) {
	service := NewUpdateService()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	binName := "hop.exe"
	w, err := zw.Create(binName)
	if err != nil {
		t.Fatalf("zip create error: %v", err)
	}
	expectedContent := []byte("binary zip payload data")
	if _, err := w.Write(expectedContent); err != nil {
		t.Fatalf("zip write error: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close error: %v", err)
	}

	extracted, err := service.extractBinary(buf.Bytes(), "windows", binName)
	if err != nil {
		t.Fatalf("extractBinary windows error: %v", err)
	}

	if !bytes.Equal(extracted, expectedContent) {
		t.Errorf("expected %q, got %q", string(expectedContent), string(extracted))
	}
}

func TestUpdateService_ExtractBinary_TarGz(t *testing.T) {
	service := NewUpdateService()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	binName := "hop"
	expectedContent := []byte("binary tar.gz payload data")
	hdr := &tar.Header{
		Name: binName,
		Mode: 0755,
		Size: int64(len(expectedContent)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar write header error: %v", err)
	}
	if _, err := tw.Write(expectedContent); err != nil {
		t.Fatalf("tar write error: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close error: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close error: %v", err)
	}

	extracted, err := service.extractBinary(buf.Bytes(), "linux", binName)
	if err != nil {
		t.Fatalf("extractBinary linux error: %v", err)
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
