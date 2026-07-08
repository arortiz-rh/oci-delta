package ocidelta

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFilesystemDataSource(t *testing.T) {
	tmpDir := t.TempDir()

	ds := NewFilesystemDataSource(tmpDir)
	if ds == nil {
		t.Fatal("NewFilesystemDataSource returned nil")
	}

	if err := ds.Cleanup(); err != nil {
		t.Errorf("Cleanup() failed: %v", err)
	}

	if err := ds.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestSimpleDataSourceCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Close()

	if err := ds.Cleanup(); err != nil {
		t.Errorf("first Cleanup() = %v, want nil", err)
	}

	// Test idempotency - calling multiple times should be safe
	if err := ds.Cleanup(); err != nil {
		t.Errorf("second Cleanup() = %v, want nil", err)
	}

	if err := ds.Cleanup(); err != nil {
		t.Errorf("third Cleanup() = %v, want nil", err)
	}
}

func TestSimpleDataSourceWithValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file in the directory
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("hello world")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	// Set current file
	if err := ds.SetCurrentFile("test.txt"); err != nil {
		t.Fatalf("SetCurrentFile() failed: %v", err)
	}

	// Read the file content
	buf := make([]byte, len(testContent))

	n, err := ds.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Read() failed: %v", err)
	}

	if n != len(testContent) {
		t.Errorf("Read() returned %d bytes, want %d", n, len(testContent))
	}

	if string(buf) != string(testContent) {
		t.Errorf("Read() content = %q, want %q", buf, testContent)
	}
}

func TestSimpleDataSourceSeek(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "seektest.txt")
	testContent := []byte("0123456789")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	if err := ds.SetCurrentFile("seektest.txt"); err != nil {
		t.Fatalf("SetCurrentFile() failed: %v", err)
	}

	// Seek to offset 5
	pos, err := ds.Seek(5, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek() failed: %v", err)
	}

	if pos != 5 {
		t.Errorf("Seek() position = %d, want 5", pos)
	}

	// Read from offset 5
	buf := make([]byte, 5)

	n, err := ds.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Read() after Seek() failed: %v", err)
	}

	expected := "56789"
	if string(buf[:n]) != expected {
		t.Errorf("Read() after Seek() = %q, want %q", buf[:n], expected)
	}

	// Seek back to start
	pos, err = ds.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek() to start failed: %v", err)
	}

	if pos != 0 {
		t.Errorf("Seek() to start position = %d, want 0", pos)
	}
}

func TestSimpleDataSourceClose(t *testing.T) {
	tmpDir := t.TempDir()
	ds := NewFilesystemDataSource(tmpDir)

	if err := ds.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}

	// Test idempotency - calling multiple times should be safe
	if err := ds.Close(); err != nil {
		t.Errorf("second Close() = %v, want nil", err)
	}
}

func TestSimpleDataSourceReadAfterClose(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()

	if err := ds.SetCurrentFile("test.txt"); err != nil {
		t.Fatalf("SetCurrentFile() failed: %v", err)
	}

	if err := ds.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	buf := make([]byte, 5)

	_, err := ds.Read(buf)
	if err == nil {
		t.Error("Read() after Close() should return an error")
	}
}

func TestSimpleDataSourceMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %q: %v", name, err)
		}
	}

	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	// Test reading each file
	for name, expectedContent := range files {
		if err := ds.SetCurrentFile(name); err != nil {
			t.Errorf("SetCurrentFile(%q) = %v, want nil", name, err)
			continue
		}

		buf := make([]byte, len(expectedContent))

		n, err := ds.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Errorf("Read() for %q failed: %v", name, err)
			continue
		}

		if string(buf[:n]) != expectedContent {
			t.Errorf("Read() for %q = %q, want %q", name, buf[:n], expectedContent)
		}
	}
}

func TestSimpleDataSourceReadWithoutSetCurrentFile(t *testing.T) {
	tmpDir := t.TempDir()
	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	buf := make([]byte, 10)

	_, err := ds.Read(buf)
	if err == nil {
		t.Error("Read() without SetCurrentFile() should return an error")
	}
}

func TestSimpleDataSourceSeekWithoutSetCurrentFile(t *testing.T) {
	tmpDir := t.TempDir()
	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	_, err := ds.Seek(0, io.SeekStart)
	if err == nil {
		t.Error("Seek() without SetCurrentFile() should return an error")
	}
}

func TestSimpleDataSourceNonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	// Try to set a nonexistent file
	err := ds.SetCurrentFile("does-not-exist.txt")
	if err == nil {
		t.Error("SetCurrentFile() for nonexistent file should fail")
	}
}

func TestSimpleDataSourceSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory with file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	testFile := filepath.Join(subDir, "nested.txt")
	testContent := []byte("nested content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to write nested test file: %v", err)
	}

	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	// Access file in subdirectory
	if err := ds.SetCurrentFile("subdir/nested.txt"); err != nil {
		t.Fatalf("SetCurrentFile() for nested file failed: %v", err)
	}

	buf := make([]byte, len(testContent))

	n, err := ds.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Read() for nested file failed: %v", err)
	}

	if string(buf[:n]) != string(testContent) {
		t.Errorf("Read() for nested file = %q, want %q", buf[:n], testContent)
	}
}

func TestSimpleDataSourceEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty file
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to write empty test file: %v", err)
	}

	ds := NewFilesystemDataSource(tmpDir)
	defer ds.Cleanup()
	defer ds.Close()

	if err := ds.SetCurrentFile("empty.txt"); err != nil {
		t.Fatalf("SetCurrentFile() for empty file failed: %v", err)
	}

	buf := make([]byte, 10)

	n, err := ds.Read(buf)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Read() for empty file error = %v, want io.EOF", err)
	}

	if n != 0 {
		t.Errorf("Read() for empty file returned %d bytes, want 0", n)
	}
}
