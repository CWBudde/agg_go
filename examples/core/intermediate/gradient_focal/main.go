// Package main ports AGG's gradient_focal.cpp demo.
//
// It renders a radial-focus gradient with a reflected profile and a gamma-aware
// 4-stop LUT, then applies inverse gamma to the final framebuffer to match the
// original AGG rendering path.
package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/gamma"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

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

func applyGammaInv(img *agg.Image, g float64) {
	if math.Abs(g-1.0) < 1e-9 {
		return
	}
	lut := gamma.NewGammaLUT8WithGamma(g)
	for i := 0; i+3 < len(img.Data); i += 4 {
		img.Data[i+0] = uint8(lut.Inv(basics.Int8u(img.Data[i+0])))
		img.Data[i+1] = uint8(lut.Inv(basics.Int8u(img.Data[i+1])))
		img.Data[i+2] = uint8(lut.Inv(basics.Int8u(img.Data[i+2])))
	}
}

func main() {
	const (
		w     = 600
		h     = 400
		r     = 100.0
		fx    = 40.0
		fy    = -10.0
		gamma = 1.0
		out   = "gradient_focal.png"
	)

	ctx := agg.NewContext(w, h)
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	cx := float64(w) * 0.5
	cy := float64(h) * 0.5

	gradientMtx := transform.NewTransAffine()
	gradientMtx.Translate(cx, cy)
	gradientMtx.Invert()

	interpolator := span.NewSpanInterpolatorLinearDefault(gradientMtx)
	gradientFunc := span.NewGradientRadialFocus(r, fx, fy)
	gradientReflect := span.NewGradientReflectAdaptor(gradientFunc)
	colorFn := span.NewGradientPrebuiltColorRGBA8[color.Linear](buildGradientFocalLUT(gamma, 1024))
	spanGen := span.NewSpanGradient(interpolator, gradientReflect, colorFn, 0, r)

	ras := a.GetInternalRasterizer()
	ras.Reset()
	ras.MoveToD(0, 0)
	ras.LineToD(w, 0)
	ras.LineToD(w, h)
	ras.LineToD(0, h)
	ras.LineToD(0, 0)
	a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	ctx.SetColor(agg.White)
	ctx.SetLineWidth(1.0)
	ctx.DrawCircle(cx, cy, r)

	applyGammaInv(ctx.GetImage(), gamma)

	if err := ctx.GetImage().SaveToPNG(out); err != nil {
		fmt.Printf("error writing %s: %v\n", out, err)
		return
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, w, h)
}
