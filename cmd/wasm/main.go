package main

import (
	"fmt"
	"syscall/js"

	agg "agg_go"
)

var (
	width, height = 800, 600
	ctx           *agg.Context
	canvasBuf     []uint8
	lionPaths     []LionPath
)

func main() {
	fmt.Println("AGG Go Web Demo Initializing...")

	// Initialize the context and buffer
	ctx = agg.NewContext(width, height)
	canvasBuf = ctx.GetImage().Data

	// Expose Go functions to JavaScript
	js.Global().Set("renderDemo", js.FuncOf(renderDemo))
	js.Global().Set("getCanvasDimensions", js.FuncOf(getCanvasDimensions))
	js.Global().Set("onMouseDown", js.FuncOf(onMouseDown))
	js.Global().Set("onMouseMove", js.FuncOf(onMouseMove))
	js.Global().Set("onMouseUp", js.FuncOf(onMouseUp))
	js.Global().Set("setAAZoom", js.FuncOf(setAAZoom))

	// Keep the Go program running
	select {}
}

func onMouseDown(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return nil
	}
	demoType := args[0].String()
	x := args[1].Float()
	y := args[2].Float()

	if demoType == "aa" {
		return handleAAMouseDown(x, y)
	}
	return false
}

func onMouseMove(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return nil
	}
	demoType := args[0].String()
	x := args[1].Float()
	y := args[2].Float()

	if demoType == "aa" {
		return handleAAMouseMove(x, y)
	}
	return false
}

func onMouseUp(this js.Value, args []js.Value) interface{} {
	demoType := args[0].String()
	if demoType == "aa" {
		handleAAMouseUp()
	}
	return nil
}

func setAAZoom(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		aaPixelSize = args[0].Float()
	}
	return nil
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
	case "blend":
		drawBlendModesDemo()
	case "bspline":
		drawBSplineDemo()
	case "dash":
		drawDashDemo()
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
