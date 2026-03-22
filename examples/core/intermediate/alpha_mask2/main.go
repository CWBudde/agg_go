// Port of AGG C++ alpha_mask2.cpp – alpha-masked lion with affine-transformed mask.
//
// Generates an alpha mask from random ellipses (with affine transform), then
// renders the lion through it. The C++ original also renders random lines,
// markers, and gradient circles through the mask — those are not yet ported.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	alphamask2demo "github.com/MeKo-Christian/agg_go/internal/demo/alphamask2"
)

const (
	frameWidth  = 512
	frameHeight = 400

	numEllipses = 10 // C++ default slider value
)

type demo struct {
	angle, scale float64
	skewX, skewY float64
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*3)
	alphamask2demo.RenderToBGR24(workBuf, w, h, alphamask2demo.Config{
		NumEllipses: numEllipses,
		Angle:       d.angle,
		Scale:       d.scale,
		SkewX:       d.skewX,
		SkewY:       d.skewY,
	})

	// Convert BGR24 work buffer to RGBA output with y-flip.
	copyBGR24FlipY(workBuf, img.Data, w, h)
}

func copyBGR24FlipY(src, dst []uint8, width, height int) {
	srcStride := width * 3
	dstStride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * srcStride
		dstOff := y * dstStride
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			dst[d+0] = src[s+2]
			dst[d+1] = src[s+1]
			dst[d+2] = src[s+0]
			dst[d+3] = 255
		}
	}
}

func main() {
	d := &demo{scale: 1.0}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Alpha Mask2",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
