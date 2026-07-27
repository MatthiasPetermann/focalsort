package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RenameImage(filePath, timestamp, cameraID, qualityCode, checksum string, dryRun bool) (string, error) {
	dir := filepath.Dir(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := fmt.Sprintf("%s_%s_%s_%s", timestamp, cameraID, qualityCode, checksum)

	candidatePath := filepath.Join(dir, baseName+ext)
	if filePath == candidatePath {
		return candidatePath, nil
	}

	collisionResolvedPath, err := uniqueTargetPath(filePath, candidatePath, ext)
	if err != nil {
		return "", err
	}

	if dryRun {
		return collisionResolvedPath, nil
	}

	if filePath == collisionResolvedPath {
		return collisionResolvedPath, nil
	}

	return collisionResolvedPath, os.Rename(filePath, collisionResolvedPath)
}

func uniqueTargetPath(sourcePath, preferredPath, ext string) (string, error) {
	exists, sameFile, err := targetState(sourcePath, preferredPath)
	if err != nil {
		return "", err
	}
	if !exists || sameFile {
		return preferredPath, nil
	}

	dir := filepath.Dir(preferredPath)
	name := strings.TrimSuffix(filepath.Base(preferredPath), ext)

	for i := 1; i <= 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%04d%s", name, i, ext))
		exists, sameFile, err := targetState(sourcePath, candidate)
		if err != nil {
			return "", err
		}
		if !exists || sameFile {
			return candidate, nil
		}
	}

	return "", errors.New("unable to resolve filename collision")
}

func targetState(sourcePath, candidatePath string) (exists bool, sameFile bool, err error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return false, false, err
	}

	targetInfo, err := os.Stat(candidatePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	return true, os.SameFile(sourceInfo, targetInfo), nil
}
