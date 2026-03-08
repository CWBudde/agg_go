package main

import (
	"fmt"
	"os"
	"path/filepath"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	agg2d := ctx.GetAgg2D()

	// Clear background to white
	agg2d.ClearAll(agg.White)

	// Try to load a system font (this will only work if FreeType is available)
	err := loadSystemFont(agg2d)
	if err != nil {
		fmt.Printf("Warning: Could not load font: %v\n", err)
		fmt.Println("Text rendering will be limited without font support")
		fmt.Println("To enable full text support, build with: go build -tags freetype")
	}

	width := ctx.GetImage().Width()
	height := ctx.GetImage().Height()

	// Render text examples
	renderTextExamples(agg2d, width, height)
}

// loadSystemFont attempts to load a system font for demonstration.
// This function will work when built with FreeType support.
func loadSystemFont(agg2d *agg.Agg2D) error {
	// Common system font paths
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",                 // Linux
		"/System/Library/Fonts/Arial.ttf",                                 // macOS
		"C:\\Windows\\Fonts\\arial.ttf",                                   // Windows
		"/usr/share/fonts/TTF/DejaVuSans.ttf",                             // Arch Linux
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf", // Linux
	}

	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err != nil {
			continue
		}

		fmt.Printf("Attempting to load font: %s\n", fontPath)
		err := agg2d.Font(fontPath, 16.0, false, false, agg.RasterFontCache, 0.0)
		if err != nil {
			fmt.Printf("Failed to load %s: %v\n", fontPath, err)
			continue
		}
		fmt.Printf("Successfully loaded font: %s\n", filepath.Base(fontPath))
		return nil
	}

	return fmt.Errorf("no suitable fonts found in common system locations")
}

// renderTextExamples renders various text examples to demonstrate functionality.
func renderTextExamples(agg2d *agg.Agg2D, width, height int) {
	// Set text color to black
	agg2d.FillColor(agg.Black)

	// Title
	agg2d.TextAlignment(agg.AlignCenter, agg.AlignTop)
	agg2d.Text(float64(width/2), 50, "AGG2D Text Rendering Examples", false, 0, 0)

	// Test different alignments
	renderAlignmentExamples(agg2d, width, height)

	// Test text with different colors
	renderColorExamples(agg2d, width, height)

	// Test text measurement
	renderMeasurementExamples(agg2d, width, height)

	// Test special characters and Unicode
	renderUnicodeExamples(agg2d, width, height)
}

// renderAlignmentExamples demonstrates different text alignment options.
func renderAlignmentExamples(agg2d *agg.Agg2D, width, height int) {
	// Draw alignment guide lines
	drawAlignmentGuides(agg2d, width, height)

	agg2d.FillColor(agg.Black)

	centerX := float64(width / 2)
	centerY := float64(height / 2)

	// Test all alignment combinations
	alignments := []struct {
		x, y   float64
		alignX agg.TextAlignment
		alignY agg.TextAlignment
		text   string
	}{
		{100, 150, agg.AlignLeft, agg.AlignTop, "Left-Top"},
		{centerX, 150, agg.AlignCenter, agg.AlignTop, "Center-Top"},
		{float64(width - 100), 150, agg.AlignRight, agg.AlignTop, "Right-Top"},

		{100, centerY, agg.AlignLeft, agg.AlignCenter, "Left-Center"},
		{centerX, centerY, agg.AlignCenter, agg.AlignCenter, "Center-Center"},
		{float64(width - 100), centerY, agg.AlignRight, agg.AlignCenter, "Right-Center"},

		{100, float64(height - 100), agg.AlignLeft, agg.AlignBottom, "Left-Bottom"},
		{centerX, float64(height - 100), agg.AlignCenter, agg.AlignBottom, "Center-Bottom"},
		{float64(width - 100), float64(height - 100), agg.AlignRight, agg.AlignBottom, "Right-Bottom"},
	}

	for _, align := range alignments {
		agg2d.TextAlignment(align.alignX, align.alignY)
		agg2d.Text(align.x, align.y, align.text, false, 0, 0)
	}
}

