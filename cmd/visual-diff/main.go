// Package main writes amplified visual diffs for PNG reference comparisons.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type diffStats struct {
	Width           int
	Height          int
	TotalPixels     int
	DifferentPixels int
	MaxDiff         uint8
	AverageDiff     float64
	RMSE            float64
}

func main() {
	refPath := flag.String("ref", "", "reference PNG")
	genPath := flag.String("gen", "", "generated PNG")
	outPath := flag.String("out", "", "output amplified diff PNG")
	factor := flag.Int("factor", 8, "difference amplification factor")
	flag.Parse()

	if *refPath == "" || *genPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: visual-diff -ref reference.png -gen generated.png -out amplified.png [-factor 8]")
		os.Exit(2)
	}
	if *factor <= 0 {
		fmt.Fprintln(os.Stderr, "-factor must be positive")
		os.Exit(2)
	}

	ref, err := loadPNG(*refPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load reference: %v\n", err)
		os.Exit(1)
	}
	gen, err := loadPNG(*genPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load generated: %v\n", err)
		os.Exit(1)
	}

	stats, err := analyzeImages(ref, gen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "analyze images: %v\n", err)
		os.Exit(1)
	}
	diff, err := amplifiedDiff(ref, gen, *factor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build diff: %v\n", err)
		os.Exit(1)
	}
	if err := savePNG(diff, *outPath); err != nil {
		fmt.Fprintf(os.Stderr, "save diff: %v\n", err)
		os.Exit(1)
	}

	ratio := 0.0
	if stats.TotalPixels > 0 {
		ratio = float64(stats.DifferentPixels) / float64(stats.TotalPixels)
	}
	fmt.Printf(
		"size=%dx%d different_pixels=%d/%d ratio=%.6f max_diff=%d avg_diff=%.4f rmse=%.4f out=%s\n",
		stats.Width,
		stats.Height,
		stats.DifferentPixels,
		stats.TotalPixels,
		ratio,
		stats.MaxDiff,
		stats.AverageDiff,
		stats.RMSE,
		*outPath,
	)
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func savePNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func analyzeImages(ref, gen image.Image) (diffStats, error) {
	refBounds := ref.Bounds()
	genBounds := gen.Bounds()
	if refBounds.Size() != genBounds.Size() {
		return diffStats{}, fmt.Errorf("dimension mismatch: ref=%v gen=%v", refBounds.Size(), genBounds.Size())
	}

	width := refBounds.Dx()
	height := refBounds.Dy()
	stats := diffStats{
		Width:       width,
		Height:      height,
		TotalPixels: width * height,
	}
	if stats.TotalPixels == 0 {
		return stats, nil
	}

	var totalDiff float64
	var sumSq float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rr, rg, rb := rgb8(ref.At(refBounds.Min.X+x, refBounds.Min.Y+y))
			gr, gg, gb := rgb8(gen.At(genBounds.Min.X+x, genBounds.Min.Y+y))
			dr := absDiff8(rr, gr)
			dg := absDiff8(rg, gg)
			db := absDiff8(rb, gb)
			maxDiff := max8(dr, max8(dg, db))
			if maxDiff != 0 {
				stats.DifferentPixels++
			}
			if maxDiff > stats.MaxDiff {
				stats.MaxDiff = maxDiff
			}
			totalDiff += float64(dr) + float64(dg) + float64(db)
			sumSq += squaredDiff(rr, gr) + squaredDiff(rg, gg) + squaredDiff(rb, gb)
		}
	}

	stats.AverageDiff = totalDiff / float64(stats.TotalPixels*3)
	stats.RMSE = math.Sqrt(sumSq / float64(stats.TotalPixels*3))
	return stats, nil
}

func amplifiedDiff(ref, gen image.Image, factor int) (*image.RGBA, error) {
	if factor <= 0 {
		return nil, errors.New("factor must be positive")
	}
	refBounds := ref.Bounds()
	genBounds := gen.Bounds()
	if refBounds.Size() != genBounds.Size() {
		return nil, fmt.Errorf("dimension mismatch: ref=%v gen=%v", refBounds.Size(), genBounds.Size())
	}

	width := refBounds.Dx()
	height := refBounds.Dy()
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rr, rg, rb := rgb8(ref.At(refBounds.Min.X+x, refBounds.Min.Y+y))
			gr, gg, gb := rgb8(gen.At(genBounds.Min.X+x, genBounds.Min.Y+y))
			dr := absDiff8(rr, gr)
			dg := absDiff8(rg, gg)
			db := absDiff8(rb, gb)
			if dr == 0 && dg == 0 && db == 0 {
				gray := uint8((int(rr) + int(rg) + int(rb)) / 10)
				out.SetRGBA(x, y, color.RGBA{R: gray, G: gray, B: gray, A: 255})
				continue
			}
			out.SetRGBA(x, y, color.RGBA{
				R: amplify(dr, factor),
				G: amplify(dg, factor),
				B: amplify(db, factor),
				A: 255,
			})
		}
	}
	return out, nil
}

func rgb8(c color.Color) (r, g, b uint8) {
	r16, g16, b16, _ := c.RGBA()
	return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8)
}

func absDiff8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func max8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func squaredDiff(a, b uint8) float64 {
	d := float64(int(a) - int(b))
	return d * d
}

func amplify(v uint8, factor int) uint8 {
	n := int(v) * factor
	if n > 255 {
		return 255
	}
	return uint8(n)
}
