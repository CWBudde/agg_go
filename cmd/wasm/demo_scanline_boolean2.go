package main

import "agg_go/internal/demo/scanlineboolean2"

// Port of AGG C++ scanline_boolean2.cpp (web variant).
var (
	sb2Mode    = 3
	sb2Fill    = 1
	sb2Scan    = 1
	sb2Op      = 2
	sb2CenterX = 0.0
	sb2CenterY = 0.0
)

func setScanlineBoolean2Mode(v int)      { sb2Mode = v }
func setScanlineBoolean2FillRule(v int)  { sb2Fill = v }
func setScanlineBoolean2Scanline(v int)  { sb2Scan = v }
func setScanlineBoolean2Operation(v int) { sb2Op = v }
func setScanlineBoolean2Center(x, y float64) {
	sb2CenterX = x
	sb2CenterY = y
}

func drawScanlineBoolean2Demo() {
	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:         sb2Mode,
		FillRule:     sb2Fill,
		ScanlineType: sb2Scan,
		Operation:    sb2Op,
		CenterX:      sb2CenterX,
		CenterY:      sb2CenterY,
	})
}
