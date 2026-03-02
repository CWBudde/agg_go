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
	js.Global().Set("setAANodes", js.FuncOf(setAANodes))
	js.Global().Set("getAANodes", js.FuncOf(getAANodes))
	js.Global().Set("setDashWidth", js.FuncOf(setDashWidth))
	js.Global().Set("setDashClosed", js.FuncOf(setDashClosed))
	js.Global().Set("setGouraudDilation", js.FuncOf(setGouraudDilation))
	js.Global().Set("setImageFilter", js.FuncOf(setImageFilter))
	js.Global().Set("setImageFilterRadius", js.FuncOf(setImageFilterRadius))
	js.Global().Set("setImageFilterAngle", js.FuncOf(setImageFilterAngle))
	js.Global().Set("setSBoolOp", js.FuncOf(setSBoolOp))
	js.Global().Set("setStrokeJoin", js.FuncOf(setStrokeJoin))
	js.Global().Set("setStrokeCap", js.FuncOf(setStrokeCap))
	js.Global().Set("setStrokeWidth", js.FuncOf(setStrokeWidth))
	js.Global().Set("setStrokeMiterLimit", js.FuncOf(setStrokeMiterLimit))
	js.Global().Set("setContourWidth", js.FuncOf(setContourWidth))
	js.Global().Set("setContourCloseMode", js.FuncOf(setContourCloseMode))
	js.Global().Set("setContourAutoDetect", js.FuncOf(setContourAutoDetect))
	// Node getters/setters for URL persistence
	js.Global().Set("getDashNodes", js.FuncOf(getDashNodes))
	js.Global().Set("setDashNodes", js.FuncOf(setDashNodes))
	js.Global().Set("getGouraudNodes", js.FuncOf(getGouraudNodes))
	js.Global().Set("setGouraudNodes", js.FuncOf(setGouraudNodes))
	js.Global().Set("getSBoolNodes", js.FuncOf(getSBoolNodes))
	js.Global().Set("setSBoolNodes", js.FuncOf(setSBoolNodes))
	js.Global().Set("getStrokeNodes", js.FuncOf(getStrokeNodes))
	js.Global().Set("setStrokeNodes", js.FuncOf(setStrokeNodes))
	js.Global().Set("setGammaValue", js.FuncOf(setGammaValue))
	js.Global().Set("setGammaThickness", js.FuncOf(setGammaThickness))
	js.Global().Set("setGammaContrast", js.FuncOf(setGammaContrast))
	js.Global().Set("setLionOutlineWidth", js.FuncOf(setLionOutlineWidth))
	js.Global().Set("setCompAlpha", js.FuncOf(setCompAlpha))
	js.Global().Set("setRRRadius", js.FuncOf(setRRRadius))
	js.Global().Set("setRROffset", js.FuncOf(setRROffset))
	js.Global().Set("setRRDarkBg", js.FuncOf(setRRDarkBg))
	js.Global().Set("getRRNodes", js.FuncOf(getRRNodes))
	js.Global().Set("setRRNodes", js.FuncOf(setRRNodes))
	js.Global().Set("getAlphaGradNodes", js.FuncOf(getAlphaGradNodes))
	js.Global().Set("setAlphaGradNodes", js.FuncOf(setAlphaGradNodes))
	js.Global().Set("setPerspectiveType", js.FuncOf(setPerspectiveTypeJS))
	js.Global().Set("toggleTransCurveAnimate", js.FuncOf(toggleTransCurveAnimateJS))

	// Keep the Go program running
	select {}
}

func setPerspectiveTypeJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setPerspectiveType(args[0].Int())
	}
	return nil
}

