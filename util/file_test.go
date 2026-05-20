package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileSafelyCreatesParentDirectoriesAndWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "file.txt")

	if err := WriteFileSafely(path, []byte("content"), 0640); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "content" {
		t.Fatalf("expected written content, got %q", string(body))
	}
	if stat, err := os.Stat(path); err != nil {
		t.Fatal(err)
	} else if stat.Mode().Perm() != 0640 {
		t.Fatalf("expected file mode 0640, got %v", stat.Mode().Perm())
	}
}

func TestWriteFileSafelyReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := WriteFileSafely(path, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "new" {
		t.Fatalf("expected replacement content, got %q", string(body))
	}
}

func TestWriteFileSafelyReturnsMkdirError(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}

	err := WriteFileSafely(filepath.Join(blocker, "child.txt"), []byte("content"), 0644)
	if err == nil {
		t.Fatal("expected parent directory creation to fail")
	}
}
