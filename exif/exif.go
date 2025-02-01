package exif

import (
    "os"
    "github.com/rwcarlsen/goexif/exif"
)

func ExtractExifDateTime(filePath string) (string, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    x, err := exif.Decode(f)
    if err != nil {
        return "", err
    }

    tm, err := x.DateTime()
    if err != nil {
        return "", err
    }

    return tm.Format("20060102_150405"), nil
}
