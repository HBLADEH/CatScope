//go:build windows

package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceExecutable(t *testing.T) {
	dir, err := os.MkdirTemp("", "catscope-update-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	target := filepath.Join(dir, "CatScope.exe")
	source := filepath.Join(dir, "CatScope.exe.new")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	backup, err := replaceExecutable(target, source)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("target contents = %q", got)
	}
	old, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if string(old) != "old" {
		t.Fatalf("backup contents = %q", old)
	}
}

func TestUpdateDirectoryValidation(t *testing.T) {
	dir, err := os.MkdirTemp("", "catscope-update-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if !isUpdateDirectory(dir) {
		t.Fatalf("expected valid update directory: %s", dir)
	}
	if isUpdateDirectory(filepath.Join(t.TempDir(), "catscope-update-fake")) {
		t.Fatal("nested directory must not be accepted")
	}
}
