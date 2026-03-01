package main

import (
	"fmt"
	"math"
	"syscall/js"

	agg "agg_go"
	"agg_go/internal/basics"
)

var (
	width, height = 800, 600
	ctx           *agg.Context
	canvasBuf     []uint8
)

func main() {
	fmt.Println("AGG Go Web Demo Initializing...")

	// Initialize the context and buffer
	ctx = agg.NewContext(width, height)
	canvasBuf = ctx.GetImage().Data

	// Expose Go functions to JavaScript
	js.Global().Set("renderDemo", js.FuncOf(renderDemo))
	js.Global().Set("getCanvasDimensions", js.FuncOf(getCanvasDimensions))

	// Keep the Go program running
	select {}
}

func getCanvasDimensions(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"width":  width,
		"height": height,
	}
}

func renderDemo(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return nil
	}

	demoType := args[0].String()
	ctx.Clear(agg.White)

	switch demoType {
	case "lines":
		drawLinesDemo()
	case "circles":
		drawCirclesDemo()
	case "starburst":
		drawStarburstDemo()
	case "rects":
		drawRectsDemo()
	case "lion":
		drawLionDemo()
	case "gradients":
		drawGradientsDemo()
	case "aa":
		drawAADemo()
	default:
		drawLinesDemo()
	}

	// Copy the rendered buffer to the JavaScript Uint8ClampedArray
	if len(args) >= 2 {
		jsBuf := args[1]
		js.CopyBytesToJS(jsBuf, canvasBuf)
	}

	return nil
}

var lionPaths []LionPath

func drawLionDemo() {
	if lionPaths == nil {
		lionPaths = ParseLion()
	}

	// Calculate bounding box and scale to fit
	// For simplicity, we use hardcoded scale/offset for now or we could calculate it.
	// The lion coordinates are roughly 0-250 for X and 0-400 for Y.

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Center and scale
	scale := 1.2
	offsetX, offsetY := 250.0, 100.0

	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()

		agg2d.ResetPath()
		// We need to iterate over the PathStorage and add vertices to agg2d
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			// Apply scaling and offset manually for now
			tx, ty := x*scale+offsetX, y*scale+offsetY

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(tx, ty)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}
}

func drawGradientsDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Linear Gradient Rectangle
	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(1.0)
	agg2d.FillLinearGradient(100, 100, 300, 300, agg.Red, agg.Blue, 1.0)
	ctx.FillRectangle(100, 100, 200, 200)
	ctx.DrawRectangle(100, 100, 200, 200)

	// 2. Radial Gradient Circle
	agg2d.FillRadialGradient(500, 200, 100, agg.Yellow, agg.Transparent, 1.0)
	ctx.FillCircle(500, 200, 100)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(500, 200, 100)

	// 3. Radial Gradient with Multi-Stop (3 colors)
	agg2d.FillRadialGradientMultiStop(400, 450, 120, agg.Green, agg.White, agg.Red)
	ctx.FillCircle(400, 450, 120)
}

func drawAADemo() {
	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(1.0)

	// Draw lines with sub-pixel increments
	for i := 0; i < 20; i++ {
		offset := float64(i) / 10.0
		ctx.DrawLine(50+offset, 50+float64(i)*20, 250+offset, 70+float64(i)*20)

		// Add some text labels if possible
		// (Assuming text is supported)
	}

	// Draw a circle with sub-pixel movement
	for i := 0; i < 10; i++ {
		offset := float64(i) / 5.0
		ctx.SetColor(agg.RGBA(0.1, 0.5, 0.8, 0.3))
		ctx.FillCircle(400+offset*10, 300+offset*10, 50)
		ctx.SetColor(agg.Black)
		ctx.DrawCircle(400+offset*10, 300+offset*10, 50)
	}
}

func drawLinesDemo() {
	// Axes
	ctx.SetColor(agg.RGB(0.9, 0.9, 0.95))
	for y := 0; y < height; y += 40 {
		ctx.DrawLine(0, float64(y), float64(width-1), float64(y))
	}
	for x := 0; x < width; x += 40 {
		ctx.DrawLine(float64(x), 0, float64(x), float64(height-1))
	}

	// Main diagonals
	ctx.SetColor(agg.Blue)
	ctx.DrawLine(0, 0, float64(width-1), float64(height-1))
	ctx.SetColor(agg.Red)
	ctx.DrawLine(float64(width-1), 0, 0, float64(height-1))

	// Thick lines showcase
	ctx.SetColor(agg.RGB(0.2, 0.2, 0.2))
	ctx.DrawThickLine(60, 420, 260, 420, 1)
	ctx.SetColor(agg.RGB(0.0, 0.4, 0.9))
	ctx.DrawThickLine(60, 390, 260, 390, 4)
	ctx.SetColor(agg.RGB(0.9, 0.3, 0.1))
	ctx.DrawThickLine(60, 360, 260, 360, 8)
	ctx.SetColor(agg.RGB(0.4, 0.7, 0.2))
	ctx.DrawThickLine(320, 360, 540, 420, 10)
}

func drawCirclesDemo() {
	ctx.SetColor(agg.RGB(0.2, 0.6, 1.0))
	for i := 0; i < 20; i++ {
		r := 10.0 + float64(i)*5.0
		ctx.DrawCircle(float64(width/2), float64(height/2), r)
	}
}

func drawStarburstDemo() {
	cx, cy := float64(width/2), float64(height/2)
	ctx.SetColor(agg.Green)
	for i := 0; i < 36; i++ {
		angle := float64(i) * (math.Pi / 18.0) // every 10 degrees
		x := cx + 250.0*math.Cos(angle)
		y := cy + 250.0*math.Sin(angle)
		ctx.DrawLine(cx, cy, x, y)
	}
}

func drawRectsDemo() {
	for i := 0; i < 15; i++ {
		x := 100.0 + float64(i)*20.0
		y := 100.0 + float64(i)*15.0
		// Using RGBA for semi-transparent fills
		ctx.SetColor(agg.RGBA(0.1, 0.4, 0.8, 0.5))
		ctx.FillRectangle(x, y, 200.0, 150.0)

		ctx.SetColor(agg.RGB(0.1, 0.4, 0.8))
		ctx.SetStrokeWidth(2.0)
		ctx.DrawRectangle(x, y, 200.0, 150.0)
	}

	// Add some rounded rectangles
	for i := 0; i < 5; i++ {
		x := 450.0 + float64(i)*10.0
		y := 100.0 + float64(i)*40.0
		ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
		ctx.FillRoundedRectangle(x, y, 150.0, 100.0, 20.0)
		ctx.SetColor(agg.RGB(0.8, 0.2, 0.1))
		ctx.DrawRoundedRectangle(x, y, 150.0, 100.0, 20.0)
	}
}
