//go:build freetype

package freetype

import (
	"testing"
)

func TestNewFontEngineFreetype(t *testing.T) {
	tests := []struct {
		name     string
		flag32   bool
		maxFaces uint
		wantErr  bool
	}{
		{
			name:     "Create 16-bit engine with default faces",
			flag32:   false,
			maxFaces: 0, // Should default to 32
			wantErr:  false,
		},
		{
			name:     "Create 32-bit engine with custom max faces",
			flag32:   true,
			maxFaces: 64,
			wantErr:  false,
		},
		{
			name:     "Create engine with minimal faces",
			flag32:   false,
			maxFaces: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewFontEngineFreetype(tt.flag32, tt.maxFaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFontEngineFreetype() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				defer engine.Close()

				// Verify initial state
				if engine.flag32 != tt.flag32 {
					t.Errorf("Expected flag32 = %v, got %v", tt.flag32, engine.flag32)
				}

				expectedMaxFaces := tt.maxFaces
				if expectedMaxFaces == 0 {
					expectedMaxFaces = 32
				}
				if engine.maxFaces != expectedMaxFaces {
					t.Errorf("Expected maxFaces = %d, got %d", expectedMaxFaces, engine.maxFaces)
				}

				if engine.resolution != 72 {
					t.Errorf("Expected default resolution = 72, got %d", engine.resolution)
				}

				if !engine.hinting {
					t.Errorf("Expected hinting enabled by default")
				}

				if engine.flipY {
					t.Errorf("Expected flipY disabled by default")
				}
			}
		})
	}
}

func TestFontEngineFreetype_BasicProperties(t *testing.T) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Test resolution
	engine.SetResolution(96)
	if engine.resolution != 96 {
		t.Errorf("Expected resolution 96, got %d", engine.resolution)
	}

	// Test height and width
	engine.SetHeight(12.0)
	if expected := 12.0; engine.GetHeight() != expected {
		t.Errorf("Expected height %f, got %f", expected, engine.GetHeight())
	}

	engine.SetWidth(10.0)
	if expected := 10.0; engine.GetWidth() != expected {
		t.Errorf("Expected width %f, got %f", expected, engine.GetWidth())
	}

	// Test hinting
	engine.SetHinting(false)
	if engine.GetHinting() {
		t.Errorf("Expected hinting disabled")
	}

	// Test flip Y
	engine.SetFlipY(true)
	if !engine.GetFlipY() {
		t.Errorf("Expected flipY enabled")
	}

	// Test change stamp increments
	initialStamp := engine.ChangeStamp()
	engine.SetHeight(14.0)
	if engine.ChangeStamp() <= initialStamp {
		t.Errorf("Expected change stamp to increment after SetHeight")
	}
}

func TestFontEngineFreetype_SignatureGeneration(t *testing.T) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Test that signature changes when properties change
	sig1 := engine.FontSignature()

	engine.SetHeight(14.0)
	sig2 := engine.FontSignature()

	if sig1 == sig2 {
		t.Errorf("Expected signature to change after SetHeight, got same signature: %s", sig1)
	}

	engine.SetHinting(false)
	sig3 := engine.FontSignature()

	if sig2 == sig3 {
		t.Errorf("Expected signature to change after SetHinting, got same signature: %s", sig2)
	}

	// Test that signature is consistent for same settings
	sig4 := engine.FontSignature()
	if sig3 != sig4 {
		t.Errorf("Expected consistent signature for same settings, got %s vs %s", sig3, sig4)
	}
}

