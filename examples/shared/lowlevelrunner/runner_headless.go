//go:build !x11 && !sdl2

package lowlevelrunner

import (
	"fmt"
	"image/png"
	"os"
	"strings"

	agg "github.com/MeKo-Christian/agg_go"
)

// Run renders the demo once and saves the result as a PNG file.
// The filename is derived from Config.Title (spaces -> underscores, + ".png").
func Run(cfg Config, demo Demo) {
	stride := cfg.Width * 4
	if cfg.FlipY {
		stride = -stride
	}
	img := agg.NewImage(make([]uint8, cfg.Width*cfg.Height*4), cfg.Width, cfg.Height, stride)
	if initDemo, ok := demo.(InitHandler); ok {
		initDemo.OnInit()
	}
	demo.Render(img)

	filename := strings.ReplaceAll(strings.ToLower(cfg.Title), " ", "_") + ".png"
	if err := savePNG(img, filename); err != nil {
		fmt.Fprintf(os.Stderr, "lowlevelrunner: save PNG: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("saved %s\n", filename)
}

func savePNG(img *agg.Image, filename string) error {
	goImg := img.ToGoImage()
	if goImg == nil {
		return fmt.Errorf("image conversion failed")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, goImg)
}