// drawAlignmentGuides draws helper lines to visualize alignment.
func drawAlignmentGuides(agg2d *agg.Agg2D, width, height int) {
	// Set line color to light gray
	agg2d.LineColor(agg.Color{R: 200, G: 200, B: 200, A: 255})
	agg2d.LineWidth(1.0)

	centerX := float64(width / 2)
	centerY := float64(height / 2)

	// Vertical guides
	agg2d.Line(100, 100, 100, float64(height-100))                               // Left
	agg2d.Line(centerX, 100, centerX, float64(height-100))                       // Center
	agg2d.Line(float64(width-100), 100, float64(width-100), float64(height-100)) // Right

	// Horizontal guides
	agg2d.Line(50, 150, float64(width-50), 150)                                 // Top
	agg2d.Line(50, centerY, float64(width-50), centerY)                         // Center
	agg2d.Line(50, float64(height-100), float64(width-50), float64(height-100)) // Bottom
}

// renderColorExamples demonstrates text rendering with different colors.
func renderColorExamples(agg2d *agg.Agg2D, width, height int) {
	agg2d.TextAlignment(agg.AlignLeft, agg.AlignTop)

	colors := []struct {
		color agg.Color
		name  string
	}{
		{agg.Red, "Red Text"},
		{agg.Green, "Green Text"},
		{agg.Blue, "Blue Text"},
		{agg.Color{R: 255, G: 165, B: 0, A: 255}, "Orange Text"},
		{agg.Color{R: 128, G: 0, B: 128, A: 255}, "Purple Text"},
	}

	startY := 100.0
	for i, colorExample := range colors {
		agg2d.FillColor(colorExample.color)
		agg2d.Text(50, startY+float64(i*30), colorExample.name, false, 0, 0)
	}
}

// renderMeasurementExamples demonstrates text width measurement.
func renderMeasurementExamples(agg2d *agg.Agg2D, width, height int) {
	agg2d.FillColor(agg.Black)
	agg2d.TextAlignment(agg.AlignLeft, agg.AlignTop)

	testText := "Measured Text"
	x, y := 50.0, 300.0

	// Render the text
	agg2d.Text(x, y, testText, false, 0, 0)

	// Show text width (even if measurement returns 0 without font engine)
	textWidth := agg2d.TextWidth(testText)
	widthText := fmt.Sprintf("Width: %.1f pixels", textWidth)
	agg2d.Text(x, y+25, widthText, false, 0, 0)

	// Draw a rectangle showing the expected text bounds
	if textWidth > 0 {
		agg2d.LineColor(agg.Red)
		agg2d.LineWidth(1.0)
		agg2d.Rectangle(x, y-20, x+textWidth, y+5)
	}
}

// renderUnicodeExamples demonstrates Unicode text rendering.
func renderUnicodeExamples(agg2d *agg.Agg2D, width, height int) {
	agg2d.FillColor(agg.Black)
	agg2d.TextAlignment(agg.AlignLeft, agg.AlignTop)

	unicodeExamples := []string{
		"English: Hello World!",
		"Français: Bonjour le monde!",
		"Deutsch: Hallo Welt!",
		"Español: ¡Hola Mundo!",
		"Русский: Привет мир!",
		"日本語: こんにちは世界!",
		"中文: 你好世界!",
		"العربية: مرحبا بالعالم!",
		"Symbols: ★★★ ♦♦♦ ●●●",
		"Emoji: 🌍 🎉 💖 🚀 🎨",
	}

	startY := 380.0
	for i, example := range unicodeExamples {
		agg2d.Text(50, startY+float64(i*20), example, false, 0, 0)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Text Rendering",
		Width:  800,
		Height: 600,
	}, &demo{})
}
