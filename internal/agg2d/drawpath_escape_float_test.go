package agg2d

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
)

// addTriangleToRasterizer pushes a filled triangle directly into the rasterizer
// (bypassing the path/curve/stroke pipeline) so the escape-hatch render methods
// can be exercised on raw rasterizer content.
func addTriangleToRasterizer(ras interface {
	Reset()
	FillingRule(basics.FillingRule)
	AddVertex(x, y float64, cmd uint32)
}) {
	ras.Reset()
	ras.FillingRule(basics.FillNonZero)
	ras.AddVertex(2, 2, uint32(basics.PathCmdMoveTo))
	ras.AddVertex(14, 2, uint32(basics.PathCmdLineTo))
	ras.AddVertex(8, 14, uint32(basics.PathCmdLineTo))
	ras.AddVertex(0, 0, uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose))
}

// TestAgg2DFloatGetInternalRasterizer verifies the accessor returns the live
// rasterizer that the render methods consume.
func TestAgg2DFloatGetInternalRasterizer(t *testing.T) {
	a, _ := setupFloatTarget(16, 16)
	if a.GetInternalRasterizer() == nil {
		t.Fatal("GetInternalRasterizer returned nil")
	}
	if a.GetInternalRasterizer() != a.rasterizer {
		t.Fatal("GetInternalRasterizer did not return the live rasterizer")
	}
}

// TestAgg2DFloatRenderRasterizerWithColor verifies raw rasterizer content can be
// painted with a solid color without resetting the rasterizer first.
func TestAgg2DFloatRenderRasterizerWithColor(t *testing.T) {
	a, dst := setupFloatTarget(16, 16)
	addTriangleToRasterizer(a.GetInternalRasterizer())

	a.RenderRasterizerWithColor(NewColor(255, 0, 0, 255))

	if in := dst.GetPixel(8, 6); !approxF(in.R, 1.0) || in.A <= 0 {
		t.Fatalf("triangle interior pixel(8,6) = %+v, want opaque red", in)
	}
	if out := dst.GetPixel(0, 15); out.A != 0 {
		t.Fatalf("outside-triangle pixel(0,15) alpha = %v, want 0", out.A)
	}
}

// TestAgg2DFloatScanlineRender verifies the rasterizer can be rendered through a
// caller-supplied solid renderer.
func TestAgg2DFloatScanlineRender(t *testing.T) {
	a, dst := setupFloatTarget(16, 16)
	ras := a.GetInternalRasterizer()
	addTriangleToRasterizer(ras)

	renderer := a.currentRenderer()
	col := color.NewRGBA32[color.Linear](0.0, 1.0, 0.0, 1.0)
	renSolid := renscan.NewRendererScanlineAASolidWithColor(renderer, col)

	a.ScanlineRender(ras, renSolid)

	if in := dst.GetPixel(8, 6); !approxF(in.G, 1.0) || in.A <= 0 {
		t.Fatalf("ScanlineRender interior pixel(8,6) = %+v, want opaque green", in)
	}
}

// solidSpanGen is a trivial span generator emitting one constant color, used to
// exercise RenderScanlinesAAWithSpanGen.
type solidSpanGen struct {
	c color.RGBA32[color.Linear]
}

func (g *solidSpanGen) Prepare() {}

func (g *solidSpanGen) Generate(span []color.RGBA32[color.Linear], x, y, length int) {
	for i := 0; i < length && i < len(span); i++ {
		span[i] = g.c
	}
}

// TestAgg2DFloatRenderScanlinesAAWithSpanGen verifies the rasterizer can be
// rendered through a caller-supplied span generator.
func TestAgg2DFloatRenderScanlinesAAWithSpanGen(t *testing.T) {
	a, dst := setupFloatTarget(16, 16)
	ras := a.GetInternalRasterizer()
	addTriangleToRasterizer(ras)

	gen := &solidSpanGen{c: color.NewRGBA32[color.Linear](0.0, 0.0, 1.0, 1.0)}
	a.RenderScanlinesAAWithSpanGen(ras, gen)

	if in := dst.GetPixel(8, 6); !approxF(in.B, 1.0) || in.A <= 0 {
		t.Fatalf("span-gen interior pixel(8,6) = %+v, want opaque blue", in)
	}
}
