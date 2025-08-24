// Basic line drawing demo using the high-level Context API.
// Renders a few lines with different orientations and saves a PNG.
package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	agg "agg_go"
)

func main() {
	const width, height = 640, 480
	fmt.Println("AGG Go - Lines Demo")
	fmt.Printf("Creating %dx%d canvas and drawing lines...\n", width, height)

	ctx := agg.NewContext(width, height)
	ctx.Clear(agg.White)

	// Axes
	ctx.SetColor(agg.RGB(0.9, 0.9, 0.95))
	for y := 0; y < height; y += 40 {
		ctx.DrawLine(0, float64(y), float64(width-1), float64(y))
	}
	for x := 0; x < width; x += 40 {
		ctx.DrawLine(float64(x), 0, float64(x), float64(height-1))
	}

	// Main diagonals
	ctx.SetColor(agg.Blue)
	ctx.DrawLine(0, 0, float64(width-1), float64(height-1))
	ctx.SetColor(agg.Red)
	ctx.DrawLine(float64(width-1), 0, 0, float64(height-1))

	// Starburst from center
	cx, cy := float64(width/2), float64(height/2)
	ctx.SetColor(agg.Green)
	for i := 0; i < 12; i++ {
		angle := float64(i) * (math.Pi / 6.0) // every 30 degrees
		x := cx + 220.0*math.Cos(angle)
		y := cy + 220.0*math.Sin(angle)
		ctx.DrawLine(cx, cy, x, y)
	}

	// Thick lines showcase (different widths and colors)
	ctx.SetColor(agg.RGB(0.2, 0.2, 0.2))
	ctx.DrawThickLine(60, 420, 260, 420, 1) // width 1
	ctx.SetColor(agg.RGB(0.0, 0.4, 0.9))
	ctx.DrawThickLine(60, 390, 260, 390, 4) // width 4
	ctx.SetColor(agg.RGB(0.9, 0.3, 0.1))
	ctx.DrawThickLine(60, 360, 260, 360, 8) // width 8
	ctx.SetColor(agg.RGB(0.4, 0.7, 0.2))
	ctx.DrawThickLine(320, 360, 540, 420, 10) // slanted thick line

	// Save PNG
	out := "lines_demo.png"
	if err := saveAsPNG(ctx.GetImage(), out); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}
	fmt.Printf("Lines demo saved to: %s\n", out)
}

// saveAsPNG converts the AGG image to PNG format and writes it to filename.
func saveAsPNG(img *agg.Image, filename string) error {
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	copy(goImg.Pix, img.Data)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, goImg)
}
