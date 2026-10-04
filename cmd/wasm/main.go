//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	agg "github.com/cwbudde/agg_go"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
)

var (
	width, height = 800, 600
	ctx           *agg.Context
	canvasBuf     []uint8
	lionData      *liondemo.LionData
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
	js.Global().Set("setImageFilters2Filter", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageFilters2Filter(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setImageFilters2Radius", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageFilters2Radius(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setImageFilters2Normalize", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setImageFilters2Normalize(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setIdeaRotate", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setIdeaRotate(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setIdeaEvenOdd", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setIdeaEvenOdd(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setIdeaDraft", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setIdeaDraft(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setIdeaRoundoff", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setIdeaRoundoff(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setIdeaAngleDelta", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setIdeaAngleDelta(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setMolViewMolecule", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setMolViewMolecule(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setMolViewThickness", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setMolViewThickness(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setMolViewTextSize", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setMolViewTextSize(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setMolViewAutoRotate", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setMolViewAutoRotate(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setSBoolOp", js.FuncOf(setSBoolOp))
	js.Global().Set("setSBoolMul1", js.FuncOf(setSBoolMul1))
	js.Global().Set("setSBoolMul2", js.FuncOf(setSBoolMul2))
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
	js.Global().Set("setTransCurveNumPoints", js.FuncOf(setTransCurveNumPointsJS))
	js.Global().Set("setTransCurveClose", js.FuncOf(setTransCurveCloseJS))
	js.Global().Set("setTransCurvePreserveXScale", js.FuncOf(setTransCurvePreserveXScaleJS))
	js.Global().Set("setTransCurveFixedLen", js.FuncOf(setTransCurveFixedLenJS))
	js.Global().Set("toggleTransCurveAnimate", js.FuncOf(toggleTransCurveAnimateJS))
	js.Global().Set("toggleTransCurve2Animate", js.FuncOf(toggleTransCurve2AnimateJS))
	js.Global().Set("setTransCurve2NumPoints", js.FuncOf(setTransCurve2NumPointsJS))
	js.Global().Set("setTransCurve2FixedLen", js.FuncOf(setTransCurve2FixedLenJS))
	js.Global().Set("setTransCurve2PreserveXScale", js.FuncOf(setTransCurve2PreserveXScaleJS))
	js.Global().Set("setBlurRadius", js.FuncOf(setBlurRadius))
	js.Global().Set("setBlurMethod", js.FuncOf(setBlurMethod))
	js.Global().Set("setBlendColorMethod", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		blendColorMethod = args[0].Int()
		return nil
	}))
	js.Global().Set("setBlendColorRadius", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		blendColorRadius = args[0].Float()
		return nil
	}))
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
	js.Global().Set("getGradientFocalFX", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		return gradientFocalFX
	}))
	js.Global().Set("getGradientFocalFY", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		return gradientFocalFY
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
	js.Global().Set("getLineThicknessBlurTime", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		return getLineThicknessBlurTime()
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
	js.Global().Set("setGraphTestMode", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGraphTestMode(args[0].Int())
		}
		return nil
	}))
	js.Global().Set("setGraphTestWidth", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGraphTestWidth(args[0].Float())
		}
		return nil
	}))
	js.Global().Set("setGraphTestTranslucent", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGraphTestTranslucent(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setGraphTestDrawNodes", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGraphTestDrawNodes(args[0].Bool())
		}
		return nil
	}))
	js.Global().Set("setGraphTestDrawEdges", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setGraphTestDrawEdges(args[0].Bool())
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
	js.Global().Set("getLinePatternClipNodesEncoded", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		return encodeLinePatternClipPoints()
	}))
	js.Global().Set("setLinePatternClipNodesEncoded", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			return setLinePatternClipPointsEncoded(args[0].String())
		}
		return false
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
	js.Global().Set("getLinePatternNodesEncoded", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		return encodeLinePatternCurves()
	}))
	js.Global().Set("setLinePatternNodesEncoded", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			return setLinePatternCurvesEncoded(args[0].String())
		}
		return false
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
		lionData = nil
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
	case "image_filters2":
		drawImageFilters2Demo()
	case "idea":
		drawIdeaDemo()
	case "mol_view":
		drawMolViewDemo()
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
	case "blend_color":
		drawBlendColorDemo()
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

func drawHandle(x, y float64) {
	ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
	ctx.FillCircle(x, y, 5)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(x, y, 5)
}
