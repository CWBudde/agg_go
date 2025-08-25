package agg2d

import (
	"testing"
)

func TestBlendModeConstants(t *testing.T) {
	tests := []struct {
		mode     BlendMode
		expected string
	}{
		{BlendAlpha, "Alpha"},
		{BlendClear, "Clear"},
		{BlendSrc, "Src"},
		{BlendDst, "Dst"},
		{BlendSrcOver, "SrcOver"},
		{BlendDstOver, "DstOver"},
		{BlendSrcIn, "SrcIn"},
		{BlendDstIn, "DstIn"},
		{BlendSrcOut, "SrcOut"},
		{BlendDstOut, "DstOut"},
		{BlendSrcAtop, "SrcAtop"},
		{BlendDstAtop, "DstAtop"},
		{BlendXor, "Xor"},
		{BlendAdd, "Add"},
		{BlendMultiply, "Multiply"},
		{BlendScreen, "Screen"},
		{BlendOverlay, "Overlay"},
		{BlendDarken, "Darken"},
		{BlendLighten, "Lighten"},
		{BlendColorDodge, "ColorDodge"},
		{BlendColorBurn, "ColorBurn"},
		{BlendHardLight, "HardLight"},
		{BlendSoftLight, "SoftLight"},
		{BlendDifference, "Difference"},
		{BlendExclusion, "Exclusion"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := BlendModeString(test.mode)
			if result != test.expected {
				t.Errorf("BlendModeString(%v) = %s, want %s", test.mode, result, test.expected)
			}
		})
	}
}

func TestBlendModeStringUnknown(t *testing.T) {
	unknown := BlendMode(999)
	result := BlendModeString(unknown)
	if result != "Unknown" {
		t.Errorf("BlendModeString(%v) = %s, want Unknown", unknown, result)
	}
}

func TestAgg2DBlendModeOperations(t *testing.T) {
	agg2d := NewAgg2D()

	// Test default blend mode
	if agg2d.GetBlendMode() != BlendAlpha {
		t.Errorf("Default blend mode should be BlendAlpha, got %v", agg2d.GetBlendMode())
	}

	// Test setting blend mode
	agg2d.SetBlendMode(BlendMultiply)
	if agg2d.GetBlendMode() != BlendMultiply {
		t.Errorf("Expected BlendMultiply, got %v", agg2d.GetBlendMode())
	}

	// Test image blend mode
	agg2d.SetImageBlendMode(BlendScreen)
	if agg2d.GetImageBlendMode() != BlendScreen {
		t.Errorf("Expected BlendScreen, got %v", agg2d.GetImageBlendMode())
	}
}

func TestImageBlendColor(t *testing.T) {
	agg2d := NewAgg2D()

	// Test setting image blend color
	testColor := Color{128, 64, 192, 255}
	agg2d.SetImageBlendColor(testColor)

	result := agg2d.GetImageBlendColor()
	if result != testColor {
		t.Errorf("Expected %v, got %v", testColor, result)
	}

	// Test setting image blend color with RGBA
	agg2d.SetImageBlendColorRGBA(255, 128, 64, 200)
	expected := Color{255, 128, 64, 200}
	result = agg2d.GetImageBlendColor()
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestBlendModeClassification(t *testing.T) {
	// Test Porter-Duff modes
	porterDuffModes := []BlendMode{
		BlendClear, BlendSrc, BlendDst, BlendSrcOver, BlendDstOver,
		BlendSrcIn, BlendDstIn, BlendSrcOut, BlendDstOut,
		BlendSrcAtop, BlendDstAtop, BlendXor,
	}

	for _, mode := range porterDuffModes {
		if !IsPorterDuffMode(mode) {
			t.Errorf("Expected %s to be Porter-Duff mode", BlendModeString(mode))
		}
		if IsExtendedBlendMode(mode) {
			t.Errorf("Expected %s to not be extended blend mode", BlendModeString(mode))
		}
	}

	// Test extended blend modes
	extendedModes := []BlendMode{
		BlendAdd, BlendMultiply, BlendScreen, BlendOverlay,
		BlendDarken, BlendLighten, BlendColorDodge, BlendColorBurn,
		BlendHardLight, BlendSoftLight, BlendDifference, BlendExclusion,
	}

	for _, mode := range extendedModes {
		if IsPorterDuffMode(mode) {
			t.Errorf("Expected %s to not be Porter-Duff mode", BlendModeString(mode))
		}
		if !IsExtendedBlendMode(mode) {
			t.Errorf("Expected %s to be extended blend mode", BlendModeString(mode))
		}
	}

	// Test alpha blend mode (special case)
	if IsPorterDuffMode(BlendAlpha) {
		t.Errorf("BlendAlpha should not be classified as Porter-Duff")
	}
	if IsExtendedBlendMode(BlendAlpha) {
		t.Errorf("BlendAlpha should not be classified as extended")
	}
}

func TestPremultipliedAlphaRequirement(t *testing.T) {
	// Porter-Duff modes should require premultiplied alpha
	porterDuffModes := []BlendMode{
		BlendClear, BlendSrc, BlendDst, BlendSrcOver, BlendDstOver,
		BlendSrcIn, BlendDstIn, BlendSrcOut, BlendDstOut,
		BlendSrcAtop, BlendDstAtop, BlendXor,
	}

	for _, mode := range porterDuffModes {
		if !RequiresPremultipliedAlpha(mode) {
			t.Errorf("Expected %s to require premultiplied alpha", BlendModeString(mode))
		}
	}

	// Alpha blend mode should require premultiplied alpha
	if !RequiresPremultipliedAlpha(BlendAlpha) {
		t.Errorf("BlendAlpha should require premultiplied alpha")
	}

	// Some extended modes might not require premultiplied alpha
	if RequiresPremultipliedAlpha(BlendAdd) {
		t.Errorf("BlendAdd should not require premultiplied alpha")
	}
}

func TestValidateBlendMode(t *testing.T) {
	// Test valid modes
	validModes := []BlendMode{
		BlendAlpha, BlendClear, BlendSrc, BlendDst, BlendSrcOver,
		BlendAdd, BlendMultiply, BlendExclusion,
	}

	for _, mode := range validModes {
		if !ValidateBlendMode(mode) {
			t.Errorf("Expected %s to be valid", BlendModeString(mode))
		}
	}

	// Test invalid modes
	invalidModes := []BlendMode{
		BlendMode(-1), BlendMode(1000),
	}

	for _, mode := range invalidModes {
		if ValidateBlendMode(mode) {
			t.Errorf("Expected %v to be invalid", mode)
		}
	}
}

func TestGetDefaultBlendMode(t *testing.T) {
	defaultMode := GetDefaultBlendMode()
	if defaultMode != BlendAlpha {
		t.Errorf("Expected default blend mode to be BlendAlpha, got %v", defaultMode)
	}
}

// Benchmark blend mode string conversion
func BenchmarkBlendModeString(b *testing.B) {
	modes := []BlendMode{
		BlendAlpha, BlendClear, BlendMultiply, BlendScreen,
		BlendOverlay, BlendExclusion,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mode := modes[i%len(modes)]
		_ = BlendModeString(mode)
	}
}

// Benchmark blend mode validation
func BenchmarkValidateBlendMode(b *testing.B) {
	modes := []BlendMode{
		BlendAlpha, BlendClear, BlendMultiply, BlendScreen,
		BlendOverlay, BlendExclusion, BlendMode(-1), BlendMode(1000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mode := modes[i%len(modes)]
		_ = ValidateBlendMode(mode)
	}
}
