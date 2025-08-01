package config

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

type Config struct {
	SourceDirectory string
	TargetDirectory string
	LogLevel        string
}

var (
	instance *Config
	once     sync.Once
)

func InitFromFlags() {
	once.Do(func() {
		sourceDirectory := flag.String("source", "", "Directory containing the images to process")
		targetDirectory := flag.String("target", "", "Directory to move processed images to")
		logLevel := flag.String("loglevel", "info", "Log level: debug, info, warn, error")

		flag.Parse()

		if *sourceDirectory == "" {
			log.Error("Error: -source flag is required")
			flag.Usage()
			os.Exit(1)
		}

		instance = &Config{
			SourceDirectory: *sourceDirectory,
			TargetDirectory: *targetDirectory,
			LogLevel:        *logLevel,
		}
	})
}

func GetConfig() *Config {
	if instance == nil {
		panic("Config not initialized: call config.InitFromFlags() in main()")
	}
	return instance
}
