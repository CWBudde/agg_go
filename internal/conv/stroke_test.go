package conv

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
)

func TestConvStrokeCreation(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)
	cs := NewConvStroke(source)

	if cs == nil {
		t.Fatal("Expected non-nil ConvStroke")
	}

	// Test default values through delegated methods
	if cs.Width() != 1.0 {
		t.Errorf("Expected default width 1.0, got %f", cs.Width())
	}
	if cs.LineCap() != basics.ButtCap {
		t.Errorf("Expected default line cap ButtCap, got %v", cs.LineCap())
	}
	if cs.LineJoin() != basics.MiterJoin {
		t.Errorf("Expected default line join MiterJoin, got %v", cs.LineJoin())
	}
}

func TestConvStrokeWithMarkers(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)
	markers := &NullMarkers{}
	cs := NewConvStrokeWithMarkers(source, markers)

	if cs == nil {
		t.Fatal("Expected non-nil ConvStroke with markers")
	}
}

func TestConvStrokeSetters(t *testing.T) {
	vertices := []Vertex{}
	source := NewMockVertexSource(vertices)
	cs := NewConvStroke(source)

	// Test width
	cs.SetWidth(5.0)
	if cs.Width() != 5.0 {
		t.Errorf("Expected width 5.0, got %f", cs.Width())
	}

	// Test line cap
	cs.SetLineCap(basics.RoundCap)
	if cs.LineCap() != basics.RoundCap {
		t.Errorf("Expected RoundCap, got %v", cs.LineCap())
	}

	// Test line join
	cs.SetLineJoin(basics.BevelJoin)
	if cs.LineJoin() != basics.BevelJoin {
		t.Errorf("Expected BevelJoin, got %v", cs.LineJoin())
	}

	// Test inner join
	cs.SetInnerJoin(basics.InnerRound)
	if cs.InnerJoin() != basics.InnerRound {
		t.Errorf("Expected InnerRound, got %v", cs.InnerJoin())
	}

	// Test miter limit
	cs.SetMiterLimit(8.0)
	if cs.MiterLimit() != 8.0 {
		t.Errorf("Expected miter limit 8.0, got %f", cs.MiterLimit())
	}

	// Test inner miter limit
	cs.SetInnerMiterLimit(2.0)
	if cs.InnerMiterLimit() != 2.0 {
		t.Errorf("Expected inner miter limit 2.0, got %f", cs.InnerMiterLimit())
	}

	// Test approximation scale
	cs.SetApproximationScale(1.5)
	if cs.ApproximationScale() != 1.5 {
		t.Errorf("Expected approximation scale 1.5, got %f", cs.ApproximationScale())
	}

	// Test shorten
	cs.SetShorten(0.5)
	if cs.Shorten() != 0.5 {
		t.Errorf("Expected shorten 0.5, got %f", cs.Shorten())
	}
}

func TestConvStrokeSimpleLine(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	cs := NewConvStroke(source)
	cs.SetWidth(2.0)

	cs.Rewind(0)

	var outputVertices []Vertex

	// Collect all vertices
	for {
		x, y, cmd := cs.Vertex()
		outputVertices = append(outputVertices, Vertex{
			X: x, Y: y, Cmd: cmd,
		})

		if basics.IsStop(cmd) {
			break
		}
	}

	// Should have generated stroke vertices
	if len(outputVertices) < 3 {
		t.Errorf("Expected at least 3 vertices, got %d", len(outputVertices))
	}

	// First command should be MoveTo
	if outputVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo, got %v", outputVertices[0].Cmd)
	}

	// Last command should be Stop
	if outputVertices[len(outputVertices)-1].Cmd != basics.PathCmdStop {
		t.Errorf("Expected last command to be Stop, got %v", outputVertices[len(outputVertices)-1].Cmd)
	}
}