func TestFontEngineFreetype_GlyphDataTypes(t *testing.T) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Test glyph preparation without a loaded font
	if engine.PrepareGlyph(65) { // 'A'
		t.Errorf("Expected PrepareGlyph to fail without loaded font")
	}

	// Test glyph properties when no glyph is prepared
	if engine.GlyphIndex() != 0 {
		t.Errorf("Expected glyph index 0 when no glyph prepared, got %d", engine.GlyphIndex())
	}

	if engine.DataSize() != 0 {
		t.Errorf("Expected data size 0 when no glyph prepared, got %d", engine.DataSize())
	}

	if engine.AdvanceX() != 0 {
		t.Errorf("Expected advance X 0 when no glyph prepared, got %f", engine.AdvanceX())
	}

	if engine.AdvanceY() != 0 {
		t.Errorf("Expected advance Y 0 when no glyph prepared, got %f", engine.AdvanceY())
	}
}

func TestFontEngineFreetype_KerningWithoutFont(t *testing.T) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Test kerning without loaded font
	dx, dy := engine.AddKerning(65, 86) // 'A', 'V'
	if dx != 0 || dy != 0 {
		t.Errorf("Expected zero kerning without font, got dx=%f, dy=%f", dx, dy)
	}
}

func TestFontEngineFreetype_PathAdaptor(t *testing.T) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	path := engine.PathAdaptor()
	if path == nil {
		t.Errorf("Expected non-nil path adaptor")
	}

	// Test that path is initially empty
	if path.TotalVertices() != 0 {
		t.Errorf("Expected empty path initially, got %d vertices", path.TotalVertices())
	}
}

func TestFontEngineFreetype_Close(t *testing.T) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test that Close() doesn't panic and can be called multiple times
	err1 := engine.Close()
	if err1 != nil {
		t.Errorf("Expected no error on first Close(), got %v", err1)
	}

	err2 := engine.Close()
	if err2 != nil {
		t.Errorf("Expected no error on second Close(), got %v", err2)
	}
}

func TestCalcCRC32(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty input",
			input: []byte{},
		},
		{
			name:  "Single byte",
			input: []byte{0x41}, // 'A'
		},
		{
			name:  "Hello World",
			input: []byte("Hello World"),
		},
		{
			name:  "Font signature example",
			input: []byte("Arial_768_768_true_false_1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result1 := calcCRC32(tt.input)
			result2 := calcCRC32(tt.input)

			// Test consistency - same input should give same output
			if result1 != result2 {
				t.Errorf("calcCRC32(%q) inconsistent: first=0x%08x, second=0x%08x", tt.input, result1, result2)
			}

			// Test that different inputs give different results (when possible)
			if len(tt.input) > 0 {
				modified := make([]byte, len(tt.input))
				copy(modified, tt.input)
				modified[0] ^= 0x01 // Flip a bit

				resultModified := calcCRC32(modified)
				if result1 == resultModified {
					t.Errorf("calcCRC32 gave same result for different inputs: original=0x%08x, modified=0x%08x", result1, resultModified)
				}
			}
		})
	}
}

func TestGlyphRenderingTypeConstants(t *testing.T) {
	// Test that rendering type constants are properly defined
	tests := []struct {
		name     string
		value    GlyphRenderingType
		expected int
	}{
		{"GlyphRenderingNative", GlyphRenderingNative, 0},
		{"GlyphRenderingOutline", GlyphRenderingOutline, 1},
		{"GlyphRenderingAAGray8", GlyphRenderingAAGray8, 2},
		{"GlyphRenderingAAMono", GlyphRenderingAAMono, 3},
		{"GlyphRenderingMono", GlyphRenderingMono, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("Expected %s = %d, got %d", tt.name, tt.expected, int(tt.value))
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkNewFontEngineFreetype(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engine, err := NewFontEngineFreetype(false, 32)
		if err != nil {
			b.Fatalf("Failed to create engine: %v", err)
		}
		engine.Close()
	}
}

func BenchmarkCalcCRC32(b *testing.B) {
	data := []byte("Arial_768_768_true_false_1_someverylongfontpathandparameters")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calcCRC32(data)
	}
}

func BenchmarkFontSignatureGeneration(b *testing.B) {
	engine, err := NewFontEngineFreetype(false, 32)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.updateSignature()
	}
}
