package rbox

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewDefaultRboxCtrl(t *testing.T) {
	// Test basic construction
	ctrl := NewDefaultRboxCtrl(10, 20, 200, 150, false)

	// Verify bounds
	if ctrl.X1() != 10 || ctrl.Y1() != 20 || ctrl.X2() != 200 || ctrl.Y2() != 150 {
		t.Errorf("Expected bounds (10,20,200,150), got (%.1f,%.1f,%.1f,%.1f)",
			ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())
	}

	// Verify initial state
	if ctrl.NumItems() != 0 {
		t.Errorf("Expected 0 items initially, got %d", ctrl.NumItems())
	}

	if ctrl.CurItem() != -1 {
		t.Errorf("Expected no item selected initially (-1), got %d", ctrl.CurItem())
	}

	// Verify FlipY setting
	if ctrl.FlipY() != false {
		t.Errorf("Expected FlipY to be false, got %t", ctrl.FlipY())
	}

	// Test with FlipY enabled
	ctrlFlipped := NewDefaultRboxCtrl(0, 0, 100, 100, true)
	if ctrlFlipped.FlipY() != true {
		t.Errorf("Expected FlipY to be true, got %t", ctrlFlipped.FlipY())
	}
}

func TestAddItem(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Test adding items
	testItems := []string{
		"Option 1",
		"Option 2",
		"Very Long Option Name That Should Still Work",
		"简体中文",    // Test Unicode
		"🎵 Music", // Test emoji
	}

	for i, item := range testItems {
		success := ctrl.AddItem(item)
		if !success {
			t.Errorf("Failed to add item %d: %s", i, item)
		}

		if ctrl.NumItems() != uint32(i+1) {
			t.Errorf("Expected %d items after adding item %d, got %d", i+1, i, ctrl.NumItems())
		}

		if ctrl.ItemText(i) != item {
			t.Errorf("Expected item %d text to be '%s', got '%s'", i, item, ctrl.ItemText(i))
		}
	}

	// Test maximum items (32)
	ctrl = NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Add exactly 32 items
	for i := 0; i < 32; i++ {
		success := ctrl.AddItem("Item " + string(rune('A'+i)))
		if !success {
			t.Errorf("Failed to add item %d (should be within limit)", i)
		}
	}

	if ctrl.NumItems() != 32 {
		t.Errorf("Expected 32 items, got %d", ctrl.NumItems())
	}

	// Try to add 33rd item (should fail)
	success := ctrl.AddItem("Item 33")
	if success {
		t.Errorf("Expected failure when adding 33rd item, but it succeeded")
	}

	if ctrl.NumItems() != 32 {
		t.Errorf("Expected 32 items after failed add, got %d", ctrl.NumItems())
	}
}

func TestItemSelection(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Add some test items
	items := []string{"Red", "Green", "Blue", "Yellow"}
	for _, item := range items {
		ctrl.AddItem(item)
	}

	// Test initial state (no selection)
	if ctrl.CurItem() != -1 {
		t.Errorf("Expected no selection initially (-1), got %d", ctrl.CurItem())
	}

	// Test selecting each item
	for i := 0; i < len(items); i++ {
		ctrl.SetCurItem(i)
		if ctrl.CurItem() != i {
			t.Errorf("Expected selected item %d, got %d", i, ctrl.CurItem())
		}
	}

	// Test deselecting all
	ctrl.SetCurItem(-1)
	if ctrl.CurItem() != -1 {
		t.Errorf("Expected no selection (-1), got %d", ctrl.CurItem())
	}

	// Test invalid selections
	ctrl.SetCurItem(len(items)) // Out of range (high)
	if ctrl.CurItem() != -1 {
		t.Errorf("Expected selection to remain -1 after invalid high index, got %d", ctrl.CurItem())
	}

	ctrl.SetCurItem(-2) // Out of range (low)
	if ctrl.CurItem() != -1 {
		t.Errorf("Expected selection to remain -1 after invalid low index, got %d", ctrl.CurItem())
	}

	// Test valid selection after invalid attempts
	ctrl.SetCurItem(1)
	if ctrl.CurItem() != 1 {
		t.Errorf("Expected valid selection to work after invalid attempts, got %d", ctrl.CurItem())
	}
}

