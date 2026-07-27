package exif

import (
	"os"
	"testing"
	"time"
)

func TestFormatTimestampUTC(t *testing.T) {
	t.Parallel()

	local := time.Date(2025, 1, 2, 3, 4, 5, 0, time.FixedZone("X", 2*3600))
	got := FormatTimestamp(local)

	if got != "20250102_010405" {
		t.Fatalf("FormatTimestamp = %s, want %s", got, "20250102_010405")
	}
}

func TestDeriveTimestampFallback(t *testing.T) {
	t.Parallel()

	file, err := os.CreateTemp(t.TempDir(), "not-a-jpeg-*.jpg")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	file.Close()

	if _, err := DeriveTimestamp(file.Name(), false); err == nil {
		t.Fatalf("DeriveTimestamp expected error without fallback")
	}

	got, err := DeriveTimestamp(file.Name(), true)
	if err != nil {
		t.Fatalf("DeriveTimestamp fallback: %v", err)
	}

	if got == "" {
		t.Fatalf("DeriveTimestamp fallback returned empty timestamp")
	}
}

func TestCompactCameraIdentifier(t *testing.T) {
	t.Parallel()

	id := CompactCameraIdentifier("Sony", "ILCE-7M3")
	if len(id) != 6 {
		t.Fatalf("unexpected camera id length %d for %q", len(id), id)
	}
	if id[:3] != "SON" {
		t.Fatalf("unexpected camera id prefix for %q", id)
	}

	unknown := CompactCameraIdentifier("", "")
	if unknown != "UNK" {
		t.Fatalf("expected UNK, got %q", unknown)
	}
}
