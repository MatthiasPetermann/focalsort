// File: main.go
package main

import (
	"fmt"
	"github.com/sigurn/crc8"
	"os"
	"path/filepath"
	"strings"

	"d2ux.net/focalsort/internal/config"
	"d2ux.net/focalsort/internal/exif"
	"d2ux.net/focalsort/internal/hash"
	_ "d2ux.net/focalsort/internal/imaging"
	"d2ux.net/focalsort/internal/util"
	"github.com/sirupsen/logrus"
)

// init sets up logging.
// This is run automatically before main().
func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)
}

// main is the CLI entry point.
// Sets up command-line flags, environment variable bindings, and delegates to the engine.
func main() {

	config.InitFromFlags()
	cfg := config.GetConfig()

	processImages(cfg.SourceDirectory)
}

func processImages(importFolder string) {

	var processedFiles = 0
	var totalFiles int

	files, err := os.ReadDir(importFolder)
	if err != nil {
		logrus.Fatal("Error reading directory:", err)
	}

	totalFiles = len(files)

	err = filepath.Walk(importFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(info.Name())) == ".jpg" {
			logMsg := "Processing file: " + info.Name()
			logrus.Info(logMsg)

			dateTime, camera, err := exif.ExtractExifData(path)
			if err != nil {
				logrus.Error("Error extracting EXIF date: " + err.Error())
				return nil
			}

			table := crc8.MakeTable(crc8.CRC8)
			cameraCrc := crc8.Checksum([]byte(camera), table)
			camera = fmt.Sprintf("%02X", cameraCrc)

			checksum, err := hash.CalculateChecksum(path)
			if err != nil {
				logrus.Error("Error calculating checksum: " + err.Error())
				return nil
			}

			//			sharpness := imaging.EvaluateImageQuality(path)

			err = utils.RenameImage(path, config.GetConfig().TargetDirectory, dateTime, camera, checksum)
			if err != nil {
				logrus.Error("Error renaming: " + err.Error())
				return nil
			}

			processedFiles++
			logrus.Infof("Progress: %d/%d", processedFiles, totalFiles)
		}
		return nil
	})

	if err != nil {
		logrus.Fatalf("Error processing import folder: %v", err)
	}
}
