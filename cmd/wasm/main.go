//go:build js && wasm
// +build js,wasm

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
	js.Global().Set("setImageFilter", js.FuncOf(setImageFilter))
	js.Global().Set("setImageFilterRadius", js.FuncOf(setImageFilterRadius))
	js.Global().Set("setImageFilterAngle", js.FuncOf(setImageFilterAngle))
	js.Global().Set("setSBoolOp", js.FuncOf(setSBoolOp))

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
	if demoType == "sbool" {
		return handleSBoolMouseDown(x, y)
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
	if demoType == "sbool" {
		return handleSBoolMouseMove(x, y)
	}
	return false
}

func onMouseUp(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return nil
	}
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
	if demoType == "sbool" {
		handleSBoolMouseUp()
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

func setImageFilter(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		imgFilterType = agg.ImageFilter(args[0].Int())
	}
	return nil
}

func setImageFilterRadius(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		imgFilterRadius = args[0].Float()
	}
	return nil
}

func setImageFilterAngle(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		imgFilterAngle = args[0].Float()
	}
	return nil
}

func setSBoolOp(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		sboolOp = SBoolOp(args[0].Int())
	}
	return nil
}

func getCanvasDimensions(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"width":  width,
		"height": height,
	}
}

const statusMsgID = "statusMsg"

func logStatus(msg string) {
	fmt.Println(msg)
	js.Global().Get("document").Call("getElementById", statusMsgID).Set("textContent", msg)
}

func renderDemo(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return nil
	}

	demoType := args[0].String()

	// Add panic recovery to prevent the WASM instance from dying silently
	defer func() {
		if r := recover(); r != nil {
			errStr := fmt.Sprintf("FATAL ERROR in %s: %v", demoType, r)
			logStatus(errStr)
			js.Global().Get("document").Call("getElementById", statusMsgID).Get("style").Set("color", "#ff3b30")
		}
	}()

	// Reset UI status color
	js.Global().Get("document").Call("getElementById", statusMsgID).Get("style").Set("color", "")
	logStatus("Rendering " + demoType + "...")

	// Release cached demo state when switching away from a demo.
	if demoType != "lion" {
		lionPaths = nil
	}
	if demoType != "imagefilters" {
		testImage = nil
	}

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
	case "gouraud":
		drawGouraudDemo()
	case "imagefilters":
		drawImageFiltersDemo()
	case "sbool":
		drawSBoolDemo()
	case "aatest":
		drawAATestDemo()
	default:
		logStatus("unknown demo type: " + demoType)
		return nil
	}

	// Copy the rendered buffer to the JavaScript Uint8ClampedArray
	if len(args) >= 2 {
		jsBuf := args[1]
		js.CopyBytesToJS(jsBuf, canvasBuf)
	}

	return nil
}
