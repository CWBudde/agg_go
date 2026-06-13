// Package imgdiff provides shared RGB image comparison statistics and amplified
// diff rendering. It is used by both cmd/visual-diff and cmd/engine-compare so
// the two tools report identical RMSE/max-diff numbers and produce identical
// amplified diff images.
package imgdiff

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
)

// Stats summarizes the per-channel RGB differences between two images.
type Stats struct {
	Width           int
	Height          int
	TotalPixels     int
	DifferentPixels int
	MaxDiff         uint8
	AverageDiff     float64
	RMSE            float64
}

// Ratio returns DifferentPixels / TotalPixels (0 when there are no pixels).
func (s Stats) Ratio() float64 {
	if s.TotalPixels == 0 {
		return 0
	}
	return float64(s.DifferentPixels) / float64(s.TotalPixels)
}

// Analyze compares two images over their RGB channels and returns difference
// statistics. The images must have identical dimensions.
func Analyze(ref, gen image.Image) (Stats, error) {
	refBounds := ref.Bounds()
	genBounds := gen.Bounds()
	if refBounds.Size() != genBounds.Size() {
		return Stats{}, fmt.Errorf("dimension mismatch: ref=%v gen=%v", refBounds.Size(), genBounds.Size())
	}

	width := refBounds.Dx()
	height := refBounds.Dy()
	stats := Stats{
		Width:       width,
		Height:      height,
		TotalPixels: width * height,
	}
	if stats.TotalPixels == 0 {
		return stats, nil
	}

	var totalDiff float64
	var sumSq float64
	for y := range height {
		for x := range width {
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

// Amplify renders an amplified diff image: identical pixels are darkened to
// grayscale, differing pixels are scaled by factor so small deltas become
// visible. The images must have identical dimensions and factor must be > 0.
func Amplify(ref, gen image.Image, factor int) (*image.RGBA, error) {
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
	for y := range height {
		for x := range width {
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
