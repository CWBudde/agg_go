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
	js.Global().Set("setDashSmooth", js.FuncOf(setDashSmooth))
	js.Global().Set("setDashCap", js.FuncOf(setDashCap))
	js.Global().Set("setDashClosed", js.FuncOf(setDashClosed))
	js.Global().Set("setDashEvenOdd", js.FuncOf(setDashEvenOdd))
	js.Global().Set("setGouraudDilation", js.FuncOf(setGouraudDilation))
	js.Global().Set("setImageFilter", js.FuncOf(setImageFilter))
	js.Global().Set("setImageFilterRadius", js.FuncOf(setImageFilterRadius))
	js.Global().Set("setImageFilterAngle", js.FuncOf(setImageFilterAngle))
	js.Global().Set("setImageFltrGraphRadius", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageFltrGraphRadius(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setImageFltrGraphMask", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageFltrGraphMask(uint32(args[0].Int()))
		}
		return nil
	}))
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
	js.Global().Set("setLionAlpha", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			lionFillAlpha = args[0].Float()
		}
		return nil
	}))
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
	js.Global().Set("setDistortionsImage", js.FuncOf(setDistortionsImageJS))
	js.Global().Set("toggleTransCurveAnimate", js.FuncOf(toggleTransCurveAnimateJS))
	js.Global().Set("toggleTransCurve2Animate", js.FuncOf(toggleTransCurve2AnimateJS))
	js.Global().Set("setBlurRadius", js.FuncOf(setBlurRadius))
	js.Global().Set("setBlurMethod", js.FuncOf(setBlurMethod))
	js.Global().Set("setCirclesSelectivity", js.FuncOf(setCirclesSelectivity))
	js.Global().Set("setCirclesSize", js.FuncOf(setCirclesSize))
	js.Global().Set("setCirclesZRange", js.FuncOf(setCirclesZRange))
	js.Global().Set("setCompOp", js.FuncOf(setCompOpJS))
	js.Global().Set("setCompAlphaSrc", js.FuncOf(setCompAlphaSrcJS))
	js.Global().Set("setCompAlphaDst", js.FuncOf(setCompAlphaDstJS))
	js.Global().Set("setMultiClipN", js.FuncOf(setMultiClipNJS))
	js.Global().Set("setMeshSize", js.FuncOf(setMeshSizeJS))
	js.Global().Set("setAlphaMask2NumEllipses", js.FuncOf(setAlphaMask2NumEllipsesJS))
	js.Global().Set("setLionLensScale", js.FuncOf(setLionLensScaleJS))
	js.Global().Set("setLionLensRadius", js.FuncOf(setLionLensRadiusJS))
	js.Global().Set("setImg1Angle", js.FuncOf(setImg1AngleJS))
	js.Global().Set("setImg1Scale", js.FuncOf(setImg1ScaleJS))
	js.Global().Set("setImgTransPolygonAngle", js.FuncOf(setImgTransPolygonAngleJS))
	js.Global().Set("setImgTransPolygonScale", js.FuncOf(setImgTransPolygonScaleJS))
	js.Global().Set("setImgTransImageAngle", js.FuncOf(setImgTransImageAngleJS))
	js.Global().Set("setImgTransImageScale", js.FuncOf(setImgTransImageScaleJS))
	js.Global().Set("setImgTransExample", js.FuncOf(setImgTransExampleJS))
	js.Global().Set("setPatFillPolygonAngle", js.FuncOf(setPatFillPolygonAngleJS))
	js.Global().Set("setPatFillPolygonScale", js.FuncOf(setPatFillPolygonScaleJS))
	js.Global().Set("setPatFillPatternAngle", js.FuncOf(setPatFillPatternAngleJS))
	js.Global().Set("setPatFillPatternSize", js.FuncOf(setPatFillPatternSizeJS))
	js.Global().Set("setGradientFocalGamma", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientFocalGamma(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientFocalFX", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientFocalFX(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientFocalFY", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientFocalFY(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setLineThicknessFactor", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLineThicknessFactor(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setLineThicknessBlur", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLineThicknessBlur(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setLineThicknessMono", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLineThicknessMono(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setLineThicknessInvert", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLineThicknessInvert(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setCompoundWidth", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setCompoundWidth(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setCompoundAlpha1", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setCompoundAlpha1(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setCompoundAlpha2", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setCompoundAlpha2(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setCompoundAlpha3", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setCompoundAlpha3(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setCompoundAlpha4", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setCompoundAlpha4(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setCompoundInvert", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setCompoundInvert(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setImageResampleType", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageResampleType(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setImageResampleBlur", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageResampleBlur(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setImageResampleQuad", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 8 {
			setImageResampleQuad(
				args[0].Float(), args[1].Float(),
				args[2].Float(), args[3].Float(),
				args[4].Float(), args[5].Float(),
				args[6].Float(), args[7].Float(),
			)
		}
		return nil
	}))
	js.Global().Set("setPatternPerspectiveType", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setPatternPerspectiveType(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setPatternPerspectiveQuad", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 8 {
			setPatternPerspectiveQuad(
				args[0].Float(), args[1].Float(),
				args[2].Float(), args[3].Float(),
				args[4].Float(), args[5].Float(),
				args[6].Float(), args[7].Float(),
			)
		}
		return nil
	}))
	js.Global().Set("setPatternResampleType", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setPatternResampleType(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setPatternResampleGamma", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setPatternResampleGamma(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setPatternResampleBlur", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setPatternResampleBlur(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setPatternResampleQuad", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 8 {
			setPatternResampleQuad(
				args[0].Float(), args[1].Float(),
				args[2].Float(), args[3].Float(),
				args[4].Float(), args[5].Float(),
				args[6].Float(), args[7].Float(),
			)
		}
		return nil
	}))
	js.Global().Set("setImagePerspectiveType", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImagePerspectiveType(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setImagePerspectiveQuad", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 8 {
			setImagePerspectiveQuad(
				args[0].Float(), args[1].Float(),
				args[2].Float(), args[3].Float(),
				args[4].Float(), args[5].Float(),
				args[6].Float(), args[7].Float(),
			)
		}
		return nil
	}))
	js.Global().Set("setLinePatternClipScaleX", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLinePatternClipScaleX(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setLinePatternClipStartX", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLinePatternClipStartX(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setLinePatternScaleX", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLinePatternScaleX(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setLinePatternStartX", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setLinePatternStartX(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setScanlineBoolean2Mode", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setScanlineBoolean2Mode(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setScanlineBoolean2FillRule", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setScanlineBoolean2FillRule(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setScanlineBoolean2Scanline", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setScanlineBoolean2Scanline(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setScanlineBoolean2Operation", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setScanlineBoolean2Operation(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setScanlineBoolean2Center", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 2 {
			setScanlineBoolean2Center(args[0].Float(), args[1].Float())
		}
		return nil
	}))
	js.Global().Set("setGPCTestScene", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGPCTestScene(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setGPCTestOperation", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGPCTestOperation(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setGPCTestCenter", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 2 {
			setGPCTestCenter(args[0].Float(), args[1].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourPolygon", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourPolygon(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourGradient", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourGradient(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourReflect", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourReflect(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourC1", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourC1(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourC2", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourC2(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourD1", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourD1(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourD2", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourD2(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGradientsContourColors", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGradientsContourColors(args[0].Int())
		}
		return nil
	}))

	// gamma_tuner setters
	js.Global().Set("setGammaTunerR", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			gammaTunerR = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setGammaTunerG", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			gammaTunerG = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setGammaTunerB", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			gammaTunerB = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setGammaTunerGamma", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			gammaTunerGamma = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setGammaTunerPattern", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			gammaTunerPattern = args[0].Int()
		}
		return nil
	}))

	// gouraud opacity setter
	js.Global().Set("setGouraudOpacity", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			gouraudOpacity = args[0].Float()
		}
		return nil
	}))

	// bezier_div setters
	js.Global().Set("setBDAngleTol", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdAngleTolVal = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setBDApproxScale", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdApproxScaleVal = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setBDCuspLimit", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdCuspLimitVal = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setBDWidth", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdWidthVal = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setBDShowPoints", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdShowPointsVal = args[0].Bool()
		}
		return nil
	}))
	js.Global().Set("setBDShowOutline", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdShowOutlineVal = args[0].Bool()
		}
		return nil
	}))
	js.Global().Set("setBDCurveType", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdCurveTypeVal = args[0].Int()
		}
		return nil
	}))
	js.Global().Set("setBDCaseType", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdCaseTypeVal = args[0].Int()
			bdHandleCaseTypeChange()
		}
		return nil
	}))
	js.Global().Set("setBDInnerJoin", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdInnerJoinVal = args[0].Int()
		}
		return nil
	}))
	js.Global().Set("setBDLineJoin", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdLineJoinVal = args[0].Int()
		}
		return nil
	}))
	js.Global().Set("setBDLineCap", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bdLineCapVal = args[0].Int()
		}
		return nil
	}))
	js.Global().Set("getBDWidth", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		return bdWidthVal
	}))

	// rasterizers setters
	js.Global().Set("setRasterizersGamma", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			rasterizersGamma = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setRasterizersAlpha", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			rasterizersAlpha = args[0].Float()
		}
		return nil
	}))

	// bspline setters
	js.Global().Set("setBSplineNumPoints", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bsplineNumPoints = args[0].Float()
		}
		return nil
	}))
	js.Global().Set("setBSplineClosed", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			bsplineClosed = args[0].Bool()
		}
		return nil
	}))

	// flash_rasterizer2 setters
	js.Global().Set("setFlash2ShapeIdx", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setFlash2ShapeIdx(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("applyFlash2Wheel", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) >= 3 {
			applyFlash2Wheel(args[0].Float(), args[1].Float(), args[2].Float())
		}
		return nil
	}))

	// Keep the Go program running
	select {}
}

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
	if demoType == "bspline" {
		return handleBSplineMouseDown(x, y)
	}
	if demoType == "interactive_polygon" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleInteractivePolygonMouseDown(x, y)
	}
	if demoType == "conv_dash_marker" {
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
	if demoType == "lion" {
		right := len(args) >= 4 && args[3].Bool()
		return handleLionMouseDown(x, y, right)
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
	if demoType == "polymorphic_renderer" {
		return handlePolyRenMouseDown(x, y)
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
	if demoType == "distortions" {
		return handleDistortionsMouseDown(x, y)
	}
	if demoType == "trans_polar" {
		return handleTransPolarMouseDown(x, y)
	}
	if demoType == "trans_curve2" {
		return handleTransCurve2MouseDown(x, y)
	}
	if demoType == "gamma_ctrl" {
		return handleGammaCtrlMouseDown(x, y)
	}
	// gamma_tuner no longer has canvas-based widgets
	if demoType == "lion_lens" {
		return handleLionLensMouseDown(x, y)
	}
	if demoType == "circles" {
		generateCircles()
		return true
	}
	if demoType == "simple_blur" {
		simpleBlurCX = x
		simpleBlurCY = y
		return true
	}
	if demoType == "alpha_mask" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return handleAlphaMaskRightMouseDown(x, y)
		}
		return handleAlphaMaskMouseDown(x, y, 0)
	}
	if demoType == "alpha_mask2" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return handleAlphaMask2RightMouseDown(x, y)
		}
		return handleAlphaMask2MouseDown(x, y, 0)
	}
	if demoType == "multi_clip" {
		return handleMultiClipMouseDown(x, y)
	}
	if demoType == "image_transforms" {
		return handleImgTransMouseDown(x, y)
	}
	if demoType == "image_resample" {
		return handleImageResampleMouseDown(x, y)
	}
	if demoType == "image_perspective" {
		return handleImagePerspectiveMouseDown(x, y)
	}
	if demoType == "pattern_perspective" {
		return handlePatternPerspectiveMouseDown(x, y)
	}
	if demoType == "pattern_resample" {
		return handlePatternResampleMouseDown(x, y)
	}
	if demoType == "scanline_boolean2" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleScanlineBoolean2MouseDown(x, y)
	}
	if demoType == "gpc_test" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleGPCTestMouseDown(x, y)
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
	if demoType == "bspline" {
		return handleBSplineMouseMove(x, y)
	}
	if demoType == "interactive_polygon" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleInteractivePolygonMouseMove(x, y)
	}
	if demoType == "conv_dash_marker" {
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
	if demoType == "lion" {
		right := len(args) >= 4 && args[3].Bool()
		return handleLionMouseMove(x, y, right)
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
	if demoType == "polymorphic_renderer" {
		return handlePolyRenMouseMove(x, y)
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
	if demoType == "distortions" {
		return handleDistortionsMouseMove(x, y)
	}
	if demoType == "trans_polar" {
		return handleTransPolarMouseMove(x, y)
	}
	if demoType == "trans_curve2" {
		return handleTransCurve2MouseMove(x, y)
	}
	if demoType == "gamma_ctrl" {
		return handleGammaCtrlMouseMove(x, y)
	}
	// gamma_tuner no longer has canvas-based widgets
	if demoType == "lion_lens" {
		return handleLionLensMouseMove(x, y)
	}
	if demoType == "simple_blur" {
		simpleBlurCX = x
		simpleBlurCY = y
		return true
	}
	if demoType == "alpha_mask" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return handleAlphaMaskRightMouseDown(x, y)
		}
		return handleAlphaMaskMouseDown(x, y, 0)
	}
	if demoType == "alpha_mask2" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return handleAlphaMask2RightMouseDown(x, y)
		}
		return handleAlphaMask2MouseDown(x, y, 0)
	}
	if demoType == "multi_clip" {
		return handleMultiClipMouseDown(x, y)
	}
	if demoType == "image_transforms" {
		return handleImgTransMouseMove(x, y)
	}
	if demoType == "image_resample" {
		return handleImageResampleMouseMove(x, y)
	}
	if demoType == "image_perspective" {
		return handleImagePerspectiveMouseMove(x, y)
	}
	if demoType == "pattern_perspective" {
		return handlePatternPerspectiveMouseMove(x, y)
	}
	if demoType == "pattern_resample" {
		return handlePatternResampleMouseMove(x, y)
	}
	if demoType == "scanline_boolean2" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleScanlineBoolean2MouseMove(x, y)
	}
	if demoType == "gpc_test" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleGPCTestMouseMove(x, y)
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
	if demoType == "bspline" {
		handleBSplineMouseUp()
	}
	if demoType == "interactive_polygon" {
		handleInteractivePolygonMouseUp()
	}
	if demoType == "conv_dash_marker" {
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
	if demoType == "lion" {
		handleLionMouseUp()
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
	if demoType == "polymorphic_renderer" {
		handlePolyRenMouseUp()
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
	if demoType == "distortions" {
		handleDistortionsMouseUp()
	}
	if demoType == "trans_polar" {
		handleTransPolarMouseUp()
	}
	if demoType == "trans_curve2" {
		handleTransCurve2MouseUp()
	}
	if demoType == "gamma_ctrl" {
		handleGammaCtrlMouseUp()
	}
	// gamma_tuner no longer has canvas-based widgets
	if demoType == "lion_lens" {
		handleLionLensMouseUp()
	}
	if demoType == "image_resample" {
		handleImageResampleMouseUp()
	}
	if demoType == "image_perspective" {
		handleImagePerspectiveMouseUp()
	}
	if demoType == "pattern_perspective" {
		handlePatternPerspectiveMouseUp()
	}
	if demoType == "pattern_resample" {
		handlePatternResampleMouseUp()
	}
	if demoType == "scanline_boolean2" {
		handleScanlineBoolean2MouseUp()
	}
	if demoType == "gpc_test" {
		handleGPCTestMouseUp()
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
		"x0": dashX[0], "y0": dashY[0],
		"x1": dashX[1], "y1": dashY[1],
		"x2": dashX[2], "y2": dashY[2],
	}
}

func setDashNodes(this js.Value, args []js.Value) interface{} {
	if len(args) >= 6 {
		dashX[0] = args[0].Float()
		dashY[0] = args[1].Float()
		dashX[1] = args[2].Float()
		dashY[1] = args[3].Float()
		dashX[2] = args[4].Float()
		dashY[2] = args[5].Float()
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
	if demoType != "lion" && demoType != "lionoutline" && demoType != "lion_lens" {
		lionPaths = nil
	}
	if demoType != "imagefilters" {
		testImage = nil
	}

	ctx.Clear(agg.White)
	ctx.GetAgg2D().ResetStyle()

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
	case "graph_test":
		drawGraphTestDemo()
	case "bspline":
		drawBSplineDemo()
	case "interactive_polygon":
		drawInteractivePolygonDemo()
	case "conv_dash_marker":
		drawDashDemo()
	case "gouraud":
		drawGouraudDemo()
	case "imagefilters":
		drawImageFiltersDemo()
	case "image_fltr_graph":
		drawImageFltrGraphDemo()
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
	case "polymorphic_renderer":
		drawPolymorphicRendererDemo()
	case "flash_rasterizer":
		drawFlashRasterizerDemo()
	case "flash_rasterizer2":
		drawFlashRasterizer2Demo()
	case "perspective":
		drawPerspectiveDemo()
	case "bezier_div":
		drawBezierDivDemo()
	case "gouraud_mesh":
		drawGouraudMeshDemo()
	case "trans_curve":
		drawTransCurveDemo()
	case "distortions":
		drawDistortionsDemo()
	case "trans_polar":
		drawTransPolarDemo()
	case "trans_curve2":
		drawTransCurve2Demo()
	case "gamma_ctrl":
		drawGammaCtrlDemo()
	case "gamma_tuner":
		drawGammaTunerDemo()
	case "lion_lens":
		drawLionLensDemo()
	case "circles":
		drawCirclesScatterDemo()
	case "blur":
		drawBlurDemo()
	case "simple_blur":
		drawSimpleBlurDemo()
	case "alpha_mask":
		drawAlphaMaskDemo()
	case "alpha_mask2":
		drawAlphaMask2Demo()
	case "alpha_mask3":
		drawAlphaMask3Demo()
	case "compositing":
		drawCompositingDemo()
	case "compositing2":
		drawCompositing2Demo()
	case "multi_clip":
		drawMultiClipDemo()
	case "image1":
		drawImage1Demo()
	case "image_transforms":
		drawImageTransformsDemo()
	case "image_alpha":
		drawImageAlphaDemo()
	case "pattern_fill":
		drawPatternFillDemo()
	case "raster_text":
		drawRasterTextDemo()
	case "gradient_focal":
		drawGradientFocalDemo()
	case "line_thickness":
		drawLineThicknessDemo()
	case "rasterizer_compound":
		drawRasterizerCompoundDemo()
	case "image_resample":
		drawImageResampleDemo()
	case "pattern_perspective":
		drawPatternPerspectiveDemo()
	case "pattern_resample":
		drawPatternResampleDemo()
	case "image_perspective":
		drawImagePerspectiveDemo()
	case "line_patterns_clip":
		drawLinePatternsClipDemo()
	case "line_patterns":
		drawLinePatternsDemo()
	case "scanline_boolean2":
		drawScanlineBoolean2Demo()
	case "gpc_test":
		drawGPCTestDemo()
	case "gradients_contour":
		drawGradientsContourDemo()
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

func toggleTransCurve2AnimateJS(this js.Value, args []js.Value) interface{} {
	toggleTransCurve2Animate()
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

func drawHandle(x, y float64) {
	ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
	ctx.FillCircle(x, y, 5)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(x, y, 5)
}
