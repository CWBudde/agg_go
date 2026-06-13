// Based on the original AGG examples: trans_curve2.cpp.
//
// The rendering core is shared with the standalone example
// (examples/core/intermediate/trans_curve2) via internal/demo/transcurve
// (DrawDouble), so both surfaces draw the identical scene. This file keeps only
// the interactive state, animation, and pointer handling specific to the web UI.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/demo/transcurve"
)

var (
	transCurve2Points1  = transcurve.DefaultPoints1
	transCurve2Points2  = transcurve.DefaultPoints2
	transCurve2Selected = -1
	transCurve2Animate  = false
	transCurve2DX1      [6]float64
	transCurve2DY1      [6]float64
	transCurve2DX2      [6]float64
	transCurve2DY2      [6]float64

	transCurve2NumPoints      = 200.0
	transCurve2FixedLen       = true
	transCurve2PreserveXScale = true
)

const (
	transCurve2RefW = 600.0
	transCurve2RefH = 600.0
)

func transCurve2FrameOffset() (float64, float64) {
	return (float64(width) - transCurve2RefW) * 0.5, (float64(height) - transCurve2RefH) * 0.5
}

func initTransCurve2Demo() {
	for i := 0; i < 6; i++ {
		transCurve2DX1[i] = (math.Mod(float64(i*1234+1), 10.0) - 5.0) * 0.5
		transCurve2DY1[i] = (math.Mod(float64(i*5678+2), 10.0) - 5.0) * 0.5
		transCurve2DX2[i] = (math.Mod(float64(i*1234+3), 10.0) - 5.0) * 0.5
		transCurve2DY2[i] = (math.Mod(float64(i*5678+4), 10.0) - 5.0) * 0.5
	}
}

func drawTransCurve2Demo() {
	initTransCurve2Demo()
	offX, offY := transCurve2FrameOffset()

	if transCurve2Animate {
		for i := 0; i < 6; i++ {
			moveTransCurve2Point(&transCurve2Points1[i*2], &transCurve2Points1[i*2+1], &transCurve2DX1[i], &transCurve2DY1[i])
			moveTransCurve2Point(&transCurve2Points2[i*2], &transCurve2Points2[i*2+1], &transCurve2DX2[i], &transCurve2DY2[i])
			d := math.Sqrt((transCurve2Points1[i*2]-transCurve2Points2[i*2])*(transCurve2Points1[i*2]-transCurve2Points2[i*2]) + (transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1])*(transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1]))
			if d > 28.28 {
				transCurve2Points2[i*2] = transCurve2Points1[i*2] + (transCurve2Points2[i*2]-transCurve2Points1[i*2])*28.28/d
				transCurve2Points2[i*2+1] = transCurve2Points1[i*2+1] + (transCurve2Points2[i*2+1]-transCurve2Points1[i*2+1])*28.28/d
			}
		}
	}

	transcurve.DrawDouble(ctx, transcurve.DoubleConfig{
		Points1:         transCurve2Points1,
		Points2:         transCurve2Points2,
		NumIntermediate: transCurve2NumPoints,
		PreserveXScale:  transCurve2PreserveXScale,
		FixedLength:     transCurve2FixedLen,
		BaseLength:      transcurve.DefaultDoubleBaseLength,
		BaseHeight:      transcurve.DefaultDoubleBaseHeight,
		Text:            transcurve.DefaultDoubleText,
		OffsetX:         offX,
		OffsetY:         offY,
	})
}

func moveTransCurve2Point(x, y, dx, dy *float64) {
	*x += *dx
	*y += *dy
	if *x < 0 || *x > transCurve2RefW {
		*dx = -*dx
	}
	if *y < 0 || *y > transCurve2RefH {
		*dy = -*dy
	}
}

func handleTransCurve2MouseDown(x, y float64) bool {
	offX, offY := transCurve2FrameOffset()
	x -= offX
	y -= offY
	transCurve2Selected = -1
	for i := 0; i < 6; i++ {
		if math.Sqrt((x-transCurve2Points1[i*2])*(x-transCurve2Points1[i*2])+(y-transCurve2Points1[i*2+1])*(y-transCurve2Points1[i*2+1])) < 15 {
			transCurve2Selected = i
			return true
		}
		if math.Sqrt((x-transCurve2Points2[i*2])*(x-transCurve2Points2[i*2])+(y-transCurve2Points2[i*2+1])*(y-transCurve2Points2[i*2+1])) < 15 {
			transCurve2Selected = i + 6
			return true
		}
	}
	return false
}

func handleTransCurve2MouseMove(x, y float64) bool {
	offX, offY := transCurve2FrameOffset()
	x -= offX
	y -= offY
	if transCurve2Selected != -1 {
		if transCurve2Selected < 6 {
			transCurve2Points1[transCurve2Selected*2] = x
			transCurve2Points1[transCurve2Selected*2+1] = y
		} else {
			idx := transCurve2Selected - 6
			transCurve2Points2[idx*2] = x
			transCurve2Points2[idx*2+1] = y
		}
		return true
	}
	return false
}

func handleTransCurve2MouseUp() {
	transCurve2Selected = -1
}

func toggleTransCurve2Animate() {
	transCurve2Animate = !transCurve2Animate
}

func setTransCurve2NumPoints(v float64) {
	transCurve2NumPoints = v
}

func setTransCurve2FixedLen(v bool) {
	transCurve2FixedLen = v
}

func setTransCurve2PreserveXScale(v bool) {
	transCurve2PreserveXScale = v
}
