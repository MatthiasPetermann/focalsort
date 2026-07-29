package exif

import (
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/rwcarlsen/goexif/exif"
)

type Metadata struct {
	Timestamp string
	CameraID  string
}

func ExtractExifDateTime(filePath string) (string, error) {
	tm, err := ExtractExifTimestamp(filePath)
	if err != nil {
		return "", err
	}
	return FormatTimestamp(tm), nil
}

func ExtractExifTimestamp(filePath string) (time.Time, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, err
	}

	tm, err := x.DateTime()
	if err != nil {
		return time.Time{}, err
	}

	return tm.UTC(), nil
}

func FallbackTimestampFromModTime(filePath string) (time.Time, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime().UTC(), nil
}

func DeriveTimestamp(filePath string, fallbackToModTime bool) (string, error) {
	metadata, err := DeriveMetadata(filePath, fallbackToModTime)
	if err != nil {
		return "", err
	}
	return metadata.Timestamp, nil
}

func DeriveMetadata(filePath string, fallbackToModTime bool) (Metadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return Metadata{}, err
	}
	defer f.Close()

	x, decodeErr := exif.Decode(f)
	if decodeErr != nil {
		return Metadata{Timestamp: FormatTimestamp(time.Unix(0, 0)), CameraID: "UNK"}, nil
	}

	tm, tsErr := x.DateTime()
	if tsErr != nil {
		if !fallbackToModTime {
			return Metadata{}, tsErr
		}
		fallback, modErr := FallbackTimestampFromModTime(filePath)
		if modErr != nil {
			return Metadata{}, errors.Join(tsErr, modErr)
		}
		return Metadata{Timestamp: FormatTimestamp(fallback), CameraID: cameraIdentifier(x)}, nil
	}

	return Metadata{Timestamp: FormatTimestamp(tm), CameraID: cameraIdentifier(x)}, nil
}

func FormatTimestamp(tm time.Time) string {
	return tm.UTC().Format("20060102_150405")
}

func cameraIdentifier(x *exif.Exif) string {
	makeValue := readExifString(x, exif.Make)
	modelValue := readExifString(x, exif.Model)
	return CompactCameraIdentifier(makeValue, modelValue)
}

func CompactCameraIdentifier(makeValue string, modelValue string) string {
	raw := strings.TrimSpace(makeValue + " " + modelValue)
	normalized := normalizeAlphaNumUpper(raw)
	if normalized == "" {
		return "UNK"
	}

	prefix := normalized
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	for len(prefix) < 3 {
		prefix += "X"
	}

	suffix := fmt.Sprintf("%03X", crc32.ChecksumIEEE([]byte(normalized))&0xFFF)
	return prefix + suffix
}

func readExifString(x *exif.Exif, field exif.FieldName) string {
	tag, err := x.Get(field)
	if err != nil {
		return ""
	}
	value, err := tag.StringVal()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(value, "\x00"))
}

func normalizeAlphaNumUpper(value string) string {
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToUpper(r))
		}
	}
	return builder.String()
}
