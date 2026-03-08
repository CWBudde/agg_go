//go:build !x11 && !sdl2

package demorunner

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	agg "agg_go"
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
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))
	copy(goImg.Pix, img.Data)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, goImg)
}
