// Package main writes amplified visual diffs for PNG reference comparisons.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/cwbudde/agg_go/internal/imgdiff"
)

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

	stats, err := imgdiff.Analyze(ref, gen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "analyze images: %v\n", err)
		os.Exit(1)
	}
	diff, err := imgdiff.Amplify(ref, gen, *factor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build diff: %v\n", err)
		os.Exit(1)
	}
	if err := savePNG(diff, *outPath); err != nil {
		fmt.Fprintf(os.Stderr, "save diff: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"size=%dx%d different_pixels=%d/%d ratio=%.6f max_diff=%d avg_diff=%.4f rmse=%.4f out=%s\n",
		stats.Width,
		stats.Height,
		stats.DifferentPixels,
		stats.TotalPixels,
		stats.Ratio(),
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
