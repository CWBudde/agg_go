package main

import (
	"fmt"
	"math"
	"syscall/js"

	agg "agg_go"
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
		ctx.FillRectangle(x, y, x+200.0, y+150.0)
		
		ctx.SetColor(agg.RGB(0.1, 0.4, 0.8))
		ctx.SetStrokeWidth(2.0)
		ctx.DrawRectangle(x, y, x+200.0, y+150.0)
	}
}
