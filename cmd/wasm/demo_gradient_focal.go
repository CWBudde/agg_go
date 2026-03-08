package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/gamma"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// Port of AGG C++ gradient_focal.cpp.
//
// Web variant keeps controls outside AGG widgets: parameters are controlled
// via JS/URL query params (`gfg`, `gfx`, `gfy`).
var (
	gradientFocalGamma = 1.0
	gradientFocalFX    = 40.0
	gradientFocalFY    = -10.0
)

func setGradientFocalGamma(v float64) {
	if v < 0.5 {
		v = 0.5
	}
	if v > 2.5 {
		v = 2.5
	}
	gradientFocalGamma = v
}

func setGradientFocalFX(v float64) { gradientFocalFX = v }

func setGradientFocalFY(v float64) { gradientFocalFY = v }

func buildGradientFocalLUT(g float64, size int) []color.RGBA8[color.Linear] {
	if size < 2 {
		size = 2
	}
	lut := make([]color.RGBA8[color.Linear], size)
	gammaLUT := gamma.NewGammaLUT8WithGamma(g)

	type stop struct {
		pos float64
		r   uint8
		g   uint8
		b   uint8
	}
	stops := []stop{
		{pos: 0.0, r: 0, g: 255, b: 0},
		{pos: 0.2, r: 120, g: 0, b: 0},
		{pos: 0.7, r: 120, g: 120, b: 0},
		{pos: 1.0, r: 0, g: 0, b: 255},
	}

	type stopGamma struct {
		pos float64
		r   float64
		g   float64
		b   float64
	}
	sg := make([]stopGamma, len(stops))
	for i, s := range stops {
		sg[i] = stopGamma{
			pos: s.pos,
			r:   float64(gammaLUT.Dir(basics.Int8u(s.r))),
			g:   float64(gammaLUT.Dir(basics.Int8u(s.g))),
			b:   float64(gammaLUT.Dir(basics.Int8u(s.b))),
		}
	}

	for i := 0; i < size; i++ {
		t := float64(i) / float64(size-1)
		j := 0
		for j < len(sg)-2 && t > sg[j+1].pos {
			j++
		}
		a := sg[j]
		b := sg[j+1]
		den := b.pos - a.pos
		u := 0.0
		if den > 0 {
			u = (t - a.pos) / den
		}
		if u < 0 {
			u = 0
		}
		if u > 1 {
			u = 1
		}
		r := uint8(a.r + (b.r-a.r)*u + 0.5)
		gv := uint8(a.g + (b.g-a.g)*u + 0.5)
		bv := uint8(a.b + (b.b-a.b)*u + 0.5)
		lut[i] = color.RGBA8[color.Linear]{R: r, G: gv, B: bv, A: 255}
	}

	return lut
}

func applyGammaInvToContextImage(g float64) {
	if math.Abs(g-1.0) < 1e-9 {
		return
	}
	lut := gamma.NewGammaLUT8WithGamma(g)
	img := ctx.GetImage()
	for i := 0; i+3 < len(img.Data); i += 4 {
		img.Data[i+0] = uint8(lut.Inv(basics.Int8u(img.Data[i+0])))
		img.Data[i+1] = uint8(lut.Inv(basics.Int8u(img.Data[i+1])))
		img.Data[i+2] = uint8(lut.Inv(basics.Int8u(img.Data[i+2])))
	}
}

func drawGradientFocalDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	cx := float64(width) * 0.5
	cy := float64(height) * 0.5
	r := 100.0

	// Match C++ trans_affine_resizing() behavior from 600x400 base window.
	sx := float64(width) / 600.0
	sy := float64(height) / 400.0

	gradientMtx := transform.NewTransAffine()
	gradientMtx.Translate(cx, cy)
	gradientMtx.Multiply(transform.NewTransAffineScalingXY(sx, sy))
	gradientMtx.Invert()

	interpolator := span.NewSpanInterpolatorLinearDefault(gradientMtx)
	gradientFunc := span.NewGradientRadialFocus(r, gradientFocalFX, gradientFocalFY)
	gradientReflect := span.NewGradientReflectAdaptor(gradientFunc)

	lut := buildGradientFocalLUT(gradientFocalGamma, 1024)
	colorFn := span.NewGradientPrebuiltColorRGBA8[color.Linear](lut)
	spanGen := span.NewSpanGradient(interpolator, gradientReflect, colorFn, 0, r)

	ras := a.GetInternalRasterizer()
	ras.Reset()
	ras.MoveToD(0, 0)
	ras.LineToD(float64(width), 0)
	ras.LineToD(float64(width), float64(height))
	ras.LineToD(0, float64(height))
	ras.LineToD(0, 0)
	a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	ctx.SetColor(agg.White)
	ctx.SetLineWidth(1.0)
	ctx.DrawEllipse(cx, cy, r*sx, r*sy)

	// C++ applies inverse gamma to the framebuffer after rasterization.
	applyGammaInvToContextImage(gradientFocalGamma)
}
