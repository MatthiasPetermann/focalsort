package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameImageDryRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	source := filepath.Join(dir, "img.jpg")
	if err := os.WriteFile(source, []byte("x"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	target, err := RenameImage(source, "20250101_101010", "SON1A2B", "Q37", "abc123", true)
	if err != nil {
		t.Fatalf("RenameImage dry-run: %v", err)
	}

	if _, err := os.Stat(source); err != nil {
		t.Fatalf("source file missing in dry-run: %v", err)
	}

	if filepath.Base(target) != "20250101_101010_SON1A2B_Q37_abc123.jpg" {
		t.Fatalf("unexpected target: %s", target)
	}
}

func TestRenameImageCollisionResolution(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	source := filepath.Join(dir, "source.jpg")
	if err := os.WriteFile(source, []byte("x"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	collision := filepath.Join(dir, "20250101_101010_SON1A2B_Q37_abc123.jpg")
	if err := os.WriteFile(collision, []byte("y"), 0o644); err != nil {
		t.Fatalf("write collision: %v", err)
	}

	target, err := RenameImage(source, "20250101_101010", "SON1A2B", "Q37", "abc123", false)
	if err != nil {
		t.Fatalf("RenameImage collision: %v", err)
	}

	if filepath.Base(target) != "20250101_101010_SON1A2B_Q37_abc123_0001.jpg" {
		t.Fatalf("unexpected collision target: %s", target)
	}

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("renamed file missing: %v", err)
	}
}
