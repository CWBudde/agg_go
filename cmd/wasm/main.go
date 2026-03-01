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
	js.Global().Set("setDashWidth", js.FuncOf(setDashWidth))
	js.Global().Set("setDashClosed", js.FuncOf(setDashClosed))
	js.Global().Set("setGouraudDilation", js.FuncOf(setGouraudDilation))

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
	if demoType == "dash" {
		return handleDashMouseDown(x, y)
	}
	if demoType == "gouraud" {
		return handleGouraudMouseDown(x, y)
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
	if demoType == "dash" {
		return handleDashMouseMove(x, y)
	}
	if demoType == "gouraud" {
		return handleGouraudMouseMove(x, y)
	}
	return false
}

func onMouseUp(this js.Value, args []js.Value) interface{} {
	demoType := args[0].String()
	if demoType == "aa" {
		handleAAMouseUp()
	}
	if demoType == "dash" {
		handleDashMouseUp()
	}
	if demoType == "gouraud" {
		handleGouraudMouseUp()
	}
	return nil
}

func setAAZoom(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		aaPixelSize = args[0].Float()
	}
	return nil
}

func setDashWidth(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		dashWidth = args[0].Float()
	}
	return nil
}

func setDashClosed(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		dashClosed = args[0].Bool()
	}
	return nil
}

func setGouraudDilation(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		gouraudDilation = args[0].Float()
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
	fmt.Printf("Rendering demo: %s\n", demoType)
	
	ctx.Clear(agg.White)
	
	// Reset specific demo state if needed
	if demoType != "lion" {
		lionPaths = nil
	}

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
	case "gouraud":
		drawGouraudDemo()
	case "aatest":
		drawAATestDemo()
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
