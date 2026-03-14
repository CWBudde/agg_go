// Go-idiomatic equivalent of AGG's trans_curve1.cpp using the embedded GSV font.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/transcurve"
)

const (
	width  = 600
	height = 600
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
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
	demorunner.Run(demorunner.Config{
		Title:  "Trans Curve 1",
		Width:  width,
		Height: height,
	}, &demo{})
}
