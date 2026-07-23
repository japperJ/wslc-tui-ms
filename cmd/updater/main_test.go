package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFileCopiesUpdaterBinary(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.exe")
	destination := filepath.Join(t.TempDir(), "destination.exe")
	contents := []byte("updater-binary")
	if err := os.WriteFile(source, contents, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(source, destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(contents) {
		t.Fatalf("copied contents = %q, want %q", got, contents)
	}
}
