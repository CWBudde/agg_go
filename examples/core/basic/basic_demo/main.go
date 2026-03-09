// Package main demonstrates the basic usage of AGG platform support.
// This example shows basic rendering operations using the demorunner framework.
package main

import (
	"fmt"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

type demo struct {
	frameCount int
}

func (d *demo) Render(ctx *agg.Context) {
	d.frameCount++

	width := ctx.Width()
	height := ctx.Height()

	// Clear background with a gradient-like effect
	bgVal := float64((d.frameCount*2)%256) / 255.0
	ctx.Clear(agg.RGB(bgVal/4, bgVal/8, bgVal/2))

	// Draw a grid pattern
	ctx.SetColor(agg.RGB(0.5, 0.5, 0.5))
	for x := 0; x < width; x += 50 {
		ctx.DrawLine(float64(x), 0, float64(x), float64(height-1))
	}
	for y := 0; y < height; y += 50 {
		ctx.DrawLine(0, float64(y), float64(width-1), float64(y))
	}

	// Draw some shapes
	centerX, centerY := float64(width/2), float64(height/2)

	// Red rectangle
	ctx.SetColor(agg.Red)
	ctx.DrawRectangle(centerX-100, centerY-50, 80, 40)
	ctx.Fill()

	// Green circle outline
	ctx.SetColor(agg.Green)
	ctx.DrawCircle(centerX, centerY, 60)

	// Blue filled circle with partial transparency
	ctx.SetColor(agg.RGBA(0, 0, 1, 0.5))
	ctx.DrawCircle(centerX+20, centerY+20, 30)
	ctx.Fill()

	fmt.Printf("Frame: %d\n", d.frameCount)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Basic Demo", Width: 800, Height: 600}, &demo{})
}
