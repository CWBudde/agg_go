//go:build js && wasm
// +build js,wasm

package main

import "syscall/js"

// Mouse event entry points exported to JavaScript. They dispatch to the
// per-demo handlers based on the active demo type.

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
	if demoType == "blend_color" {
		return handleBlendColorMouseDown(x, y)
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
	if demoType == "mol_view" {
		right := len(args) >= 4 && args[3].Bool()
		return handleMolViewMouseDown(x, y, right)
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
	if demoType == "line_patterns_clip" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleLinePatternsClipMouseDown(x, y)
	}
	if demoType == "line_patterns" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleLinePatternsMouseDown(x, y)
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
	if demoType == "gradient_focal" {
		return handleGradientFocalMouseDown(x, y)
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
	if demoType == "blend_color" {
		return handleBlendColorMouseMove(x, y)
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
	if demoType == "mol_view" {
		right := len(args) >= 4 && args[3].Bool()
		return handleMolViewMouseMove(x, y, right)
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
	if demoType == "line_patterns_clip" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleLinePatternsClipMouseMove(x, y)
	}
	if demoType == "line_patterns" {
		right := len(args) >= 4 && args[3].Bool()
		if right {
			return false
		}
		return handleLinePatternsMouseMove(x, y)
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
	if demoType == "gradient_focal" {
		return handleGradientFocalMouseMove(x, y)
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
	if demoType == "blend_color" {
		handleBlendColorMouseUp()
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
	if demoType == "mol_view" {
		handleMolViewMouseUp()
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
	if demoType == "line_patterns_clip" {
		handleLinePatternsClipMouseUp()
	}
	if demoType == "line_patterns" {
		handleLinePatternsMouseUp()
	}
	if demoType == "scanline_boolean2" {
		handleScanlineBoolean2MouseUp()
	}
	if demoType == "gpc_test" {
		handleGPCTestMouseUp()
	}
	if demoType == "gradient_focal" {
		handleGradientFocalMouseUp()
	}
	return nil
}