func toggleTransCurveAnimateJS(this js.Value, args []js.Value) interface{} {
	toggleTransCurveAnimate()
	return nil
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
	if demoType == "convstroke" {
		return handleConvStrokeMouseDown(x, y)
	}
	if demoType == "gamma" {
		return handleGammaCorrectionMouseDown(x, y)
	}
	if demoType == "lionoutline" {
		right := len(args) >= 4 && args[3].Bool()
		return handleLionOutlineMouseDown(x, y, right)
	}
	if demoType == "roundedrect" {
		return handleRoundedRectMouseDown(x, y)
	}
	if demoType == "alphagrad" {
		return handleAlphaGradMouseDown(x, y)
	}
	if demoType == "rasterizers" {
		return handleRasterizersMouseDown(x, y)
	}
	if demoType == "perspective" {
		return handlePerspectiveMouseDown(x, y)
	}
	if demoType == "bezier_div" {
		return handleBezierDivMouseDown(x, y)
	}
	if demoType == "trans_curve" {
		return handleTransCurveMouseDown(x, y)
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
	if demoType == "convstroke" {
		return handleConvStrokeMouseMove(x, y)
	}
	if demoType == "gamma" {
		return handleGammaCorrectionMouseMove(x, y)
	}
	if demoType == "lionoutline" {
		right := len(args) >= 4 && args[3].Bool()
		return handleLionOutlineMouseMove(x, y, right)
	}
	if demoType == "roundedrect" {
		return handleRoundedRectMouseMove(x, y)
	}
	if demoType == "alphagrad" {
		return handleAlphaGradMouseMove(x, y)
	}
	if demoType == "rasterizers" {
		return handleRasterizersMouseMove(x, y)
	}
	if demoType == "perspective" {
		return handlePerspectiveMouseMove(x, y)
	}
	if demoType == "bezier_div" {
		return handleBezierDivMouseMove(x, y)
	}
	if demoType == "trans_curve" {
		return handleTransCurveMouseMove(x, y)
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
	if demoType == "convstroke" {
		handleConvStrokeMouseUp()
	}
	if demoType == "lionoutline" {
		handleLionOutlineMouseUp()
	}
	if demoType == "roundedrect" {
		handleRoundedRectMouseUp()
	}
	if demoType == "alphagrad" {
		handleAlphaGradMouseUp()
	}
	if demoType == "rasterizers" {
		handleRasterizersMouseUp()
	}
	if demoType == "perspective" {
		handlePerspectiveMouseUp()
	}
	if demoType == "bezier_div" {
		handleBezierDivMouseUp()
	}
	if demoType == "trans_curve" {
		handleTransCurveMouseUp()
	}
	return nil
}

func setAAZoom(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		aaPixelSize = args[0].Float()
	}
	return nil
}

func setAANodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		aaTriangleX[0] = args[0].Float()
		aaTriangleY[0] = args[1].Float()
		aaTriangleX[1] = args[2].Float()
		aaTriangleY[1] = args[3].Float()
		aaTriangleX[2] = args[4].Float()
		aaTriangleY[2] = args[5].Float()
	}
	return nil
}

func getAANodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"x0": aaTriangleX[0], "y0": aaTriangleY[0],
		"x1": aaTriangleX[1], "y1": aaTriangleY[1],
		"x2": aaTriangleX[2], "y2": aaTriangleY[2],
	}
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

func setStrokeJoin(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		v := args[0].Int()
		if v >= 0 && v < len(strokeJoins) {
			strokeJoin = v
		}
	}
	return nil
}

func setStrokeCap(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		v := args[0].Int()
		if v >= 0 && v < len(strokeCaps) {
			strokeCap = v
		}
	}
	return nil
}

func setStrokeWidth(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		strokeWidth = args[0].Float()
	}
	return nil
}

func setStrokeMiterLimit(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		strokeMiterLimit = args[0].Float()
	}
	return nil
}

func setContourWidth(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		contourWidth = args[0].Float()
	}
	return nil
}

func setContourCloseMode(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		contourCloseMode = args[0].Int()
	}
	return nil
}

func setContourAutoDetect(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		contourAutoDetect = args[0].Bool()
	}
	return nil
}

func setGammaValue(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		gammaValue = args[0].Float()
	}
	return nil
}

func setGammaThickness(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		gammaThick = args[0].Float()
	}
	return nil
}

func setGammaContrast(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		gammaContrast = args[0].Float()
	}
	return nil
}

func setLionOutlineWidth(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		lionOutlineWidth = args[0].Float()
	}
	return nil
}

func setCompAlpha(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		v := args[0].Int()
		if v < 0 {
			v = 0
		} else if v > 255 {
			v = 255
		}
		compAlpha = v
	}
	return nil
}

func setRRRadius(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		rrRadius = args[0].Float()
	}
	return nil
}

func setRROffset(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		rrOffset = args[0].Float()
	}
	return nil
}

func setRRDarkBg(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		rrDarkBg = args[0].Bool()
	}
	return nil
}

func getRRNodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"x0": rrPts[0][0], "y0": rrPts[0][1],
		"x1": rrPts[1][0], "y1": rrPts[1][1],
	}
}

func setRRNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 4 {
		rrPts[0][0] = args[0].Float()
		rrPts[0][1] = args[1].Float()
		rrPts[1][0] = args[2].Float()
		rrPts[1][1] = args[3].Float()
	}
	return nil
}

// --- Node getters/setters for URL persistence ---

func getDashNodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"x0": dashTriangleX[0], "y0": dashTriangleY[0],
		"x1": dashTriangleX[1], "y1": dashTriangleY[1],
		"x2": dashTriangleX[2], "y2": dashTriangleY[2],
	}
}

func setDashNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		dashTriangleX[0] = args[0].Float()
		dashTriangleY[0] = args[1].Float()
		dashTriangleX[1] = args[2].Float()
		dashTriangleY[1] = args[3].Float()
		dashTriangleX[2] = args[4].Float()
		dashTriangleY[2] = args[5].Float()
	}
	return nil
}

func getGouraudNodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"x0": gouraudX[0], "y0": gouraudY[0],
		"x1": gouraudX[1], "y1": gouraudY[1],
		"x2": gouraudX[2], "y2": gouraudY[2],
	}
}

func setGouraudNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		gouraudX[0] = args[0].Float()
		gouraudY[0] = args[1].Float()
		gouraudX[1] = args[2].Float()
		gouraudY[1] = args[3].Float()
		gouraudX[2] = args[4].Float()
		gouraudY[2] = args[5].Float()
	}
	return nil
}

func getSBoolNodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"p1x0": sboolPoly1X[0], "p1y0": sboolPoly1Y[0],
		"p1x1": sboolPoly1X[1], "p1y1": sboolPoly1Y[1],
		"p1x2": sboolPoly1X[2], "p1y2": sboolPoly1Y[2],
		"p1x3": sboolPoly1X[3], "p1y3": sboolPoly1Y[3],
		"p2x0": sboolPoly2X[0], "p2y0": sboolPoly2Y[0],
		"p2x1": sboolPoly2X[1], "p2y1": sboolPoly2Y[1],
		"p2x2": sboolPoly2X[2], "p2y2": sboolPoly2Y[2],
		"p2x3": sboolPoly2X[3], "p2y3": sboolPoly2Y[3],
	}
}

func setSBoolNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 16 {
		sboolPoly1X[0] = args[0].Float()
		sboolPoly1Y[0] = args[1].Float()
		sboolPoly1X[1] = args[2].Float()
		sboolPoly1Y[1] = args[3].Float()
		sboolPoly1X[2] = args[4].Float()
		sboolPoly1Y[2] = args[5].Float()
		sboolPoly1X[3] = args[6].Float()
		sboolPoly1Y[3] = args[7].Float()
		sboolPoly2X[0] = args[8].Float()
		sboolPoly2Y[0] = args[9].Float()
		sboolPoly2X[1] = args[10].Float()
		sboolPoly2Y[1] = args[11].Float()
		sboolPoly2X[2] = args[12].Float()
		sboolPoly2Y[2] = args[13].Float()
		sboolPoly2X[3] = args[14].Float()
		sboolPoly2Y[3] = args[15].Float()
	}
	return nil
}

func getStrokeNodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"x0": strokePts[0][0], "y0": strokePts[0][1],
		"x1": strokePts[1][0], "y1": strokePts[1][1],
		"x2": strokePts[2][0], "y2": strokePts[2][1],
	}
}

func setStrokeNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		strokePts[0][0] = args[0].Float()
		strokePts[0][1] = args[1].Float()
		strokePts[1][0] = args[2].Float()
		strokePts[1][1] = args[3].Float()
		strokePts[2][0] = args[4].Float()
		strokePts[2][1] = args[5].Float()
	}
	return nil
}

func getAlphaGradNodes(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"x0": alphaGradPts[0][0], "y0": alphaGradPts[0][1],
		"x1": alphaGradPts[1][0], "y1": alphaGradPts[1][1],
		"x2": alphaGradPts[2][0], "y2": alphaGradPts[2][1],
	}
}

func setAlphaGradNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		alphaGradPts[0][0] = args[0].Float()
		alphaGradPts[0][1] = args[1].Float()
		alphaGradPts[1][0] = args[2].Float()
		alphaGradPts[1][1] = args[3].Float()
		alphaGradPts[2][0] = args[4].Float()
		alphaGradPts[2][1] = args[5].Float()
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
	if demoType != "lion" && demoType != "lionoutline" {
		lionPaths = nil
	}
	if demoType != "imagefilters" {
		testImage = nil
	}

	ctx.Clear(agg.White)

	switch demoType {
	case "agg2d":
		drawAgg2DDemo()
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
	case "convstroke":
		drawConvStrokeDemo()
	case "convcontour":
		drawConvContourDemo()
	case "gamma":
		drawGammaCorrectionDemo()
	case "lionoutline":
		drawLionOutlineDemo()
	case "roundedrect":
		drawRoundedRectDemo()
	case "component":
		drawComponentRenderingDemo()
	case "alphagrad":
		drawAlphaGradientDemo()
	case "rasterizers":
		drawRasterizersDemo()
	case "flash_rasterizer":
		drawFlashRasterizerDemo()
	case "perspective":
		drawPerspectiveDemo()
	case "bezier_div":
		drawBezierDivDemo()
	case "gouraud_mesh":
		drawGouraudMeshDemo()
	case "trans_curve":
		drawTransCurveDemo()
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
