package graphtest

import (
	"image/color"
	"testing"

	agg "github.com/cwbudde/agg_go"
	internalcolor "github.com/cwbudde/agg_go/internal/color"
)

func TestGraphTypeControlMatchesCPPVisualOrder(t *testing.T) {
	ctrl := newTypeControl(Config{Mode: 0})

	if ctrl.X1() != -1 || ctrl.Y1() != -1 || ctrl.X2() != -1 || ctrl.Y2() != -1 {
		t.Fatalf("type control bounds = (%v,%v,%v,%v), want C++ degenerate panel bounds (-1,-1,-1,-1)",
			ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())
	}

	want := []string{
		"Solid lines",
		"Bezier curves",
		"Dashed curves",
		"Poygons AA",
		"Poygons Bin",
	}
	if got := int(ctrl.NumItems()); got != len(want) {
		t.Fatalf("NumItems() = %d, want %d", got, len(want))
	}
	for i, text := range want {
		if got := ctrl.ItemText(i); got != text {
			t.Fatalf("ItemText(%d) = %q, want %q", i, got, text)
		}
	}
	if got := ctrl.CurItem(); got != 0 {
		t.Fatalf("CurItem() = %d, want 0", got)
	}
}

func TestGraphArrowsRenderOnTopOfNodes(t *testing.T) {
	g := &Graph{
		nodes: []node{
			{x: 0.2, y: 0.5},
			{x: 0.8, y: 0.5},
		},
		edges:    []edge{{n1: 0, n2: 1}},
		prepared: make(map[preparedKey]*preparedGraph),
	}

	img := agg.NewImage(make([]byte, 100*100*4), 100, 100, 100*4)
	Draw(agg.NewContextForImage(img), g, Config{
		Mode:      0,
		Width:     2,
		DrawNodes: true,
		DrawEdges: true,
	})

	got := img.ToGoImage().RGBAAt(73, 50)
	if !isDarkArrowPixel(got) {
		t.Fatalf("arrow sample at destination node = rgba(%d,%d,%d,%d), want dark edge/arrow color on top of node gradient",
			got.R, got.G, got.B, got.A)
	}
}

func TestGraphEdgeColorsMatchCPPsRGBConversion(t *testing.T) {
	g := &Graph{
		nodes: []node{
			{x: 0.2, y: 0.5},
			{x: 0.8, y: 0.5},
		},
		edges:    []edge{{n1: 0, n2: 1}},
		prepared: make(map[preparedKey]*preparedGraph),
	}

	img := agg.NewImage(make([]byte, 100*100*4), 100, 100, 100*4)
	Draw(agg.NewContextForImage(img), g, Config{
		Mode:      0,
		Width:     2,
		DrawNodes: false,
		DrawEdges: true,
	})

	rng := newClibcRandSeed(100)
	srgb := internalcolor.RGBA8[internalcolor.SRGB]{
		R: uint8(rng.randN(128)),
		G: uint8(rng.randN(128)),
		B: uint8(rng.randN(128)),
		A: 255,
	}
	want := internalcolor.ConvertRGBA8SRGBToLinear(srgb)

	got := img.ToGoImage().RGBAAt(50, 50)
	if got.R != want.R || got.G != want.G || got.B != want.B || got.A != want.A {
		t.Fatalf("edge sample = rgba(%d,%d,%d,%d), want C++ srgba8 converted to linear rgba(%d,%d,%d,%d)",
			got.R, got.G, got.B, got.A, want.R, want.G, want.B, want.A)
	}
}

func isDarkArrowPixel(c color.RGBA) bool {
	return c.R < 150 && c.G < 150 && c.B < 150 && c.A == 255
}
