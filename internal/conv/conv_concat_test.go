package conv

import (
	"testing"

	"agg_go/internal/basics"
)

func TestConvConcat_NewConvConcat(t *testing.T) {
	source1 := NewMockVertexSource([]Vertex{})
	source2 := NewMockVertexSource([]Vertex{})
	concat := NewConvConcat(source1, source2)

	if concat == nil {
		t.Error("NewConvConcat should return non-nil converter")
	}
}

func TestConvConcat_BasicConcatenation(t *testing.T) {
	// First path: simple line
	vertices1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
	}

	// Second path: another line
	vertices2 := []Vertex{
		{20, 20, basics.PathCmdMoveTo},
		{30, 20, basics.PathCmdLineTo},
		{30, 30, basics.PathCmdLineTo},
	}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	concat := NewConvConcat(source1, source2)

	concat.Rewind(0)

	// Expected output: all vertices from source1, then all from source2
	expected := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{20, 20, basics.PathCmdMoveTo},
		{30, 20, basics.PathCmdLineTo},
		{30, 30, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := concat.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	if len(output) != len(expected) {
		t.Errorf("Expected %d vertices, got %d", len(expected), len(output))
	}

	for i, v := range expected {
		if i >= len(output) {
			t.Errorf("Missing vertex at index %d", i)
			continue
		}
		if output[i].X != v.X || output[i].Y != v.Y || output[i].Cmd != v.Cmd {
			t.Errorf("Vertex %d: expected (%g, %g, %d), got (%g, %g, %d)",
				i, v.X, v.Y, v.Cmd, output[i].X, output[i].Y, output[i].Cmd)
		}
	}
}