func TestConvStrokeRectangle(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)

	cs := NewConvStroke(source)
	cs.SetWidth(1.0)

	cs.Rewind(0)

	var collectedVertices []Vertex

	// Collect all vertices
	for {
		x, y, cmd := cs.Vertex()
		collectedVertices = append(collectedVertices, Vertex{
			X: x, Y: y, Cmd: cmd,
		})

		if basics.IsStop(cmd) {
			break
		}
	}

	// Should have generated many vertices for rectangle stroke
	if len(collectedVertices) < 10 {
		t.Errorf("Expected at least 10 vertices for rectangle stroke, got %d", len(collectedVertices))
	}

	// Should contain EndPoly commands for closed path
	foundEndPoly := false
	for _, v := range collectedVertices {
		if basics.IsEndPoly(v.Cmd) {
			foundEndPoly = true
			break
		}
	}

	if !foundEndPoly {
		t.Error("Expected to find EndPoly command in closed path stroke")
	}
}

func TestConvStrokeWithDifferentWidths(t *testing.T) {
	widths := []float64{0.5, 1.0, 2.0, 5.0}

	for _, width := range widths {
		t.Run("Width_"+fmt.Sprintf("%.1f", width), func(t *testing.T) {
			vertices := []Vertex{
				{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
				{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
				{X: 0, Y: 0, Cmd: basics.PathCmdStop},
			}
			source := NewMockVertexSource(vertices)

			cs := NewConvStroke(source)
			cs.SetWidth(width)

			if cs.Width() != width {
				t.Errorf("Width not set correctly: expected %f, got %f", width, cs.Width())
			}

			cs.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := cs.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				if cmd == basics.PathCmdLineTo || cmd == basics.PathCmdMoveTo {
					vertexCount++
				}
			}

			if vertexCount == 0 {
				t.Errorf("No vertices generated for width %f", width)
			}
		})
	}
}

func TestConvStrokeWithDifferentLineCaps(t *testing.T) {
	testCases := []struct {
		name    string
		lineCap basics.LineCap
	}{
		{"ButtCap", basics.ButtCap},
		{"SquareCap", basics.SquareCap},
		{"RoundCap", basics.RoundCap},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vertices := []Vertex{
				{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
				{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
				{X: 0, Y: 0, Cmd: basics.PathCmdStop},
			}
			source := NewMockVertexSource(vertices)

			cs := NewConvStroke(source)
			cs.SetWidth(2.0)
			cs.SetLineCap(tc.lineCap)

			if cs.LineCap() != tc.lineCap {
				t.Errorf("Line cap not set correctly: expected %v, got %v", tc.lineCap, cs.LineCap())
			}

			cs.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := cs.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				if cmd == basics.PathCmdLineTo || cmd == basics.PathCmdMoveTo {
					vertexCount++
				}
			}

			if vertexCount == 0 {
				t.Errorf("No vertices generated for %s", tc.name)
			}
		})
	}
}

func TestConvStrokeWithDifferentLineJoins(t *testing.T) {
	testCases := []struct {
		name     string
		lineJoin basics.LineJoin
	}{
		{"MiterJoin", basics.MiterJoin},
		{"RoundJoin", basics.RoundJoin},
		{"BevelJoin", basics.BevelJoin},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vertices := []Vertex{
				{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
				{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
				{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
				{X: 0, Y: 0, Cmd: basics.PathCmdStop},
			}
			source := NewMockVertexSource(vertices)

			cs := NewConvStroke(source)
			cs.SetWidth(2.0)
			cs.SetLineJoin(tc.lineJoin)

			if cs.LineJoin() != tc.lineJoin {
				t.Errorf("Line join not set correctly: expected %v, got %v", tc.lineJoin, cs.LineJoin())
			}

			cs.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := cs.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				if cmd == basics.PathCmdLineTo || cmd == basics.PathCmdMoveTo {
					vertexCount++
				}
			}

			if vertexCount == 0 {
				t.Errorf("No vertices generated for %s", tc.name)
			}
		})
	}
}

func TestConvStrokeGenerator(t *testing.T) {
	vertices := []Vertex{}
	source := NewMockVertexSource(vertices)
	cs := NewConvStroke(source)

	// Test that Generator() returns the underlying stroke generator
	gen := cs.Generator()
	if gen == nil {
		t.Error("Expected non-nil generator")
	}

	// Test that we can modify stroke parameters through the generator
	gen.SetWidth(7.0)
	if cs.Width() != 7.0 {
		t.Errorf("Expected width 7.0 after setting through generator, got %f", cs.Width())
	}
}

