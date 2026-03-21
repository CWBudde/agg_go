package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	f, err := os.Open("mask_debug.png")
	if err != nil {
		fmt.Println("Run alpha_mask2 with mask dump first")
		// Try to read existing output instead
		f2, _ := os.Open("alpha_mask2.png")
		if f2 != nil {
			img, _ := png.Decode(f2)
			f2.Close()
			for _, pt := range [][2]int{{300, 100}, {250, 150}, {350, 80}} {
				x, y := pt[0], pt[1]
				c := img.At(x, y).(color.RGBA)
				fmt.Printf("output(%d,%d): R=%d G=%d B=%d\n", x, y, c.R, c.G, c.B)
			}
		}
		return
	}
	defer f.Close()
	img, _ := png.Decode(f)
	_ = image.Point{}

	fmt.Println("Mask values:")
	for _, pt := range [][2]int{{300, 100}, {250, 150}, {350, 80}} {
		x, y := pt[0], pt[1]
		c := img.At(x, y).(color.Gray)
		fmt.Printf("mask(%d,%d) = %d\n", x, y, c.Y)
	}
}
