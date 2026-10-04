//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/basics"
)

// Setter and getter callbacks exported to JavaScript for demo controls.

func setMultiClipNJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setMultiClipN(args[0].Float())
	}
	return nil
}

func setCompOpJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setCompOp(args[0].Int())
	}
	return nil
}

func setCompAlphaSrcJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setCompAlphaSrc(args[0].Float())
	}
	return nil
}

func setCompAlphaDstJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setCompAlphaDst(args[0].Float())
	}
	return nil
}

func setPerspectiveTypeJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setPerspectiveType(args[0].Int())
	}
	return nil
}

func setDistortionsImageJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setDistortionsImageType(args[0].Int())
	}
	return nil
}

func toggleTransCurveAnimateJS(this js.Value, args []js.Value) interface{} {
	toggleTransCurveAnimate()
	return nil
}

func setTransCurveNumPointsJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurveNumPoints(args[0].Float())
	}
	return nil
}

func setTransCurveCloseJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurveClose(args[0].Bool())
	}
	return nil
}

func setTransCurvePreserveXScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurvePreserveXScale(args[0].Bool())
	}
	return nil
}

func setTransCurveFixedLenJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurveFixedLen(args[0].Bool())
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

func setDashSmooth(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		dashSmooth = args[0].Float()
	}
	return nil
}

func setDashCap(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		dashCap = args[0].Int()
	}
	return nil
}

func setDashClosed(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		dashClosed = args[0].Bool()
	}
	return nil
}

func setDashEvenOdd(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		dashEvenOdd = args[0].Bool()
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
		sboolOp = args[0].Int()
	}
	return nil
}

func setSBoolMul1(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		sboolMul1 = args[0].Float()
	}
	return nil
}

func setSBoolMul2(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		sboolMul2 = args[0].Float()
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
		"x0": dashPts[0].X, "y0": dashPts[0].Y,
		"x1": dashPts[1].X, "y1": dashPts[1].Y,
		"x2": dashPts[2].X, "y2": dashPts[2].Y,
	}
}

func setDashNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		dashPts[0] = basics.PointD{X: args[0].Float(), Y: args[1].Float()}
		dashPts[1] = basics.PointD{X: args[2].Float(), Y: args[3].Float()}
		dashPts[2] = basics.PointD{X: args[4].Float(), Y: args[5].Float()}
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
	if !sboolInited {
		sboolInit()
	}
	q1 := getSBoolQuad1()
	q2 := getSBoolQuad2()
	return map[string]interface{}{
		"p1x0": q1[0], "p1y0": q1[1],
		"p1x1": q1[2], "p1y1": q1[3],
		"p1x2": q1[4], "p1y2": q1[5],
		"p1x3": q1[6], "p1y3": q1[7],
		"p2x0": q2[0], "p2y0": q2[1],
		"p2x1": q2[2], "p2y1": q2[3],
		"p2x2": q2[4], "p2y2": q2[5],
		"p2x3": q2[6], "p2y3": q2[7],
	}
}

func setSBoolNodes(this js.Value, args []js.Value) interface{} {
	if !sboolInited {
		sboolInit()
	}
	if len(args) >= 16 {
		sboolQuad1[0] = args[0].Float()
		sboolQuad1[1] = args[1].Float()
		sboolQuad1[2] = args[2].Float()
		sboolQuad1[3] = args[3].Float()
		sboolQuad1[4] = args[4].Float()
		sboolQuad1[5] = args[5].Float()
		sboolQuad1[6] = args[6].Float()
		sboolQuad1[7] = args[7].Float()
		sboolQuad2[0] = args[8].Float()
		sboolQuad2[1] = args[9].Float()
		sboolQuad2[2] = args[10].Float()
		sboolQuad2[3] = args[11].Float()
		sboolQuad2[4] = args[12].Float()
		sboolQuad2[5] = args[13].Float()
		sboolQuad2[6] = args[14].Float()
		sboolQuad2[7] = args[15].Float()
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

// --- image1 JS setters ---
func setImg1AngleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		img1Angle = args[0].Float()
	}
	return nil
}

func setImg1ScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		img1Scale = args[0].Float()
	}
	return nil
}

// --- image_transforms JS setters ---
func setImgTransPolygonAngleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setImgTransPolygonAngle(args[0].Float())
	}
	return nil
}

func setImgTransPolygonScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setImgTransPolygonScale(args[0].Float())
	}
	return nil
}

func setImgTransImageAngleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setImgTransImageAngle(args[0].Float())
	}
	return nil
}

func setImgTransImageScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setImgTransImageScale(args[0].Float())
	}
	return nil
}

func setImgTransExampleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setImgTransExample(args[0].Int())
	}
	return nil
}

// --- pattern_fill JS setters ---
func setPatFillPolygonAngleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setPatFillPolygonAngle(args[0].Float())
	}
	return nil
}

func setPatFillPolygonScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setPatFillPolygonScale(args[0].Float())
	}
	return nil
}

func setPatFillPatternAngleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setPatFillPatternAngle(args[0].Float())
	}
	return nil
}

func setPatFillPatternSizeJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setPatFillPatternSize(args[0].Float())
	}
	return nil
}

func getCanvasDimensions(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"width":  width,
		"height": height,
	}
}

func toggleTransCurve2AnimateJS(this js.Value, args []js.Value) interface{} {
	toggleTransCurve2Animate()
	return nil
}

func setTransCurve2NumPointsJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurve2NumPoints(args[0].Float())
	}
	return nil
}

func setTransCurve2FixedLenJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurve2FixedLen(args[0].Bool())
	}
	return nil
}

func setTransCurve2PreserveXScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setTransCurve2PreserveXScale(args[0].Bool())
	}
	return nil
}

func setMeshSizeJS(this js.Value, args []js.Value) interface{} {
	if len(args) >= 2 {
		setMeshSize(args[0].Int(), args[1].Int())
	}
	return nil
}

func setAlphaMask2NumEllipsesJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setAlphaMask2NumEllipses(args[0].Float())
	}
	return nil
}

func setLionLensScaleJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setLionLensScale(args[0].Float())
	}
	return nil
}

func setLionLensRadiusJS(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		setLionLensRadius(args[0].Float())
	}
	return nil
}

func setBlurRadius(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		blurRadius = args[0].Float()
	}
	return nil
}

func setBlurMethod(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		blurMethod = args[0].Int()
	}
	return nil
}

func setCirclesSelectivity(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		selectivity = args[0].Float()
	}
	return nil
}

func setCirclesSize(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		sizeScale = args[0].Float()
	}
	return nil
}

func setCirclesZRange(this js.Value, args []js.Value) interface{} {
	if len(args) >= 2 {
		zRangeLow = args[0].Float()
		zRangeHigh = args[1].Float()
	}
	return nil
}
