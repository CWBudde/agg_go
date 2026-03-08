package main

import (
	"path/filepath"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Checkerboard background like the C++ example.
	ctx.Clear(agg.RGB(1.0, 1.0, 1.0))
	for y := 0; y < 400; y += 8 {
		xStart := 0
		if ((y >> 3) & 1) == 1 {
			xStart = 8
		}
		for x := xStart; x < 600; x += 16 {
			a.FillColor(agg.NewColor(223, 223, 223, 255))
			a.NoLine()
			a.ResetPath()
			a.MoveTo(float64(x), float64(y))
			a.LineTo(float64(x+7), float64(y))
			a.LineTo(float64(x+7), float64(y+7))
			a.LineTo(float64(x), float64(y+7))
			a.ClosePolygon()
			a.DrawPath(agg.FillOnly)
		}
	}

	// Destination image (BMP variant from shared art assets).
	dstImg, err := agg.LoadImageFromFile(filepath.Join("examples", "shared", "art", "compositing.bmp"))
	if err == nil {
		// Same placement as C++: blend_from loaded image at (250, 180).
		_ = a.BlendImageSimple(dstImg, 250, 180, 255)
	}

	// Destination radial circle.
	a.BlendMode(agg.BlendAlpha)
	a.FillRadialGradient(
		(70*3+37*3)/2.0,
		(100+24*3+100+79*3)/2.0,
		70,
		agg.NewColor(0xFD, 0xF0, 0x6F, 255),
		agg.NewColor(0xFE, 0x9F, 0x34, 255),
		1.0,
	)
	a.NoLine()
	a.FillCircle((70*3+37*3)/2.0, (100+24*3+100+79*3)/2.0, 63)

	// Source shape composited with default operator from C++ (src-over).
	a.BlendMode(agg.BlendSrcOver)
	a.FillLinearGradient(
		350, 172, 157, 337,
		agg.NewColor(0x7F, 0xC1, 0xFF, 191),
		agg.NewColor(0x05, 0x00, 0x5F, 191),
		1.0,
	)
	a.NoLine()
	a.RoundedRect(350, 172, 157, 337, 40)
	a.DrawPath(agg.FillOnly)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Compositing", Width: 600, Height: 400}, &demo{})
}
