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

func TestDeriveMetadataUsesEpochForUnparseableExif(t *testing.T) {
	t.Parallel()

	file, err := os.CreateTemp(t.TempDir(), "not-a-jpeg-*.jpg")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	file.Close()

	metadata, err := DeriveMetadata(file.Name(), false)
	if err != nil {
		t.Fatalf("DeriveMetadata: %v", err)
	}

	if metadata.Timestamp != "19700101_000000" {
		t.Fatalf("Timestamp = %q, want %q", metadata.Timestamp, "19700101_000000")
	}
	if metadata.CameraID != "UNK" {
		t.Fatalf("CameraID = %q, want %q", metadata.CameraID, "UNK")
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
