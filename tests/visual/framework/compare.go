// Package framework provides image comparison utilities for visual testing.
package framework

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// CompareImages compares two images and returns detailed comparison results.
func CompareImages(reference, generated image.Image, options ComparisonOptions) *ComparisonResult {
	refBounds := reference.Bounds()
	genBounds := generated.Bounds()

	// Check if dimensions match
	if refBounds.Size() != genBounds.Size() {
		return &ComparisonResult{
			Passed:            false,
			DifferentPixels:   refBounds.Size().X * refBounds.Size().Y, // All pixels different
			TotalPixels:       refBounds.Size().X * refBounds.Size().Y,
			MaxDifference:     255,
			AverageDifference: 255.0,
		}
	}

	width := refBounds.Size().X
	height := refBounds.Size().Y
	totalPixels := width * height

	var diffImage *image.RGBA
	if options.GenerateDiffImage {
		diffImage = image.NewRGBA(image.Rect(0, 0, width, height))
	}

	differentPixels := 0
	var totalDifference float64
	var maxDifference uint8

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			refColor := reference.At(refBounds.Min.X+x, refBounds.Min.Y+y)
			genColor := generated.At(genBounds.Min.X+x, genBounds.Min.Y+y)

			refR, refG, refB, refA := refColor.RGBA()
			genR, genG, genB, genA := genColor.RGBA()

			// Convert from 16-bit to 8-bit
			refR8, refG8, refB8, refA8 := uint8(refR>>8), uint8(refG>>8), uint8(refB>>8), uint8(refA>>8)
			genR8, genG8, genB8, genA8 := uint8(genR>>8), uint8(genG>>8), uint8(genB>>8), uint8(genA>>8)

			// Calculate per-channel differences
			diffR := absDiffUint8(refR8, genR8)
			diffG := absDiffUint8(refG8, genG8)
			diffB := absDiffUint8(refB8, genB8)
			diffA := absDiffUint8(refA8, genA8)

			if options.IgnoreAlpha {
				diffA = 0
			}

			// Find maximum difference across all channels
			pixelMaxDiff := maxUint8(diffR, maxUint8(diffG, maxUint8(diffB, diffA)))
			if pixelMaxDiff > maxDifference {
				maxDifference = pixelMaxDiff
			}

			// Check if pixel is different based on tolerance
			pixelDifferent := false
			if options.ExactMatch {
				pixelDifferent = diffR > 0 || diffG > 0 || diffB > 0 || (!options.IgnoreAlpha && diffA > 0)
			} else {
				pixelDifferent = pixelMaxDiff > options.Tolerance
			}

			if pixelDifferent {
				differentPixels++
			}

			// Accumulate total difference for average calculation
			totalDifference += float64(diffR + diffG + diffB + diffA)

			// Generate diff image pixel
			if diffImage != nil {
				if pixelDifferent {
					// Highlight differences in red
					intensity := uint8(255)
					if pixelMaxDiff < 255 {
						// Scale intensity based on difference magnitude
						intensity = uint8((float64(pixelMaxDiff) / 255.0) * 255)
					}
					diffImage.Set(x, y, color.RGBA{R: intensity, G: 0, B: 0, A: 255})
				} else {
					// Show original pixel with reduced brightness
					grayValue := uint8((float64(refR8) + float64(refG8) + float64(refB8)) / 3.0 * 0.3)
					diffImage.Set(x, y, color.RGBA{R: grayValue, G: grayValue, B: grayValue, A: 255})
				}
			}
		}
	}

	averageDifference := totalDifference / float64(totalPixels*4) // 4 channels per pixel

	return &ComparisonResult{
		Passed:            differentPixels == 0,
		DifferentPixels:   differentPixels,
		TotalPixels:       totalPixels,
		MaxDifference:     maxDifference,
		AverageDifference: averageDifference,
		DiffImage:         diffImage,
	}
}

// CompareImageFiles compares two image files and returns comparison results.
func CompareImageFiles(referencePath, generatedPath string, options ComparisonOptions) (*ComparisonResult, error) {
	// Load reference image
	refFile, err := os.Open(referencePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open reference image %s: %v", referencePath, err)
	}
	defer refFile.Close()

	refImage, _, err := image.Decode(refFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reference image %s: %v", referencePath, err)
	}

	// Load generated image
	genFile, err := os.Open(generatedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open generated image %s: %v", generatedPath, err)
	}
	defer genFile.Close()

	genImage, _, err := image.Decode(genFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode generated image %s: %v", generatedPath, err)
	}

	return CompareImages(refImage, genImage, options), nil
}

// SaveDiffImage saves a diff image to the specified path.
func SaveDiffImage(diffImage *image.RGBA, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create diff image file %s: %v", outputPath, err)
	}
	defer file.Close()

	// Encode as PNG
	if err := png.Encode(file, diffImage); err != nil {
		return fmt.Errorf("failed to encode diff image: %v", err)
	}

	return nil
}

// LoadImage loads an image from a file path.
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image %s: %v", path, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %v", path, err)
	}

	return img, nil
}

// SaveImage saves an image to a PNG file.
func SaveImage(img image.Image, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Create output file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create image file %s: %v", path, err)
	}
	defer file.Close()

	// Encode as PNG
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

// Helper functions

func absDiffUint8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}