func TestBorderStyling(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Test initial border width
	if ctrl.BorderWidth() != 1.0 {
		t.Errorf("Expected initial border width 1.0, got %.2f", ctrl.BorderWidth())
	}

	// Test setting border width
	ctrl.SetBorderWidth(2.5, 1.0)
	if ctrl.BorderWidth() != 2.5 {
		t.Errorf("Expected border width 2.5, got %.2f", ctrl.BorderWidth())
	}

	// Verify that inner bounds are recalculated (xs1 should be x1 + borderWidth)
	expectedXs1 := ctrl.X1() + 2.5
	if ctrl.xs1 != expectedXs1 {
		t.Errorf("Expected xs1 to be %.2f after border change, got %.2f", expectedXs1, ctrl.xs1)
	}
}

func TestTextStyling(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Test initial text settings
	if ctrl.TextThickness() != 1.5 {
		t.Errorf("Expected initial text thickness 1.5, got %.2f", ctrl.TextThickness())
	}

	if ctrl.TextHeight() != 9.0 {
		t.Errorf("Expected initial text height 9.0, got %.2f", ctrl.TextHeight())
	}

	if ctrl.TextWidth() != 0.0 {
		t.Errorf("Expected initial text width 0.0, got %.2f", ctrl.TextWidth())
	}

	// Test setting text thickness
	ctrl.SetTextThickness(2.0)
	if ctrl.TextThickness() != 2.0 {
		t.Errorf("Expected text thickness 2.0, got %.2f", ctrl.TextThickness())
	}

	// Test setting text size
	ctrl.SetTextSize(12.0, 8.0)
	if ctrl.TextHeight() != 12.0 {
		t.Errorf("Expected text height 12.0, got %.2f", ctrl.TextHeight())
	}

	if ctrl.TextWidth() != 8.0 {
		t.Errorf("Expected text width 8.0, got %.2f", ctrl.TextWidth())
	}

	// Verify that dy is updated (should be height * 2.0)
	expectedDy := 12.0 * 2.0
	if ctrl.dy != expectedDy {
		t.Errorf("Expected dy to be %.2f after text size change, got %.2f", expectedDy, ctrl.dy)
	}
}

func TestColorManagement(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Test setting and getting colors
	testColors := []color.RGBA{
		color.NewRGBA(1.0, 0.0, 0.0, 1.0), // Red
		color.NewRGBA(0.0, 1.0, 0.0, 1.0), // Green
		color.NewRGBA(0.0, 0.0, 1.0, 1.0), // Blue
		color.NewRGBA(1.0, 1.0, 0.0, 1.0), // Yellow
		color.NewRGBA(1.0, 0.0, 1.0, 1.0), // Magenta
	}

	// Set colors
	ctrl.SetBackgroundColor(testColors[0])
	ctrl.SetBorderColor(testColors[1])
	ctrl.SetTextColor(testColors[2])
	ctrl.SetInactiveColor(testColors[3])
	ctrl.SetActiveColor(testColors[4])

	// Verify colors through Color interface
	for i, expectedColor := range testColors {
		actualColor := ctrl.Color(uint(i))
		if actualColor != expectedColor {
			t.Errorf("Expected color %d to be %v, got %v", i, expectedColor, actualColor)
		}
	}

	// Test invalid path ID (should return background color)
	invalidColorVal := ctrl.Color(10)
	if invalidColorVal != testColors[0] {
		t.Errorf("Expected invalid path ID to return background color %v, got %v", testColors[0], invalidColorVal)
	}
}

func TestMouseInteraction(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(10, 10, 200, 100, false)

	// Add some items
	ctrl.AddItem("Option 1")
	ctrl.AddItem("Option 2")
	ctrl.AddItem("Option 3")

	// Test clicking on first radio button
	// Radio buttons are positioned at xs1 + dy/1.3, ys1 + dy*item + dy/1.3
	x1 := ctrl.xs1 + ctrl.dy/1.3
	y1 := ctrl.ys1 + ctrl.dy/1.3

	handled := ctrl.OnMouseButtonDown(x1, y1)
	if !handled {
		t.Errorf("Expected mouse click on first radio button to be handled")
	}

	if ctrl.CurItem() != 0 {
		t.Errorf("Expected first item to be selected after click, got %d", ctrl.CurItem())
	}

	// Test clicking on second radio button
	y2 := ctrl.ys1 + ctrl.dy + ctrl.dy/1.3
	handled = ctrl.OnMouseButtonDown(x1, y2)
	if !handled {
		t.Errorf("Expected mouse click on second radio button to be handled")
	}

	if ctrl.CurItem() != 1 {
		t.Errorf("Expected second item to be selected after click, got %d", ctrl.CurItem())
	}

	// Test clicking outside radio buttons
	handled = ctrl.OnMouseButtonDown(x1+50, y1) // Far to the right
	if handled {
		t.Errorf("Expected mouse click outside radio buttons to not be handled")
	}

	if ctrl.CurItem() != 1 { // Should remain unchanged
		t.Errorf("Expected selection to remain unchanged after clicking outside, got %d", ctrl.CurItem())
	}

	// Test OnMouseButtonUp and OnMouseMove (should return false)
	if ctrl.OnMouseButtonUp(x1, y1) {
		t.Errorf("Expected OnMouseButtonUp to return false")
	}

	if ctrl.OnMouseMove(x1, y1, false) {
		t.Errorf("Expected OnMouseMove to return false")
	}
}