func TestConvConcat_EmptyFirstSource(t *testing.T) {
	// First path: empty
	vertices1 := []Vertex{}

	// Second path: simple line
	vertices2 := []Vertex{
		{20, 20, basics.PathCmdMoveTo},
		{30, 20, basics.PathCmdLineTo},
	}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	concat := NewConvConcat(source1, source2)

	concat.Rewind(0)

	// Should get only vertices from source2
	expected := []Vertex{
		{20, 20, basics.PathCmdMoveTo},
		{30, 20, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	var output []Vertex
	for {
		x, y, cmd := concat.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	if len(output) != len(expected) {
		t.Errorf("Expected %d vertices, got %d", len(expected), len(output))
	}

	for i, v := range expected {
		if i >= len(output) {
			t.Errorf("Missing vertex at index %d", i)
			continue
		}
		if output[i].X != v.X || output[i].Y != v.Y || output[i].Cmd != v.Cmd {
			t.Errorf("Vertex %d: expected (%g, %g, %d), got (%g, %g, %d)",
				i, v.X, v.Y, v.Cmd, output[i].X, output[i].Y, output[i].Cmd)
		}
	}
}

func TestConvConcat_EmptySecondSource(t *testing.T) {
	// First path: simple line
	vertices1 := []Vertex{
		{10, 10, basics.PathCmdMoveTo},
		{20, 20, basics.PathCmdLineTo},
	}

	// Second path: empty
	vertices2 := []Vertex{}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	concat := NewConvConcat(source1, source2)

	concat.Rewind(0)

	// Should get only vertices from source1
	expected := []Vertex{
		{10, 10, basics.PathCmdMoveTo},
		{20, 20, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	var output []Vertex
	for {
		x, y, cmd := concat.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	if len(output) != len(expected) {
		t.Errorf("Expected %d vertices, got %d", len(expected), len(output))
	}

	for i, v := range expected {
		if i >= len(output) {
			t.Errorf("Missing vertex at index %d", i)
			continue
		}
		if output[i].X != v.X || output[i].Y != v.Y || output[i].Cmd != v.Cmd {
			t.Errorf("Vertex %d: expected (%g, %g, %d), got (%g, %g, %d)",
				i, v.X, v.Y, v.Cmd, output[i].X, output[i].Y, output[i].Cmd)
		}
	}
}

func TestConvConcat_BothSourcesEmpty(t *testing.T) {
	source1 := NewMockVertexSource([]Vertex{})
	source2 := NewMockVertexSource([]Vertex{})
	concat := NewConvConcat(source1, source2)

	concat.Rewind(0)

	// Should immediately return Stop
	x, y, cmd := concat.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop command, got %d", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("Expected (0, 0) coordinates for stop, got (%g, %g)", x, y)
	}
}

func TestConvConcat_Attach1(t *testing.T) {
	// Initial sources
	initial1 := NewMockVertexSource([]Vertex{{0, 0, basics.PathCmdMoveTo}})
	initial2 := NewMockVertexSource([]Vertex{{10, 10, basics.PathCmdMoveTo}})
	concat := NewConvConcat(initial1, initial2)

	// New source1
	new1 := NewMockVertexSource([]Vertex{{100, 100, basics.PathCmdMoveTo}})
	concat.Attach1(new1)

	concat.Rewind(0)

	// First vertex should come from new source1
	x, y, cmd := concat.Vertex()
	if x != 100 || y != 100 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected (100, 100, MoveTo) from new source1, got (%g, %g, %d)", x, y, cmd)
	}
}

func TestConvConcat_Attach2(t *testing.T) {
	// Initial sources
	initial1 := NewMockVertexSource([]Vertex{})
	initial2 := NewMockVertexSource([]Vertex{{10, 10, basics.PathCmdMoveTo}})
	concat := NewConvConcat(initial1, initial2)

	// New source2
	new2 := NewMockVertexSource([]Vertex{{200, 200, basics.PathCmdMoveTo}})
	concat.Attach2(new2)

	concat.Rewind(0)

	// Should get vertex from new source2 (since source1 is empty)
	x, y, cmd := concat.Vertex()
	if x != 200 || y != 200 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected (200, 200, MoveTo) from new source2, got (%g, %g, %d)", x, y, cmd)
	}
}

func TestConvConcat_RewindFunctionality(t *testing.T) {
	vertices1 := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 2, basics.PathCmdLineTo},
	}
	vertices2 := []Vertex{
		{3, 3, basics.PathCmdMoveTo},
		{4, 4, basics.PathCmdLineTo},
	}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	concat := NewConvConcat(source1, source2)

	// Read all vertices once
	concat.Rewind(0)
	var firstRun []Vertex
	for {
		x, y, cmd := concat.Vertex()
		firstRun = append(firstRun, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Rewind and read again
	concat.Rewind(0)
	var secondRun []Vertex
	for {
		x, y, cmd := concat.Vertex()
		secondRun = append(secondRun, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Both runs should be identical
	if len(firstRun) != len(secondRun) {
		t.Errorf("Rewind failed: first run had %d vertices, second run had %d",
			len(firstRun), len(secondRun))
	}

	for i := range firstRun {
		if i >= len(secondRun) {
			t.Errorf("Second run shorter than first at index %d", i)
			continue
		}
		if firstRun[i].X != secondRun[i].X ||
			firstRun[i].Y != secondRun[i].Y ||
			firstRun[i].Cmd != secondRun[i].Cmd {
			t.Errorf("Rewind failed at vertex %d: first (%g, %g, %d), second (%g, %g, %d)",
				i, firstRun[i].X, firstRun[i].Y, firstRun[i].Cmd,
				secondRun[i].X, secondRun[i].Y, secondRun[i].Cmd)
		}
	}
}

func TestConvConcat_WithCurveCommands(t *testing.T) {
	// First path with curve
	vertices1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdCurve3},
		{20, 10, basics.PathCmdCurve3},
	}

	// Second path with different curve
	vertices2 := []Vertex{
		{30, 30, basics.PathCmdMoveTo},
		{40, 30, basics.PathCmdCurve4},
		{50, 40, basics.PathCmdCurve4},
		{60, 50, basics.PathCmdCurve4},
	}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	concat := NewConvConcat(source1, source2)

	concat.Rewind(0)

	// Collect all output
	var output []Vertex
	for {
		x, y, cmd := concat.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Should have all vertices from both sources plus stop
	expectedLen := len(vertices1) + len(vertices2) + 1
	if len(output) != expectedLen {
		t.Errorf("Expected %d vertices including Stop, got %d", expectedLen, len(output))
	}

	// Verify curve commands are preserved
	curveCount := 0
	for _, v := range output {
		if basics.IsCurve(v.Cmd) {
			curveCount++
		}
	}

	expectedCurveCount := 5 // 2 from vertices1, 3 from vertices2
	if curveCount != expectedCurveCount {
		t.Errorf("Expected %d curve commands, got %d", expectedCurveCount, curveCount)
	}
}
