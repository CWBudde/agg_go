// Go-idiomatic equivalent of AGG's trans_curve1.cpp: a text paragraph stroked
// and warped along a single interactive B-spline path. The portable embedded GSV
// vector font stands in for the platform "Times New Roman" TrueType face so the
// output is deterministic and needs no external font asset; the FreeType intent
// is covered by the trans_curve1_ft variant.
//
// Standalone/web parity: the rendering core lives in internal/demo/transcurve
// (Draw) and is shared verbatim with the WASM demo (cmd/wasm/demo_trans_curve.go),
// so both surfaces draw the identical scene.
//
// C++ reference: ../agg-2.6/agg-src/examples/trans_curve1.cpp.
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
	transcurve.Draw(ctx, transcurve.Config{
		Points:          transcurve.DefaultPoints,
		NumIntermediate: 200,
		PreserveXScale:  true,
		FixedLength:     true,
		BaseLength:      transcurve.DefaultBaseLength,
		Text:            transcurve.DefaultText,
	})
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Trans Curve 1",
		Width:  width,
		Height: height,
	}, &demo{})
}
