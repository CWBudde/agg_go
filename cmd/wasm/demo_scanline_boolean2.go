package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/demo/scanlineboolean2"
)

// Port of AGG C++ scanline_boolean2.cpp (web variant).
var (
	sb2Mode    = 3
	sb2Fill    = 1
	sb2Scan    = 1
	sb2Op      = 2
	sb2CenterX = math.NaN()
	sb2CenterY = math.NaN()
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
	setScanlineBoolean2Center(x, y)
	return true
}

func handleScanlineBoolean2MouseMove(x, y float64) bool {
	setScanlineBoolean2Center(x, y)
	return true
}

func handleScanlineBoolean2MouseUp() {}

func drawScanlineBoolean2Demo() {
	cx, cy := sb2CenterX, sb2CenterY
	if math.IsNaN(cx) || math.IsNaN(cy) {
		cx = float64(width) * 0.5
		cy = float64(height) * 0.5
	}
	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:         sb2Mode,
		FillRule:     sb2Fill,
		ScanlineType: sb2Scan,
		Operation:    sb2Op,
		CenterX:      cx,
		CenterY:      cy,
	})
}
