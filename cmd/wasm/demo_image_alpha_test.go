package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/transform"
)

func TestImageAlphaWASMClipSourceMatchesCXXEllipse(t *testing.T) {
	src := newImageAlphaClipSource(320, 300, transform.NewTransAffine())
	src.Rewind(0)

	var x, y float64
	cmd := basics.PathCommand(src.Vertex(&x, &y))
	if !basics.IsMoveTo(cmd) {
		t.Fatalf("first clip command = %v, want move_to", cmd)
	}
	if x != 320.0/2.0+320.0/1.9 || y != 300.0/2.0 {
		t.Fatalf("first clip vertex = (%g,%g), want C++ ellipse start (%g,%g)", x, y, 320.0/2.0+320.0/1.9, 300.0/2.0)
	}

	for i := 1; i < imageAlphaClipEllipseSteps; i++ {
		cmd = basics.PathCommand(src.Vertex(&x, &y))
		if !basics.IsVertex(cmd) {
			t.Fatalf("clip source command %d = %v, want vertex", i, cmd)
		}
	}

	cmd = basics.PathCommand(src.Vertex(&x, &y))
	if !basics.IsEndPoly(cmd) || !basics.IsClosed(uint32(cmd)) || !basics.IsCCW(uint32(cmd)) {
		t.Fatalf("clip source end command = %v, want C++ ellipse end_poly|close|ccw", cmd)
	}
}

func TestImageAlphaWASMRandomEllipseUsesGCCCXXRGBAOrder(t *testing.T) {
	initImgAlphaDemo()

	want := imageAlphaSRGBA8(236, 74, 255, 81)
	if got := imgAlphaEllipses[0].color; got != want {
		t.Fatalf("first ellipse color = %+v, want GCC C++ rand argument order converted from srgba8 %+v", got, want)
	}
}
