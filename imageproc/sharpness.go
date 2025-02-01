package imageproc

import (
    "image"
    "image/color"
    "image/jpeg"
    "os"
    "math"
)

func LoadAndGrayscale(filePath string) (*image.Gray, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    img, err := jpeg.Decode(file)
    if err != nil {
        return nil, err
    }

    bounds := img.Bounds()
    grayImg := image.NewGray(bounds)

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            grayColor := color.GrayModel.Convert(img.At(x, y))
            grayImg.Set(x, y, grayColor)
        }
    }

    return grayImg, nil
}

func EvaluateImageQuality(filePath string) float64 {
    img, err := LoadAndGrayscale(filePath)
    if err != nil {
        return 0
    }

    bounds := img.Bounds()
    totalVariance := 0.0
    pixelCount := 0.0

    for y := 1; y < bounds.Max.Y-1; y++ {
        for x := 1; x < bounds.Max.X-1; x++ {
            center := float64(img.GrayAt(x, y).Y)
            laplaceSum := -4 * center +
                float64(img.GrayAt(x-1, y).Y) +
                float64(img.GrayAt(x+1, y).Y) +
                float64(img.GrayAt(x, y-1).Y) +
                float64(img.GrayAt(x, y+1).Y)

            totalVariance += laplaceSum * laplaceSum
            pixelCount++
        }
    }

    return math.Sqrt(totalVariance / pixelCount)
}
