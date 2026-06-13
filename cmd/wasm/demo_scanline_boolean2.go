package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/demo/scanlineboolean2"
)

// Port of AGG C++ scanline_boolean2.cpp (web variant).
var (
	sb2Mode    = 3
	sb2Fill    = 1
	sb2Scan    = 1
	sb2Op      = 2
	sb2CenterX = math.NaN()
	sb2CenterY = math.NaN()

	// sb2FlipImg is a dedicated bottom-up (negative-stride) buffer. The shared
	// scanlineboolean2.Draw is authored for the standalone example, which runs
	// through lowlevelrunner with FlipY=true: it keeps all geometry in the
	// original C++ Y-up reference frame and relies on a negative-stride buffer to
	// perform the Y-flip at the memory level (scene shapes, the boolean overlay's
	// blendPixel, and the gsv text are all keyed off the stride sign). The web
	// canvas is top-down (positive stride), so rendering the shared Draw directly
	// into the shared ctx leaves everything except the text upside down. Render
	// into this matching bottom-up buffer instead.
	sb2FlipImg *agg.Image
	sb2FlipCtx *agg.Context
)

func setScanlineBoolean2Mode(v int)      { sb2Mode = v }
func setScanlineBoolean2FillRule(v int)  { sb2Fill = v }
func setScanlineBoolean2Scanline(v int)  { sb2Scan = v }
func setScanlineBoolean2Operation(v int) { sb2Op = v }
func setScanlineBoolean2Center(x, y float64) {
	sb2CenterX = x
	sb2CenterY = y
}

func handleScanlineBoolean2MouseDown(x, y float64) bool {
	setScanlineBoolean2Center(x, flipMouseY(y))
	return true
}

func handleScanlineBoolean2MouseMove(x, y float64) bool {
	setScanlineBoolean2Center(x, flipMouseY(y))
	return true
}

// flipMouseY converts a top-down canvas Y coordinate into the Y-up reference
// frame the demo geometry uses, mirroring lowlevelrunner.flipMouseY for the
// FlipY example so the spiral / shape-B centre follows the cursor.
func flipMouseY(y float64) float64 {
	return float64(height-1) - y
}

func handleScanlineBoolean2MouseUp() {}

func drawScanlineBoolean2Demo() {
	cx, cy := sb2CenterX, sb2CenterY
	if math.IsNaN(cx) || math.IsNaN(cy) {
		cx = float64(width) * 0.5
		cy = float64(height) * 0.5
	}

	if sb2FlipImg == nil {
		// Negative stride => bottom-up buffer, matching the FlipY example.
		sb2FlipImg = agg.NewImage(make([]uint8, width*height*4), width, height, -width*4)
		sb2FlipCtx = agg.NewContextForImage(sb2FlipImg)
	}

	scanlineboolean2.Draw(sb2FlipCtx, scanlineboolean2.Config{
		Mode:         sb2Mode,
		FillRule:     sb2Fill,
		ScanlineType: sb2Scan,
		Operation:    sb2Op,
		CenterX:      cx,
		CenterY:      cy,
	})

	// A bottom-up buffer stores its rows in top-down display order (row 0 of the
	// backing slice is the visual top), so the raw bytes can be copied straight
	// into the top-down canvas buffer without a per-row reversal.
	copy(canvasBuf, sb2FlipImg.Data)
}
