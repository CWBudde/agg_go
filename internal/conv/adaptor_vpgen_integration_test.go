package conv

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/vpgen"
)

func TestConvAdaptorVPGen_ClipPolygon(t *testing.T) {
	// Create a vertex source with a polygon that crosses clip bounds
	vertices := []CurveVertex{
		{X: -10, Y: -10, Cmd: basics.PathCmdMoveTo}, // Outside
		{X: 50, Y: -10, Cmd: basics.PathCmdLineTo},  // Cross
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},   // Inside
		{X: -10, Y: 50, Cmd: basics.PathCmdLineTo},  // Cross
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
	}

	source := NewCurveVertexSource(vertices)
	clipPolygon := vpgen.NewVPGenClipPolygon()
	clipPolygon.SetClipBox(0, 0, 100, 100)

	adaptor := NewConvAdaptorVPGen(source, clipPolygon)
	adaptor.Rewind(0)

	// Collect clipped vertices
	var resultVertices []CurveVertex
	for {
		x, y, cmd := adaptor.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have clipped vertices
	if len(resultVertices) == 0 {
		t.Error("Expected clipped vertices, got none")
	}

	// All vertices should be within clip bounds or on boundary
	for i, v := range resultVertices {
		if v.X < 0 || v.X > 100 || v.Y < 0 || v.Y > 100 {
			t.Errorf("Vertex %d at (%.1f, %.1f) is outside clip bounds [0,0,100,100]", i, v.X, v.Y)
		}
	}

	// First vertex should be MoveTo
	if len(resultVertices) > 0 && resultVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", resultVertices[0].Cmd)
	}
}

func TestConvAdaptorVPGen_ClipPolyline(t *testing.T) {
	// Create a polyline that crosses clip bounds
	vertices := []CurveVertex{
		{X: -20, Y: 50, Cmd: basics.PathCmdMoveTo}, // Outside
		{X: 120, Y: 50, Cmd: basics.PathCmdLineTo}, // Cross entire clip box
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	clipPolyline := vpgen.NewVPGenClipPolyline()
	clipPolyline.SetClipBox(0, 0, 100, 100)

	adaptor := NewConvAdaptorVPGen(source, clipPolyline)
	adaptor.Rewind(0)

	// Collect clipped vertices
	var resultVertices []CurveVertex
	for {
		x, y, cmd := adaptor.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have exactly 2 vertices (entry and exit points)
	if len(resultVertices) != 2 {
		t.Errorf("Expected 2 clipped vertices, got %d", len(resultVertices))
	}

	if len(resultVertices) >= 2 {
		// Should be clipped to box boundaries
		if resultVertices[0].X != 0 || resultVertices[0].Y != 50 {
			t.Errorf("First vertex should be (0, 50), got (%.1f, %.1f)", resultVertices[0].X, resultVertices[0].Y)
		}
		if resultVertices[1].X != 100 || resultVertices[1].Y != 50 {
			t.Errorf("Second vertex should be (100, 50), got (%.1f, %.1f)", resultVertices[1].X, resultVertices[1].Y)
		}
	}
}

func TestConvAdaptorVPGen_Segmentator(t *testing.T) {
	// Create a long line that should be segmented
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo}, // 10 units long
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	segmentator := vpgen.NewVPGenSegmentator()
	segmentator.SetApproximationScale(1.0) // 1 unit per segment

	adaptor := NewConvAdaptorVPGen(source, segmentator)
	adaptor.Rewind(0)

	// Collect segmented vertices
	var resultVertices []CurveVertex
	for {
		x, y, cmd := adaptor.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have multiple segments
	if len(resultVertices) < 3 {
		t.Errorf("Expected multiple segments (3+ vertices), got %d", len(resultVertices))
	}

	// First vertex should be MoveTo at start
	if len(resultVertices) > 0 && resultVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", resultVertices[0].Cmd)
	}

	// Last vertex should end at the line endpoint
	if len(resultVertices) > 0 {
		lastVertex := resultVertices[len(resultVertices)-1]
		if lastVertex.X != 10 || lastVertex.Y != 0 {
			t.Errorf("Last vertex should be at (10, 0), got (%.1f, %.1f)", lastVertex.X, lastVertex.Y)
		}
	}
}

func TestConvAdaptorVPGen_InterfaceCompliance(t *testing.T) {
	// Test that all vpgen components implement the VPGen interface
	var _ VPGen = (*vpgen.VPGenClipPolygon)(nil)
	var _ VPGen = (*vpgen.VPGenClipPolyline)(nil)
	var _ VPGen = (*vpgen.VPGenSegmentator)(nil)

	// Test that they can be used with ConvAdaptorVPGen
	vertices := []CurveVertex{{X: 0, Y: 0, Cmd: basics.PathCmdStop}}
	source := NewCurveVertexSource(vertices)

	// Should compile without issues
	_ = NewConvAdaptorVPGen(source, vpgen.NewVPGenClipPolygon())
	_ = NewConvAdaptorVPGen(source, vpgen.NewVPGenClipPolyline())
	_ = NewConvAdaptorVPGen(source, vpgen.NewVPGenSegmentator())

	// Test successful interface compliance
	t.Log("All vpgen components successfully implement VPGen interface")
}