func TestArrowKeyNavigation(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Add some items
	ctrl.AddItem("First")
	ctrl.AddItem("Second")
	ctrl.AddItem("Third")

	// Initially no item selected - arrow keys should not work
	if ctrl.OnArrowKeys(false, false, false, true) { // Up
		t.Errorf("Expected arrow keys to not work when no item selected")
	}

	// Select first item
	ctrl.SetCurItem(0)

	// Test right arrow (should go to next item)
	if !ctrl.OnArrowKeys(false, true, false, false) {
		t.Errorf("Expected right arrow key to be handled")
	}

	if ctrl.CurItem() != 1 {
		t.Errorf("Expected second item to be selected after right arrow, got %d", ctrl.CurItem())
	}

	// Test up arrow (should also go to next item)
	if !ctrl.OnArrowKeys(false, false, false, true) {
		t.Errorf("Expected up arrow key to be handled")
	}

	if ctrl.CurItem() != 2 {
		t.Errorf("Expected third item to be selected after up arrow, got %d", ctrl.CurItem())
	}

	// Test right arrow at end (should wrap to beginning)
	if !ctrl.OnArrowKeys(false, true, false, false) {
		t.Errorf("Expected right arrow key to be handled at end")
	}

	if ctrl.CurItem() != 0 {
		t.Errorf("Expected first item to be selected after wrap-around, got %d", ctrl.CurItem())
	}

	// Test left arrow (should go to previous item, wrapping to end)
	if !ctrl.OnArrowKeys(true, false, false, false) {
		t.Errorf("Expected left arrow key to be handled")
	}

	if ctrl.CurItem() != 2 {
		t.Errorf("Expected last item to be selected after left arrow wrap, got %d", ctrl.CurItem())
	}

	// Test down arrow (should also go to previous item)
	if !ctrl.OnArrowKeys(false, false, true, false) {
		t.Errorf("Expected down arrow key to be handled")
	}

	if ctrl.CurItem() != 1 {
		t.Errorf("Expected second item to be selected after down arrow, got %d", ctrl.CurItem())
	}

	// Test no keys pressed
	if ctrl.OnArrowKeys(false, false, false, false) {
		t.Errorf("Expected no action when no arrow keys are pressed")
	}
}

func TestVertexGeneration(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Add some test items
	ctrl.AddItem("Item 1")
	ctrl.AddItem("Item 2")
	ctrl.SetCurItem(0) // Select first item

	// Test NumPaths
	if ctrl.NumPaths() != 5 {
		t.Errorf("Expected 5 paths, got %d", ctrl.NumPaths())
	}

	// Test each path generates vertices without panicking
	for pathID := uint(0); pathID < 5; pathID++ {
		ctrl.Rewind(pathID)

		vertexCount := 0
		maxVertices := 1000 // Safety limit to prevent infinite loops

		for vertexCount < maxVertices {
			x, y, cmd := ctrl.Vertex()

			if cmd == basics.PathCmdStop {
				break
			}

			// Verify that coordinates are reasonable (not NaN, not infinite)
			if x != x || y != y { // NaN check
				t.Errorf("Path %d produced NaN coordinates at vertex %d: (%.2f, %.2f)", pathID, vertexCount, x, y)
			}

			// Check for reasonable bounds (within some reasonable range)
			if x < -1000 || x > 1000 || y < -1000 || y > 1000 {
				t.Errorf("Path %d produced extreme coordinates at vertex %d: (%.2f, %.2f)", pathID, vertexCount, x, y)
			}

			vertexCount++
		}

		if vertexCount >= maxVertices {
			t.Errorf("Path %d generated too many vertices (possible infinite loop)", pathID)
		}

		if vertexCount == 0 {
			// Only path 4 (active circle) can legitimately have 0 vertices when no item is selected
			if pathID != 4 || ctrl.CurItem() >= 0 {
				t.Errorf("Path %d generated no vertices", pathID)
			}
		}
	}
}

