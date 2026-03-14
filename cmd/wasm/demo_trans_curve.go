// Go-idiomatic equivalent of AGG's trans_curve1.cpp using the embedded GSV font.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/demo/transcurve"
)

var (
	transCurvePoints         = transcurve.DefaultPoints
	transCurveSelected       = -1
	transCurveAnimate        = false
	transCurveClose          = false
	transCurvePreserveXScale = true
	transCurveFixedLen       = true
	transCurveNumPoints      = 200.0
	transCurveAnimation      = transcurve.NewAnimationState()
)

const (
	transCurveRefW = 600.0
	transCurveRefH = 600.0
)

func transCurveFrameOffset() (float64, float64) {
	return (float64(width) - transCurveRefW) * 0.5, (float64(height) - transCurveRefH) * 0.5
}

func drawTransCurveDemo() {
	if transCurveAnimate {
		transcurve.AnimatePoints(&transCurvePoints, &transCurveAnimation, transCurveRefW, transCurveRefH)
	}

	offX, offY := transCurveFrameOffset()
	transcurve.Draw(ctx, transcurve.Config{
		Points:          transCurvePoints,
		NumIntermediate: transCurveNumPoints,
		Close:           transCurveClose,
		PreserveXScale:  transCurvePreserveXScale,
		FixedLength:     transCurveFixedLen,
		BaseLength:      transcurve.DefaultBaseLength,
		Text:            transcurve.DefaultText,
		OffsetX:         offX,
		OffsetY:         offY,
	})
}

func handleTransCurveMouseDown(x, y float64) bool {
	offX, offY := transCurveFrameOffset()
	x -= offX
	y -= offY
	transCurveSelected = -1
	for i := 0; i < transcurve.ControlPointCount; i++ {
		dx := x - transCurvePoints[i*2]
		dy := y - transCurvePoints[i*2+1]
		if math.Hypot(dx, dy) < 15 {
			transCurveSelected = i
			return true
		}
	}
	return false
}

func handleTransCurveMouseMove(x, y float64) bool {
	offX, offY := transCurveFrameOffset()
	x -= offX
	y -= offY
	if transCurveSelected == -1 {
		return false
	}
	transCurvePoints[transCurveSelected*2] = x
	transCurvePoints[transCurveSelected*2+1] = y
	return true
}

func handleTransCurveMouseUp() {
	transCurveSelected = -1
}

func toggleTransCurveAnimate() {
	transCurveAnimate = !transCurveAnimate
}

func setTransCurveNumPoints(v float64) {
	transCurveNumPoints = v
}

func setTransCurveClose(v bool) {
	transCurveClose = v
}

func setTransCurvePreserveXScale(v bool) {
	transCurvePreserveXScale = v
}

func setTransCurveFixedLen(v bool) {
	transCurveFixedLen = v
}
