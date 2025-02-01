package utils

import (
    "os"
    "path/filepath"
    "strings"
    "fmt"
)

func RenameImage(filePath, dateTime, checksum string) error {
    dir := filepath.Dir(filePath)
    ext := strings.ToLower(filepath.Ext(filePath))
    newFilename := fmt.Sprintf("%s_%s%s", dateTime, checksum, ext)
    newFilePath := filepath.Join(dir, newFilename)

    return os.Rename(filePath, newFilePath)
}
