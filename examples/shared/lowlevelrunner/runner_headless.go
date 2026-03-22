//go:build !x11 && !sdl2

package lowlevelrunner

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	agg "github.com/MeKo-Christian/agg_go"
)

// Run renders the demo once and saves the result as a PNG file.
// The filename is derived from Config.Title (spaces -> underscores, + ".png").
func Run(cfg Config, demo Demo) {
	// Match C++ platform_support: when flip_y=true the rendering buffer is
	// attached with a negative stride so that row 0 is at the physical bottom
	// of the buffer.  The PNG is then written top-to-bottom from the raw bytes
	// without any additional row-reversal — exactly as C++ would blit.
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
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))
	srcStride := img.Width() * 4
	for y := range img.Height() {
		srcOff := y * srcStride
		dstOff := y * goImg.Stride
		for x := range img.Width() {
			srcIdx := srcOff + x*4
			dstIdx := dstOff + x*4
			goImg.Pix[dstIdx] = img.Data[srcIdx]
			goImg.Pix[dstIdx+1] = img.Data[srcIdx+1]
			goImg.Pix[dstIdx+2] = img.Data[srcIdx+2]
			goImg.Pix[dstIdx+3] = 255
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, goImg)
}
