package utils

import (
	"fmt"
	"golang.org/x/sys/unix"
)

// Set extended attributes on FreeBSD
func SetXattrs(filePath string, dateTime string, sharpness float64, checksum string) error {
	err := unix.Setxattr(filePath, "user.focalsort.exif_date", []byte(dateTime), 0)
	if err != nil {
		return err
	}
	err = unix.Setxattr(filePath, "user.focalsort.sharpness", []byte(fmt.Sprintf("%f", sharpness)), 0)
	if err != nil {
		return err
	}
	err = unix.Setxattr(filePath, "user.focalsort.checksum", []byte(checksum), 0)
	return err
}
