//go:build !js || !wasm
// +build !js !wasm

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	agg "agg_go"
)

var (
	width, height = 800, 600
	ctx           *agg.Context
	canvasBuf     []uint8
	lionPaths     []LionPath
)

// logStatus prints a status message to stdout (replaces the JS DOM update in main.go).
func logStatus(msg string) {
	fmt.Println(msg)
}

func main() {
	demos := os.Args[1:]
	if len(demos) == 0 {
		demos = []string{
			"agg2d",
			"lion", "gradients", "aa", "blend",
			"bspline", "conv_dash_marker", "gouraud", "imagefilters",
			"sbool", "aatest", "convstroke", "convcontour", "gamma", "lionoutline",
			"roundedrect", "component", "alphagrad",
			"rasterizers", "flash_rasterizer", "perspective", "bezier_div",
			"gouraud_mesh", "trans_curve", "distortions", "trans_polar",
			"trans_curve2", "gamma_ctrl", "gamma_tuner", "lion_lens", "circles", "blur", "simple_blur",
		}
	}

	ctx = agg.NewContext(width, height)
	canvasBuf = ctx.GetImage().Data

	outDir := "."
	if dir := os.Getenv("AGG_OUT"); dir != "" {
		outDir = dir
	}

	var failed []string
	for _, demo := range demos {
		start := time.Now()
		if err := renderDemoToFile(demo, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "error rendering %s: %v\n", demo, err)
			failed = append(failed, demo)
		} else {
			fmt.Printf("Rendered %s in %v\n", demo, time.Since(start))
		}
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "failed: %s\n", strings.Join(failed, ", "))
		os.Exit(1)
	}
}

func renderDemoToFile(demoType, outDir string) error {
	if demoType != "lion" && demoType != "lionoutline" {
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
	case "bspline":
		drawBSplineDemo()
	case "conv_dash_marker":
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
	default:
		return fmt.Errorf("unknown demo: %q", demoType)
	}

	out := filepath.Join(outDir, demoType+".png")
	if err := ctx.GetImage().SaveToPNG(out); err != nil {
		return err
	}
	fmt.Printf("saved: %s\n", out)
	return nil
}

func drawHandle(x, y float64) {
	ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
	ctx.FillCircle(x, y, 5)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(x, y, 5)
}
