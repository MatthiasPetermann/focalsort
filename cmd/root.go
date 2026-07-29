package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"focalsort/exif"
	"focalsort/hash"
	"focalsort/imageproc"
	"focalsort/tui"
	"focalsort/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var importFolder string
var useTUI bool
var recursive bool
var dryRun bool
var fallbackToModTime bool
var checksumLength int

var rootCmd = &cobra.Command{
	Use:   "focalsort --import-folder <path>",
	Short: "Analyze JPG images and rename them deterministically",
	Long:  "FocalSort analyzes JPG/JPEG images and renames files in deterministic order.",
	Example: "focalsort --import-folder ./photos\n" +
		"focalsort --import-folder ./photos --recursive=false\n" +
		"focalsort --import-folder ./photos --dry-run\n" +
		"focalsort --import-folder ./photos --fallback-to-modtime",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&importFolder, "import-folder", "i", "", "Path to import folder (required)")
	rootCmd.Flags().BoolVarP(&useTUI, "tui", "t", false, "Enable TUI mode")
	rootCmd.Flags().BoolVar(&recursive, "recursive", true, "Process files recursively")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print changes without renaming files")
	rootCmd.Flags().BoolVar(&fallbackToModTime, "fallback-to-modtime", false, "Use file modification time when EXIF timestamp is missing")
	rootCmd.Flags().IntVar(&checksumLength, "checksum-length", 16, "Length of checksum suffix in output filename (1-40)")

	if err := rootCmd.MarkFlagRequired("import-folder"); err != nil {
		panic(err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Error executing command: %v", err)
	}
}

func run() error {
	if checksumLength < 1 || checksumLength > 40 {
		return fmt.Errorf("invalid checksum-length %d, expected 1-40", checksumLength)
	}

	files, err := collectImageFiles(importFolder, recursive)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		logMessage("No JPG/JPEG files found")
		return nil
	}

	if useTUI {
		tui.StartTUI()
		defer tui.StopTUI()
		tui.SetRunContext(importFolder, recursive, dryRun, fallbackToModTime, checksumLength)
	}

	logMessage(fmt.Sprintf("Processing %d images in folder: %s", len(files), importFolder))

	successCount := 0
	failedCount := 0
	skippedCount := 0
	alreadyNamedCount := 0
	processedCount := 0

	for _, path := range files {
		if useTUI {
			tui.SetCurrentFile(path)
		}
		logMessage("Processing file: " + filepath.Base(path))

		metadata, err := exif.DeriveMetadata(path, fallbackToModTime)
		if err != nil {
			failedCount++
			setProcessingError(fmt.Sprintf("metadata: %v", err))
			logError("Error deriving metadata", err)
			updateCounters(successCount, failedCount, skippedCount, alreadyNamedCount)
			continue
		}

		checksum, err := hash.CalculateChecksum(path)
		if err != nil {
			failedCount++
			setProcessingError(fmt.Sprintf("checksum: %v", err))
			logError("Error calculating checksum", err)
			updateCounters(successCount, failedCount, skippedCount, alreadyNamedCount)
			continue
		}

		sharpness, err := imageproc.EvaluateImageQuality(path)
		qualityCode := imageproc.CompactQualityCode(sharpness)
		if err != nil {
			skippedCount++
			qualityCode = "Q00"
			logMessage(fmt.Sprintf("Skipping image quality for %s: %v", filepath.Base(path), err))
		}

		shortChecksum := hash.ShortChecksum(checksum, checksumLength)
		newPath, unchanged, err := utils.RenameImage(path, metadata.Timestamp, metadata.CameraID, qualityCode, shortChecksum, dryRun)
		if err != nil {
			failedCount++
			setProcessingError(fmt.Sprintf("rename: %v", err))
			logError("Error renaming image", err)
			updateCounters(successCount, failedCount, skippedCount, alreadyNamedCount)
			continue
		}

		if unchanged {
			alreadyNamedCount++
			logMessage(fmt.Sprintf("Already named correctly: %s", filepath.Base(path)))
		} else {
			successCount++
			setLastRename(path, newPath)
		}
		processedCount++
		clearLastError()
		if dryRun {
			logMessage(fmt.Sprintf("Dry-run rename: %s -> %s", path, newPath))
		}

		updateProgress(processedCount, len(files))
		updateCounters(successCount, failedCount, skippedCount, alreadyNamedCount)
	}

	logMessage(fmt.Sprintf("Processing complete: success=%d failed=%d skipped=%d already-named=%d", successCount, failedCount, skippedCount, alreadyNamedCount))
	return nil
}

func collectImageFiles(root string, recursive bool) ([]string, error) {
	files := make([]string, 0)

	if recursive {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if isImageFile(path) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(root, entry.Name())
			if isImageFile(path) {
				files = append(files, path)
			}
		}
	}

	sort.Strings(files)
	return files, nil
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg"
}

func logMessage(msg string) {
	if useTUI {
		tui.LogMessage(msg)
		return
	}
	logrus.Info(msg)
}

func logError(context string, err error) {
	logMessage(fmt.Sprintf("%s: %v", context, err))
}

func setProcessingError(message string) {
	if useTUI {
		tui.SetLastError(message)
	}
}

func clearLastError() {
	if useTUI {
		tui.SetLastError("")
	}
}

func setLastRename(from string, to string) {
	if useTUI {
		tui.SetLastRename(from, to)
	}
}

func updateCounters(success int, failed int, skipped int, alreadyNamed int) {
	if useTUI {
		tui.UpdateCounters(success, failed, skipped, alreadyNamed)
	}
}

func updateProgress(processed int, total int) {
	if useTUI {
		tui.UpdateStatus(processed, total)
		return
	}
	logrus.Infof("Progress: %d/%d", processed, total)
}
