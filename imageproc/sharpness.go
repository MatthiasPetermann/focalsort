package imageproc

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"

	"github.com/anthonynsimon/bild/effect"
)

func LoadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func EvaluateImageQuality(filePath string) (float64, error) {
	img, err := LoadImage(filePath)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	if bounds.Dx() < 3 || bounds.Dy() < 3 {
		return 0, errors.New("image too small for sharpness evaluation")
	}

	edges := effect.Sobel(img)
	total := 0.0
	count := 0.0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := edges.At(x, y).RGBA()
			rf := float64(r >> 8)
			gf := float64(g >> 8)
			bf := float64(b >> 8)
			luma := 0.299*rf + 0.587*gf + 0.114*bf
			total += luma
			count++
		}
	}

	if count == 0 {
		return 0, errors.New("no pixels available for sharpness evaluation")
	}

	return total / count, nil
}

func CompactQualityCode(sharpness float64) string {
	if sharpness < 0 || math.IsNaN(sharpness) || math.IsInf(sharpness, 0) {
		return "Q00"
	}

	normalized := math.Log1p(sharpness) / math.Log1p(255)
	scaled := int(math.Round(normalized * 99))
	if scaled < 0 {
		scaled = 0
	}
	if scaled > 99 {
		scaled = 99
	}
	return fmt.Sprintf("Q%02d", scaled)
}
