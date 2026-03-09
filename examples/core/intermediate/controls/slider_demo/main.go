// Slider Control Demo
//
// This example demonstrates the AGG slider control with various configurations.
package main

import (
	"fmt"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.FillColor(agg.Black)
	a.TextAlignment(agg.AlignLeft, agg.AlignTop)

	// Create slider configurations
	basic := slider.NewSliderCtrl(10, 10, 210, 30, false)
	basic.SetRange(0, 100)
	basic.SetValue(50)
	basic.SetLabel("Basic: %.0f")

	temp := slider.NewSliderCtrl(10, 40, 210, 60, false)
	temp.SetRange(-10, 40)
	temp.SetValue(20)
	temp.SetLabel("Temperature: %.1f°C")
	temp.SetPointerColor(color.NewRGBA(0.8, 0.3, 0.3, 1.0))

	volume := slider.NewSliderCtrl(10, 70, 210, 90, false)
	volume.SetRange(0, 10)
	volume.SetNumSteps(10)
	volume.SetValue(7)
	volume.SetLabel("Volume: %.0f")
	volume.SetPointerColor(color.NewRGBA(0.3, 0.8, 0.3, 1.0))

	// Display info
	lines := []string{
		"Slider Control Demo",
		"",
		fmt.Sprintf("Basic slider: value=%.1f, paths=%d", basic.Value(), basic.NumPaths()),
		fmt.Sprintf("Temperature slider: value=%.1f", temp.Value()),
		fmt.Sprintf("Volume slider: value=%.1f, steps=%d", volume.Value(), 10),
	}

	for i, line := range lines {
		a.Text(10, float64(100+i*20), line, false, 0, 0)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Slider Demo",
		Width:  400,
		Height: 300,
	}, &demo{})
}
