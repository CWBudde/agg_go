package main

import (
	"fmt"
	"os"

	"agg_go"
)

func main() {
	// Create a 400x300 context
	ctx := agg.NewContext(400, 300)

	// Clear with white background
	ctx.Clear(agg.White)

	// Draw a red filled ellipse
	ctx.SetColor(agg.Red)
	ctx.FillEllipse(200, 150, 80, 60)

	// Draw a blue ellipse outline
	ctx.SetColor(agg.Blue)
	ctx.DrawEllipse(200, 150, 100, 80)

	// Draw some smaller ellipses
	ctx.SetColor(agg.Green)
	ctx.FillEllipse(100, 100, 30, 20)

	ctx.SetColor(agg.Yellow)
	ctx.FillEllipse(300, 200, 25, 40)

	// Get the image and write to a simple PPM file for verification
	img := ctx.GetImage()

	// Write PPM file
	file, err := os.Create("ellipse_test.ppm")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	// PPM header
	fmt.Fprintf(file, "P6\n%d %d\n255\n", img.Width(), img.Height())

	// Write pixel data (convert RGBA to RGB)
	for i := 0; i < len(img.Data); i += 4 {
		// Write RGB, skip A
		file.Write([]byte{img.Data[i], img.Data[i+1], img.Data[i+2]})
	}

	fmt.Println("Ellipse test image written to ellipse_test.ppm")
	fmt.Println("You can view this with any image viewer that supports PPM format")
}
