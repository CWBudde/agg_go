//go:build !js || !wasm
// +build !js !wasm

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agg "agg_go"
)

var (
	width, height = 800, 600
	ctx           *agg.Context
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
			"lines", "circles", "starburst", "rects",
			"lion", "gradients", "aa", "blend",
			"bspline", "dash", "gouraud", "imagefilters",
			"sbool", "aatest", "convstroke", "convcontour", "gamma",
		}
	}

	ctx = agg.NewContext(width, height)
	canvasBuf := ctx.GetImage().Data
	_ = canvasBuf

	outDir := "."
	if dir := os.Getenv("AGG_OUT"); dir != "" {
		outDir = dir
	}

	var failed []string
	for _, demo := range demos {
		if err := renderDemoToFile(demo, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "error rendering %s: %v\n", demo, err)
			failed = append(failed, demo)
		}
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "failed: %s\n", strings.Join(failed, ", "))
		os.Exit(1)
	}
}

func renderDemoToFile(demoType, outDir string) error {
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
	case "convstroke":
		drawConvStrokeDemo()
	case "convcontour":
		drawConvContourDemo()
	case "gamma":
		drawGammaCorrectionDemo()
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
