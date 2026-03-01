package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

func TestTmpTransformedStrokeDebug(t *testing.T) {
	width, height := 120, 120
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255})
	ctx.LineColor(agg2d.Color{255, 0, 255, 255})
	ctx.LineWidth(4.0)
	ctx.LineJoin(agg2d.JoinMiter)
	ctx.LineCap(agg2d.CapButt)
	ctx.ResetPath()
	addRectPath(ctx, 55, 65, 95, 85)
	ctx.DrawPath(agg2d.StrokeOnly)

	found := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := getPixel(buffer, stride, x, y)
			if pixel != [4]uint8{255, 255, 255, 255} {
				found = true
				t.Logf("pixel at %d,%d = %v", x, y, pixel)
				return
			}
		}
	}

	if !found {
		t.Fatal("no transformed stroke pixels found")
	}
}
