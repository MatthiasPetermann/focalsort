package exif

import (
	"os"
	"testing"
	"time"
)

func TestFormatTimestampPreservesLocalTime(t *testing.T) {
	t.Parallel()

	local := time.Date(2025, 1, 2, 3, 4, 5, 0, time.FixedZone("X", 2*3600))
	got := FormatTimestamp(local)

	if got != "20250102_030405" {
		t.Fatalf("FormatTimestamp = %s, want %s", got, "20250102_030405")
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

func TestDeriveMetadataUsesEpochForInvalidExifTimestamp(t *testing.T) {
	t.Parallel()

	filePath := t.TempDir() + "/invalid-date.jpg"
	jpeg := []byte{
		0xff, 0xd8, 0xff, 0xe1, 0x00, 0x36,
		'E', 'x', 'i', 'f', 0x00, 0x00,
		'M', 'M', 0x00, 0x2a, 0x00, 0x00, 0x00, 0x08,
		0x00, 0x01,
		0x01, 0x32, 0x00, 0x02, 0x00, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00, 0x1a,
		0x00, 0x00, 0x00, 0x00,
		'2', '0', '2', '5', ':', '1', '3', ':', '0', '1', ' ', '0', '0', ':', '0', '0', ':', '0', '0', 0x00,
		0xff, 0xd9,
	}
	if err := os.WriteFile(filePath, jpeg, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	metadata, err := DeriveMetadata(filePath, false)
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
