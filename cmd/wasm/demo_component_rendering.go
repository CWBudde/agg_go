// Port of AGG C++ component_rendering.cpp – component (channel) rendering.
//
// Three large circles are each rendered into an individual color channel
// (Red, Green, Blue) by using per-channel grayscale rendering into a
// temporary gray8 buffer and then applying the result as a channel darkening
// on the main RGBA buffer. The effect shows subtractive CMY mixing:
//
//	Red ∩ Green  → Blue   (Cyan × Magenta)
//	Red ∩ Blue   → Green  (Cyan × Yellow)
//	Green ∩ Blue → Red    (Magenta × Yellow)
//	All three    → Black
//
// An alpha slider controls how strongly each channel is darkened.
package main

import (
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// --- State ---

var compAlpha = 255 // 0..255

// --- Drawing ---

// renderComponentEllipseToChannel renders an anti-aliased ellipse into a
// temporary gray8 buffer and applies the result as a per-channel darkening
// to the main RGBA buffer. This emulates the C++ pixfmt_alpha_blend_gray
// per-channel technique.
//
// grayVal=0 with alpha means: dst_channel = dst_channel * g / 255
// where g is the gray value produced by blending gray(0, alpha) onto white.
func renderComponentEllipseToChannel(
	mainBuf []byte, w, h int,
	cx, cy, rx, ry float64,
	alpha uint8,
	channelOffset int, // 0=R, 1=G, 2=B in RGBA layout
) {
	// Render the ellipse into a temporary gray8 buffer to get coverage.
	grayBuf := make([]byte, w*h)
	grayRbuf := buffer.NewRenderingBufferU8WithData(grayBuf, w, h, w)
	grayPixf := pixfmt.NewPixFmtGray8(grayRbuf)
	grayRb := renderer.NewRendererBaseWithPixfmt(grayPixf)
	grayRb.Clear(color.Gray8[color.Linear]{V: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
	ras.AddPath(&ellipseVS{ell: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, grayRb, color.Gray8[color.Linear]{V: 0, A: alpha})

	// Apply the gray coverage to the target channel:
	// the gray buffer started at 255; after rendering gray(0,alpha) it contains
	// a value g < 255 inside the ellipse. Multiply the channel by g/255.
	stride := w * 4
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g := grayBuf[y*w+x]
			if g == 255 {
				continue
			}
			idx := y*stride + x*4 + channelOffset
			mainBuf[idx] = uint8(uint16(mainBuf[idx]) * uint16(g) / 255)
		}
	}
}

func drawComponentRenderingDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	// Work in a separate RGBA32 buffer (linear), then copy to img.
	workBuf := make([]byte, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	fw := float64(w)
	fh := float64(h)
	alpha := uint8(compAlpha)

	// Native demo size is 320×320; canvas is 800×600.
	// Scale offsets and radii proportionally to the canvas.
	const nativeSize = 320.0
	const nativeOffset = 50.0
	const nativeRadius = 100.0

	scaleX := fw / nativeSize
	scaleY := fh / nativeSize
	cx := fw / 2
	cy := fh / 2
	offsetX := nativeOffset * scaleX
	offsetY := nativeOffset * scaleY
	rx := nativeRadius * scaleX
	ry := nativeRadius * scaleY

	// Red channel: ellipse at (cx - 0.87*offset, cy - 0.5*offset).
	renderComponentEllipseToChannel(workBuf, w, h,
		cx-0.87*offsetX, cy-0.5*offsetY, rx, ry,
		alpha, 0)

	// Green channel: ellipse at (cx + 0.87*offset, cy - 0.5*offset).
	renderComponentEllipseToChannel(workBuf, w, h,
		cx+0.87*offsetX, cy-0.5*offsetY, rx, ry,
		alpha, 1)

	// Blue channel: ellipse at (cx, cy + offset).
	renderComponentEllipseToChannel(workBuf, w, h,
		cx, cy+offsetY, rx, ry,
		alpha, 2)

	// Copy work buffer straight into the image (no y-flip for web canvas).
	copy(img.Data, workBuf)

	// Gamma-encode linear→sRGB.
	applyLinearToSRGB(img)
}
