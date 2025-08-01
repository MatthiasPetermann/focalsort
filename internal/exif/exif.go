package exif

import (
	"github.com/rwcarlsen/goexif/exif"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"os"
)

func ExtractExifData(filePath string) (string, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return "", "", err
	}

	tm, err := x.DateTime()
	if err != nil {
		return "", "", err
	}

	jsonByte, err := x.MarshalJSON()
	if err != nil {
		log.Error(err.Error())
	}

	jsonString := string(jsonByte)
	camera := gjson.Get(jsonString, "Model").String()

	return tm.Format("20060102_150405"), camera, nil
}
