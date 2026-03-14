package main

import (
	"os"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
)

func BenchmarkDemos(b *testing.B) {
	width, height = 800, 600
	ctx = agg.NewContext(width, height)
	canvasBuf = ctx.GetImage().Data

	demos := []string{
		"agg2d", "lion", "gradients", "aa", "blend", "interactive_polygon", "graph_test",
		"bspline", "conv_dash_marker", "gouraud", "imagefilters", "image_fltr_graph", "image_filters2",
		"sbool", "aatest", "convstroke", "convcontour", "gamma", "lionoutline",
		"roundedrect", "component", "alphagrad",
		"rasterizers", "flash_rasterizer", "perspective", "bezier_div",
		"gouraud_mesh", "trans_curve", "distortions", "trans_polar",
		"trans_curve2", "circles", "blur", "simple_blur",
		"gamma_ctrl", "gamma_tuner", "lion_lens", "gradient_focal", "line_thickness", "rasterizer_compound", "image_resample", "line_patterns_clip", "line_patterns", "scanline_boolean2", "gpc_test",
		"pattern_perspective", "pattern_resample", "image_perspective",
	}

	for _, demo := range demos {
		b.Run(demo, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				renderDemoForBenchmark(demo)
			}
		})
	}
}

func renderDemoForBenchmark(demoType string) {
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
	case "distortions":
		drawDistortionsDemo()
	case "trans_polar":
		drawTransPolarDemo()
	case "trans_curve2":
		drawTransCurve2Demo()
	case "circles":
		drawCirclesScatterDemo()
	case "blur":
		drawBlurDemo()
	case "simple_blur":
		drawSimpleBlurDemo()
	case "gamma_ctrl":
		drawGammaCtrlDemo()
	case "gamma_tuner":
		drawGammaTunerDemo()
	case "lion_lens":
		drawLionLensDemo()
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
	}
}

func TestMain(m *testing.M) {
	// Initialize things needed for demos
	// We might need to mock some JS things if any, but main_stub.go should be fine.
	os.Exit(m.Run())
}
