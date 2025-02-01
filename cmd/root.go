package cmd

import (
    "focalsort/exif"
    "focalsort/hash"
    "focalsort/imageproc"
    "focalsort/utils"
    "focalsort/tui"
    "github.com/spf13/cobra"
    "github.com/sirupsen/logrus"
    "path/filepath"
    "strings"
    "os"
)

var importFolder string
var useTUI bool
var totalFiles int
var processedFiles int

var rootCmd = &cobra.Command{
    Use:   "focalsort",
    Short: "Image sorting and analysis tool with optional TUI",
    Run: func(cmd *cobra.Command, args []string) {
        if importFolder == "" {
            logrus.Fatal("No import folder specified. Use --import-folder")
        }

        files, err := os.ReadDir(importFolder)
        if err != nil {
            logrus.Fatal("Error reading directory:", err)
        }

        totalFiles = len(files)
        processedFiles = 0

        if useTUI {
            tui.StartTUI()
            tui.LogMessage("Processing images in folder: " + importFolder)
        } else {
            logrus.Infof("Processing images in folder: %s", importFolder)
        }

        processImages(importFolder)

        if useTUI {
            tui.LogMessage("Processing complete!")
            tui.WaitForExit()
        }
    },
}

func init() {
    rootCmd.Flags().StringVarP(&importFolder, "import-folder", "i", "", "Path to import folder (required)")
    rootCmd.Flags().BoolVarP(&useTUI, "tui", "t", false, "Enable TUI mode")
    rootCmd.MarkFlagRequired("import-folder")
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        logrus.Fatalf("Error executing command: %v", err)
    }
}

func processImages(importFolder string) {
    err := filepath.Walk(importFolder, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() && strings.ToLower(filepath.Ext(info.Name())) == ".jpg" {
            logMsg := "Processing file: " + info.Name()
            if useTUI {
                tui.LogMessage(logMsg)
            } else {
                logrus.Info(logMsg)
            }

            dateTime, err := exif.ExtractExifDateTime(path)
            if err != nil {
                tui.LogMessage("Error extracting EXIF date: " + err.Error())
                return nil
            }

            checksum, err := hash.CalculateChecksum(path)
            if err != nil {
                tui.LogMessage("Error calculating checksum: " + err.Error())
                return nil
            }

            sharpness := imageproc.EvaluateImageQuality(path)

            err = utils.SetXattrs(path, dateTime, sharpness, checksum)
            if err != nil {
                tui.LogMessage("Error setting xattrs: " + err.Error())
                return nil
            }

            err = utils.RenameImage(path, dateTime, checksum)
            if err != nil {
                tui.LogMessage("Error renaming: " + err.Error())
                return nil
            }

            processedFiles++
            if useTUI {
                tui.UpdateStatus(processedFiles, totalFiles)
            } else {
                logrus.Infof("Progress: %d/%d", processedFiles, totalFiles)
            }
        }
        return nil
    })

    if err != nil {
        logrus.Fatalf("Error processing import folder: %v", err)
    }
}
