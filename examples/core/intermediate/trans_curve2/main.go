// Port of AGG C++ trans_curve2.cpp – double-path (two-rail) curve transform.
//
// Renders a text paragraph warped between two B-spline rails via
// transform.TransDoublePath, faithfully mirroring the C++ demo (which strokes
// "Times New Roman" outline glyphs along the two paths). Like trans_curve, the
// portable embedded GSV vector font stands in for the platform TrueType face so
// the output is deterministic and needs no external font asset; the FreeType
// intent is covered by the trans_curve2_ft variant.
//
// Standalone/web parity: the rendering core lives in internal/demo/transcurve
// (DrawDouble) and is shared verbatim with the WASM demo (cmd/wasm/
// demo_trans_curve2.go), so both surfaces draw the identical scene.
//
// C++ reference: ../agg-2.6/agg-src/examples/trans_curve2.cpp.
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/demo/transcurve"
)

const (
	width  = 600
	height = 600
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	transcurve.DrawDouble(ctx, transcurve.DoubleConfig{
		Points1:         transcurve.DefaultPoints1,
		Points2:         transcurve.DefaultPoints2,
		NumIntermediate: 200,
		PreserveXScale:  true,
		FixedLength:     true,
		BaseLength:      transcurve.DefaultDoubleBaseLength,
		BaseHeight:      transcurve.DefaultDoubleBaseHeight,
		Text:            transcurve.DefaultDoubleText,
	})
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Trans Curve 2",
		Width:  width,
		Height: height,
	}, &demo{})
}