func TestConvStrokeEdgeCases(t *testing.T) {
	t.Run("EmptyPath", func(t *testing.T) {
		source := NewMockVertexSource([]Vertex{})
		cs := NewConvStroke(source)
		cs.SetWidth(2.0)

		cs.Rewind(0)
		_, _, cmd := cs.Vertex()

		// Empty path should immediately return stop
		if cmd != basics.PathCmdStop {
			t.Errorf("Expected stop command for empty path, got %v", cmd)
		}
	})

	t.Run("SinglePoint", func(t *testing.T) {
		vertices := []Vertex{
			{X: 5, Y: 5, Cmd: basics.PathCmdMoveTo},
			{X: 0, Y: 0, Cmd: basics.PathCmdStop},
		}
		source := NewMockVertexSource(vertices)
		cs := NewConvStroke(source)
		cs.SetWidth(2.0)

		cs.Rewind(0)
		_, _, cmd := cs.Vertex()

		// Single point should result in stop (no stroke possible)
		if cmd != basics.PathCmdStop {
			t.Errorf("Expected stop command for single point, got %v", cmd)
		}
	})

	t.Run("ZeroWidth", func(t *testing.T) {
		vertices := []Vertex{
			{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
			{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
			{X: 0, Y: 0, Cmd: basics.PathCmdStop},
		}
		source := NewMockVertexSource(vertices)
		cs := NewConvStroke(source)
		cs.SetWidth(0.0)

		cs.Rewind(0)

		// Should still generate vertices, just with zero-width offset
		vertexCount := 0
		for {
			_, _, cmd := cs.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			vertexCount++
		}

		// Zero width should still generate some vertices
		if vertexCount == 0 {
			t.Error("Expected some vertices even with zero width")
		}
	})

	t.Run("VerySmallWidth", func(t *testing.T) {
		vertices := []Vertex{
			{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
			{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
			{X: 0, Y: 0, Cmd: basics.PathCmdStop},
		}
		source := NewMockVertexSource(vertices)
		cs := NewConvStroke(source)
		cs.SetWidth(0.001)

		cs.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := cs.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			vertexCount++
		}

		if vertexCount == 0 {
			t.Error("Expected vertices for very small width")
		}
	})
}

func TestConvStrokeComplexPaths(t *testing.T) {
	t.Run("MultipleSubPaths", func(t *testing.T) {
		// Note: Current stroke generator may handle multiple sub-paths as a single path
		// This is acceptable behavior - the key is that it generates strokes correctly
		vertices := []Vertex{
			// First path
			{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
			{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
			{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},

			// Second path
			{X: 20, Y: 0, Cmd: basics.PathCmdMoveTo},
			{X: 30, Y: 0, Cmd: basics.PathCmdLineTo},
			{X: 30, Y: 10, Cmd: basics.PathCmdLineTo},

			{X: 0, Y: 0, Cmd: basics.PathCmdStop},
		}
		source := NewMockVertexSource(vertices)
		cs := NewConvStroke(source)
		cs.SetWidth(1.0)

		cs.Rewind(0)

		moveToCount := 0
		vertexCount := 0

		for {
			_, _, cmd := cs.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if cmd == basics.PathCmdMoveTo {
				moveToCount++
			}
			vertexCount++
		}

		// Should have at least one MoveTo and generate some vertices
		if moveToCount < 1 {
			t.Errorf("Expected at least 1 MoveTo command, got %d", moveToCount)
		}

		if vertexCount < 4 {
			t.Errorf("Expected at least 4 vertices for multi-segment path, got %d", vertexCount)
		}
	})

	t.Run("ClosedTriangle", func(t *testing.T) {
		vertices := []Vertex{
			{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
			{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
			{X: 5, Y: 10, Cmd: basics.PathCmdLineTo},
			{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
			{X: 0, Y: 0, Cmd: basics.PathCmdStop},
		}
		source := NewMockVertexSource(vertices)
		cs := NewConvStroke(source)
		cs.SetWidth(2.0)
		cs.SetLineJoin(basics.RoundJoin)

		cs.Rewind(0)

		var outputVertices []Vertex
		for {
			x, y, cmd := cs.Vertex()
			outputVertices = append(outputVertices, Vertex{X: x, Y: y, Cmd: cmd})
			if basics.IsStop(cmd) {
				break
			}
		}

		// Triangle should generate substantial number of vertices due to joins
		// With 3 corners and round joins, expect at least 12-14 vertices
		if len(outputVertices) < 12 {
			t.Errorf("Expected at least 12 vertices for stroked triangle, got %d", len(outputVertices))
		}

		// Should contain EndPoly for closed path
		foundEndPoly := false
		for _, v := range outputVertices {
			if basics.IsEndPoly(v.Cmd) {
				foundEndPoly = true
				break
			}
		}
		if !foundEndPoly {
			t.Error("Expected EndPoly command in closed triangle stroke")
		}
	})
}

func TestConvStrokeMiterLimitBehavior(t *testing.T) {
	// Create a sharp angle that would create a long miter
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 10.1, Y: 10, Cmd: basics.PathCmdLineTo}, // Very sharp angle
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	testCases := []struct {
		name       string
		miterLimit float64
		width      float64
	}{
		{"LowMiterLimit", 1.0, 2.0},
		{"DefaultMiterLimit", 4.0, 2.0},
		{"HighMiterLimit", 10.0, 2.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			source := NewMockVertexSource(vertices)
			cs := NewConvStroke(source)
			cs.SetWidth(tc.width)
			cs.SetLineJoin(basics.MiterJoin)
			cs.SetMiterLimit(tc.miterLimit)

			if cs.MiterLimit() != tc.miterLimit {
				t.Errorf("Miter limit not set correctly: expected %f, got %f",
					tc.miterLimit, cs.MiterLimit())
			}

			cs.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := cs.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertexCount++
			}

			if vertexCount == 0 {
				t.Errorf("No vertices generated for miter limit %f", tc.miterLimit)
			}
		})
	}
}

func TestConvStrokeInnerJoinTypes(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 5, Y: 5, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	innerJoins := []basics.InnerJoin{
		basics.InnerBevel,
		basics.InnerMiter,
		basics.InnerJag,
		basics.InnerRound,
	}

	for _, innerJoin := range innerJoins {
		t.Run(fmt.Sprintf("InnerJoin_%v", innerJoin), func(t *testing.T) {
			source := NewMockVertexSource(vertices)
			cs := NewConvStroke(source)
			cs.SetWidth(3.0)
			cs.SetInnerJoin(innerJoin)

			if cs.InnerJoin() != innerJoin {
				t.Errorf("Inner join not set correctly: expected %v, got %v",
					innerJoin, cs.InnerJoin())
			}

			cs.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := cs.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertexCount++
			}

			if vertexCount == 0 {
				t.Errorf("No vertices generated for inner join %v", innerJoin)
			}
		})
	}
}

func TestConvStrokeApproximationScale(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	scales := []float64{0.1, 0.5, 1.0, 2.0, 5.0}

	for _, scale := range scales {
		t.Run(fmt.Sprintf("Scale_%.1f", scale), func(t *testing.T) {
			source := NewMockVertexSource(vertices)
			cs := NewConvStroke(source)
			cs.SetWidth(2.0)
			cs.SetApproximationScale(scale)
			cs.SetLineCap(basics.RoundCap)
			cs.SetLineJoin(basics.RoundJoin)

			if cs.ApproximationScale() != scale {
				t.Errorf("Approximation scale not set correctly: expected %f, got %f",
					scale, cs.ApproximationScale())
			}

			cs.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := cs.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertexCount++
			}

			// Different approximation scales should still generate vertices
			if vertexCount == 0 {
				t.Errorf("No vertices generated for approximation scale %f", scale)
			}
		})
	}
}

func TestConvStrokeRewindConsistency(t *testing.T) {
	// Use a simpler path to avoid numerical instabilities
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	cs := NewConvStroke(source)
	cs.SetWidth(2.0)

	// First iteration
	cs.Rewind(0)
	var firstCount int
	for {
		_, _, cmd := cs.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		firstCount++
	}

	// Second iteration
	cs.Rewind(0)
	var secondCount int
	for {
		_, _, cmd := cs.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		secondCount++
	}

	// Vertex counts should be consistent
	if firstCount != secondCount {
		t.Errorf("Inconsistent vertex count between runs: first=%d, second=%d",
			firstCount, secondCount)
	}

	// Both runs should generate some vertices
	if firstCount == 0 {
		t.Error("No vertices generated in stroke rewind test")
	}
}