func TestEmptyRboxVertexGeneration(t *testing.T) {
	// Test vertex generation with no items
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Background and border should still generate vertices
	for _, pathID := range []uint{0, 1} {
		ctrl.Rewind(pathID)
		_, _, cmd := ctrl.Vertex()
		if cmd == basics.PathCmdStop {
			t.Errorf("Expected path %d to generate vertices even with no items", pathID)
		}
	}

	// Text, inactive circles, and active circle should generate no vertices
	for _, pathID := range []uint{2, 3, 4} {
		ctrl.Rewind(pathID)
		_, _, cmd := ctrl.Vertex()
		if cmd != basics.PathCmdStop {
			t.Errorf("Expected path %d to generate no vertices when no items exist", pathID)
		}
	}
}

func TestInRect(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(10, 20, 110, 120, false)

	// Test points inside the control
	testCases := []struct {
		x, y     float64
		expected bool
		desc     string
	}{
		{50, 70, true, "center point"},
		{10, 20, true, "top-left corner"},
		{110, 120, true, "bottom-right corner"},
		{60, 20, true, "top edge"},
		{60, 120, true, "bottom edge"},
		{10, 70, true, "left edge"},
		{110, 70, true, "right edge"},
		{9, 70, false, "left of control"},
		{111, 70, false, "right of control"},
		{60, 19, false, "above control"},
		{60, 121, false, "below control"},
	}

	for _, tc := range testCases {
		result := ctrl.InRect(tc.x, tc.y)
		if result != tc.expected {
			t.Errorf("InRect(%.1f, %.1f) for %s: expected %t, got %t",
				tc.x, tc.y, tc.desc, tc.expected, result)
		}
	}
}

func TestItemTextAccess(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Test with no items
	if ctrl.ItemText(0) != "" {
		t.Errorf("Expected empty string for non-existent item, got '%s'", ctrl.ItemText(0))
	}

	if ctrl.ItemText(-1) != "" {
		t.Errorf("Expected empty string for negative index, got '%s'", ctrl.ItemText(-1))
	}

	// Add some items
	items := []string{"Alpha", "Beta", "Gamma"}
	for _, item := range items {
		ctrl.AddItem(item)
	}

	// Test valid indices
	for i, expectedText := range items {
		actualText := ctrl.ItemText(i)
		if actualText != expectedText {
			t.Errorf("Expected ItemText(%d) to be '%s', got '%s'", i, expectedText, actualText)
		}
	}

	// Test invalid indices
	if ctrl.ItemText(len(items)) != "" {
		t.Errorf("Expected empty string for out-of-range index, got '%s'", ctrl.ItemText(len(items)))
	}

	if ctrl.ItemText(-1) != "" {
		t.Errorf("Expected empty string for negative index, got '%s'", ctrl.ItemText(-1))
	}
}

func TestDefaultColors(t *testing.T) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)

	// Test that default colors are set and reasonable
	expectedDefaults := []struct {
		pathID uint
		desc   string
	}{
		{0, "background"},
		{1, "border"},
		{2, "text"},
		{3, "inactive"},
		{4, "active"},
	}

	for _, def := range expectedDefaults {
		rgba := ctrl.Color(def.pathID)
		// Verify RGBA values are in valid range [0, 1]
		if rgba.R < 0 || rgba.R > 1 || rgba.G < 0 || rgba.G > 1 ||
			rgba.B < 0 || rgba.B > 1 || rgba.A < 0 || rgba.A > 1 {
			t.Errorf("Default %s color has invalid RGBA values: %v", def.desc, rgba)
		}
	}
}

// Benchmark tests to ensure reasonable performance
func BenchmarkAddItems(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)
		for j := 0; j < 20; j++ { // Add 20 items
			ctrl.AddItem("Test Item")
		}
	}
}

func BenchmarkVertexGeneration(b *testing.B) {
	ctrl := NewDefaultRboxCtrl(0, 0, 100, 100, false)
	for i := 0; i < 10; i++ {
		ctrl.AddItem("Item")
	}
	ctrl.SetCurItem(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for pathID := uint(0); pathID < 5; pathID++ {
			ctrl.Rewind(pathID)
			for {
				_, _, cmd := ctrl.Vertex()
				if cmd == basics.PathCmdStop {
					break
				}
			}
		}
	}
}
