package imageproc

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateImageQualityTooSmall(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "small.jpg")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	if err := jpeg.Encode(file, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	file.Close()

	if _, err := EvaluateImageQuality(path); err == nil {
		t.Fatalf("expected error for too-small image")
	}
}

func TestEvaluateImageQualityValidImage(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "valid.jpg")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	img := image.NewGray(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x+y)%2 == 0 {
				img.SetGray(x, y, color.Gray{Y: 10})
			} else {
				img.SetGray(x, y, color.Gray{Y: 240})
			}
		}
	}

	if err := jpeg.Encode(file, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	file.Close()

	got, err := EvaluateImageQuality(path)
	if err != nil {
		t.Fatalf("EvaluateImageQuality: %v", err)
	}

	if got <= 0 {
		t.Fatalf("expected positive sharpness, got %f", got)
	}
}

func TestCompactQualityCode(t *testing.T) {
	t.Parallel()

	if got := CompactQualityCode(0); got != "Q00" {
		t.Fatalf("CompactQualityCode(0)=%q", got)
	}
	if got := CompactQualityCode(152.2); got != "Q90" {
		t.Fatalf("CompactQualityCode(152.2)=%q", got)
	}
	if got := CompactQualityCode(99999); got != "Q99" {
		t.Fatalf("CompactQualityCode(99999)=%q", got)
	}
}
