package benchmark

import (
	"math"
	"runtime"
	"testing"

	"agg_go/internal/agg2d"
)

var (
	benchWhite = agg2d.Color{255, 255, 255, 255}
	benchBlack = agg2d.Color{0, 0, 0, 255}
)

func BenchmarkAgg2DSceneSolidFill(b *testing.B) {
	benchAgg2DScene(b, "256x256", 256, 256, renderSolidFillScene)
	benchAgg2DScene(b, "800x600", 800, 600, renderSolidFillScene)
}

func BenchmarkAgg2DSceneStrokeAndCurves(b *testing.B) {
	benchAgg2DScene(b, "256x256", 256, 256, renderStrokeAndCurveScene)
	benchAgg2DScene(b, "800x600", 800, 600, renderStrokeAndCurveScene)
}

func BenchmarkAgg2DSceneGradientClip(b *testing.B) {
	benchAgg2DScene(b, "256x256", 256, 256, renderGradientClipScene)
	benchAgg2DScene(b, "800x600", 800, 600, renderGradientClipScene)
}

func benchAgg2DScene(b *testing.B, name string, width, height int, draw func(ctx *agg2d.Agg2D, width, height int)) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		stride := width * 4
		buffer := make([]uint8, height*stride)

		ctx := agg2d.NewAgg2D()
		ctx.Attach(buffer, width, height, stride)

		b.ReportAllocs()
		b.SetBytes(int64(len(buffer)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx.ClearAll(benchWhite)
			draw(ctx, width, height)
		}

		// Keep the renderer output live so benchmarked work is not elided.
		runtime.KeepAlive(buffer)
	})
}

func renderSolidFillScene(ctx *agg2d.Agg2D, width, height int) {
	cols := 16
	rows := 12
	cellW := float64(width) / float64(cols)
	cellH := float64(height) / float64(rows)

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			r := uint8((x * 255) / (cols - 1))
			g := uint8((y * 255) / (rows - 1))
			b := uint8(((x + y) * 255) / (cols + rows - 2))
			ctx.FillColor(agg2d.Color{r, g, b, 200})

			x0 := float64(x) * cellW
			y0 := float64(y) * cellH
			x1 := x0 + cellW*0.9
			y1 := y0 + cellH*0.9
			drawRectPath(ctx, x0, y0, x1, y1)
		}
	}
}

func renderStrokeAndCurveScene(ctx *agg2d.Agg2D, width, height int) {
	centerX := float64(width) * 0.5
	centerY := float64(height) * 0.5
	baseRadius := math.Min(float64(width), float64(height)) * 0.35

	ctx.LineColor(benchBlack)
	ctx.LineWidth(1.75)

	for i := 0; i < 24; i++ {
		a0 := float64(i) * (2 * math.Pi / 24)
		a1 := a0 + math.Pi/10
		a2 := a1 + math.Pi/10

		x0 := centerX + baseRadius*math.Cos(a0)
		y0 := centerY + baseRadius*math.Sin(a0)
		x1 := centerX + (baseRadius*0.55)*math.Cos(a1)
		y1 := centerY + (baseRadius*0.55)*math.Sin(a1)
		x2 := centerX + (baseRadius*0.95)*math.Cos(a2)
		y2 := centerY + (baseRadius*0.95)*math.Sin(a2)

		ctx.ResetPath()
		ctx.MoveTo(x0, y0)
		ctx.QuadricCurveTo(x1, y1, x2, y2)
		ctx.DrawPath(agg2d.StrokeOnly)
	}

	ctx.FillColor(agg2d.Color{20, 80, 180, 160})
	ctx.ResetPath()
	for i := 0; i < 10; i++ {
		angle := float64(i) * (2 * math.Pi / 10)
		radius := baseRadius
		if i%2 == 1 {
			radius *= 0.45
		}
		x := centerX + radius*math.Cos(angle-math.Pi/2)
		y := centerY + radius*math.Sin(angle-math.Pi/2)
		if i == 0 {
			ctx.MoveTo(x, y)
		} else {
			ctx.LineTo(x, y)
		}
	}
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.FillOnly)
}

func renderGradientClipScene(ctx *agg2d.Agg2D, width, height int) {
	clipPadX := float64(width) * 0.12
	clipPadY := float64(height) * 0.12
	ctx.ClipBox(clipPadX, clipPadY, float64(width)-clipPadX, float64(height)-clipPadY)

	ctx.FillLinearGradient(
		0, 0,
		float64(width), float64(height),
		agg2d.Color{240, 60, 40, 255},
		agg2d.Color{40, 90, 220, 255},
		1.0,
	)

	for i := 0; i < 20; i++ {
		offset := float64(i) * 0.03
		x0 := float64(width) * (0.05 + offset)
		y0 := float64(height) * (0.05 + offset)
		x1 := float64(width) * (0.95 - offset)
		y1 := float64(height) * (0.95 - offset)
		drawRectPath(ctx, x0, y0, x1, y1)
	}

	ctx.ResetTransformations()
	ctx.ClipBox(0, 0, float64(width), float64(height))
}

func drawRectPath(ctx *agg2d.Agg2D, x0, y0, x1, y1 float64) {
	ctx.ResetPath()
	ctx.MoveTo(x0, y0)
	ctx.LineTo(x1, y0)
	ctx.LineTo(x1, y1)
	ctx.LineTo(x0, y1)
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.FillOnly)
}
