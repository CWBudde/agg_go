// Radio Button Group (Rbox) Control Demo
//
// This example demonstrates the AGG radio button group control.
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/color"
	"agg_go/internal/ctrl/rbox"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.FillColor(agg.Black)
	a.TextAlignment(agg.AlignLeft, agg.AlignTop)

	// Create some rbox controls and display their state
	colorRbox := rbox.NewDefaultRboxCtrl(10, 10, 150, 120, false)
	colorRbox.AddItem("Red")
	colorRbox.AddItem("Green")
	colorRbox.AddItem("Blue")
	colorRbox.AddItem("Yellow")
	colorRbox.SetCurItem(0)

	qualityRbox := rbox.NewDefaultRboxCtrl(170, 10, 300, 100, false)
	qualityRbox.AddItem("Low Quality")
	qualityRbox.AddItem("Medium Quality")
	qualityRbox.AddItem("High Quality")
	qualityRbox.SetCurItem(1)
	qualityRbox.SetBackgroundColor(color.NewRGBA(0.95, 0.95, 1.0, 1.0))
	qualityRbox.SetBorderColor(color.NewRGBA(0.2, 0.2, 0.6, 1.0))
	qualityRbox.SetTextColor(color.NewRGBA(0.1, 0.1, 0.4, 1.0))
	qualityRbox.SetInactiveColor(color.NewRGBA(0.5, 0.5, 0.7, 1.0))
	qualityRbox.SetActiveColor(color.NewRGBA(0.8, 0.2, 0.2, 1.0))

	// Display info about the rboxes
	lines := []string{
		"Radio Button Group Demo",
		"",
		fmt.Sprintf("Color Group: %d items, selected: %d", colorRbox.NumItems(), colorRbox.CurItem()),
		fmt.Sprintf("Quality Group: %d items, selected: %d", qualityRbox.NumItems(), qualityRbox.CurItem()),
	}

	for i, line := range lines {
		a.Text(10, float64(140+i*20), line, false, 0, 0)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Radio Button (Rbox) Demo",
		Width:  400,
		Height: 300,
	}, &demo{})
}
