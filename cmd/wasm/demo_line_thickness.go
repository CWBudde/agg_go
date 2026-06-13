package main

import "github.com/cwbudde/agg_go/internal/demo/linethickness"

// Port of AGG C++ line_thickness.cpp.
//
// Web variant keeps controls outside AGG widgets: parameters are controlled
// via JS/URL query params.
var (
	lineThicknessState = linethickness.DefaultState()
)

func setLineThicknessFactor(v float64) {
	lineThicknessState.Thickness = v
	lineThicknessState.Clamp()
}

func setLineThicknessBlur(v float64) {
	lineThicknessState.Blur = v
	lineThicknessState.Clamp()
}

func setLineThicknessMono(v bool) { lineThicknessState.Mono = v }

func setLineThicknessInvert(v bool) { lineThicknessState.Invert = v }

func getLineThicknessBlurTime() float64 { return linethickness.LastBlurMS() }

func drawLineThicknessDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	// The scene uses fixed C++ coordinates designed for a 640x480 frame (the
	// row of lines spans x=20..610, the wheel is centred at 320,180). Render it
	// into a 640x480 work buffer exactly like the standalone example, then
	// composite that frame centred into the (potentially larger) canvas so the
	// scene stays centred instead of anchored to a corner.
	const fw, fh = linethickness.Width, linethickness.Height
	workBuf := make([]uint8, fw*fh*4)
	linethickness.Draw(workBuf, fw, fh, lineThicknessState)

	// Fill the whole canvas with the scene background so the margins around the
	// centred frame match the frame's own background colour.
	_, bg := linethickness.Colors(lineThicknessState)
	for i := 0; i+3 < len(img.Data); i += 4 {
		img.Data[i], img.Data[i+1], img.Data[i+2], img.Data[i+3] = bg.R, bg.G, bg.B, bg.A
	}

	// Centre offsets of the frame within the canvas (clamped to >= 0 so a
	// smaller canvas crops symmetrically rather than reading out of bounds).
	ox := (w - fw) / 2
	oy := (h - fh) / 2

	// Copy the work buffer into the canvas with a y-flip (row 0 of workBuf is
	// the bottom of the logical frame) and the centring offset applied.
	srcStride := fw * 4
	dstStride := w * 4
	for sy := range fh {
		dy := oy + (fh - 1 - sy) // flipped destination row
		if dy < 0 || dy >= h {
			continue
		}
		for sx := range fw {
			dx := ox + sx
			if dx < 0 || dx >= w {
				continue
			}
			s := sy*srcStride + sx*4
			d := dy*dstStride + dx*4
			img.Data[d], img.Data[d+1], img.Data[d+2], img.Data[d+3] =
				workBuf[s], workBuf[s+1], workBuf[s+2], workBuf[s+3]
		}
	}

	applyLinearToSRGB(img)
}
