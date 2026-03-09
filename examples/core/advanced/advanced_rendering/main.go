// Example demonstrating Phase 5: Advanced Rendering features
// This showcases transformations, viewport handling, and advanced stroke/fill features
package main

import (
	"fmt"

	"agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	agg2d := ctx.GetAgg2D()
	agg2d.ClearAll(agg.White)

	agg2d.FillColor(agg.Black)
	agg2d.TextAlignment(agg.AlignLeft, agg.AlignTop)

	lines := []string{
		"AGG2D Phase 5: Advanced Rendering Demo",
		"",
		"Transformation Demonstrations:",
		"  Translation: Translate(100, 50) -> moves origin",
		"  Scaling: Scale(2.0, 1.5) -> stretches by factors",
		"  Rotation: Rotate(45°) -> rotates coordinate system",
		"  Combined: Scale + Rotate + Translate creates complex transformation",
		"",
		"Viewport Transformations:",
		"  Anisotropic, XMidYMid, XMinYMin, XMaxYMax supported",
		"",
		"Advanced Stroke Features:",
	}

	// Line attributes demo
	agg2d.LineWidth(3.0)
	agg2d.LineCap(agg.CapRound)
	agg2d.LineJoin(agg.JoinBevel)
	agg2d.MiterLimit(4.0)

	lines = append(lines,
		fmt.Sprintf("  Line width: %g, Cap: Round, Join: Bevel", agg2d.GetLineWidth()),
		fmt.Sprintf("  Miter limit: %g", agg2d.GetMiterLimit()),
		"",
		"Fill Rules:",
	)

	agg2d.FillEvenOdd(true)
	lines = append(lines, fmt.Sprintf("  Even-odd: %s", agg2d.FillRuleDescription()))
	agg2d.FillEvenOdd(false)
	lines = append(lines, fmt.Sprintf("  Non-zero: %s", agg2d.FillRuleDescription()))

	for i, line := range lines {
		agg2d.Text(10, float64(10+i*16), line, false, 0, 0)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Advanced Rendering",
		Width:  600,
		Height: 400,
	}, &demo{})
}
