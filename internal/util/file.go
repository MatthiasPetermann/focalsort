package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RenameImage(filePath, targetDirectory, dateTime, camera, checksum string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	newFilename := fmt.Sprintf("%s_%s_%s%s", dateTime, camera, checksum, ext)

	var destDir string
	if targetDirectory != "" {
		if err := os.MkdirAll(targetDirectory, os.ModePerm); err != nil {
			return fmt.Errorf("could not create target path: %w", err)
		}

		t, err := time.Parse("20060102_150405", dateTime)
		if err != nil {
			return fmt.Errorf("invalid dateTime format: %w", err)
		}

		subDir := filepath.Join(targetDirectory, fmt.Sprintf("%04d", t.Year()), fmt.Sprintf("%02d", int(t.Month())))

		if err := os.MkdirAll(subDir, os.ModePerm); err != nil {
			return fmt.Errorf("could not create subdirectory: %w", err)
		}

		destDir = subDir
	} else {
		destDir = filepath.Dir(filePath)
	}

	newFilePath := filepath.Join(destDir, newFilename)
	return os.Rename(filePath, newFilePath)
}
