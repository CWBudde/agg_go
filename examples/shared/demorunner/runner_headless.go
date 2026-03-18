//go:build !x11 && !sdl2

package demorunner

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	agg "github.com/MeKo-Christian/agg_go"
)

// Run renders the demo once and saves the result as a PNG file.
// The filename is derived from Config.Title (spaces → underscores, + ".png").
func Run(cfg Config, demo Demo) {
	ctx := agg.NewContext(cfg.Width, cfg.Height)
	demo.Render(ctx)

	filename := strings.ReplaceAll(strings.ToLower(cfg.Title), " ", "_") + ".png"
	if err := savePNG(ctx, filename); err != nil {
		fmt.Fprintf(os.Stderr, "demorunner: save PNG: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("saved %s\n", filename)
}

func savePNG(ctx *agg.Context, filename string) error {
	img := ctx.GetImage()
	src := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))
	copy(src.Pix, img.Data)
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))
	rowBytes := img.Width() * 4
	for y := 0; y < img.Height(); y++ {
		srcOff := (img.Height() - 1 - y) * src.Stride
		dstOff := y * goImg.Stride
		copy(goImg.Pix[dstOff:dstOff+rowBytes], src.Pix[srcOff:srcOff+rowBytes])
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, goImg)
}
