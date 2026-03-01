//go:build freetype

package freetype2

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"agg_go/internal/fonts"
)

// TestFontEngineCreation tests the basic creation of font engines.
func TestFontEngineCreation(t *testing.T) {
	t.Run("Int16Engine", func(t *testing.T) {
		engine, err := NewFontEngineInt16Default()
		if err != nil {
			t.Fatalf("Failed to create Int16 engine: %v", err)
		}
		defer engine.Close()

		if engine.Is32Bit() {
			t.Error("Int16 engine should not report as 32-bit")
		}
	})

	t.Run("Int32Engine", func(t *testing.T) {
		engine, err := NewFontEngineInt32Default()
		if err != nil {
			t.Fatalf("Failed to create Int32 engine: %v", err)
		}
		defer engine.Close()

		if !engine.Is32Bit() {
			t.Error("Int32 engine should report as 32-bit")
		}
	})
}

// TestFontEngineWithCustomParams tests engine creation with custom parameters.
func TestFontEngineWithCustomParams(t *testing.T) {
	maxFaces := uint32(64)

	t.Run("Int16WithCustomParams", func(t *testing.T) {
		engine, err := NewFontEngineInt16(maxFaces, nil)
		if err != nil {
			t.Fatalf("Failed to create Int16 engine with custom params: %v", err)
		}
		defer engine.Close()

		if engine.FontEngine.maxFaces != maxFaces {
			t.Errorf("Expected max faces %d, got %d", maxFaces, engine.FontEngine.maxFaces)
		}
	})

	t.Run("Int32WithCustomParams", func(t *testing.T) {
		engine, err := NewFontEngineInt32(maxFaces, nil)
		if err != nil {
			t.Fatalf("Failed to create Int32 engine with custom params: %v", err)
		}
		defer engine.Close()

		if engine.FontEngine.maxFaces != maxFaces {
			t.Errorf("Expected max faces %d, got %d", maxFaces, engine.FontEngine.maxFaces)
		}
	})
}

// TestFontEngineRecommendation tests the recommendation system for engine types.
func TestFontEngineRecommendation(t *testing.T) {
	testCases := []struct {
		glyphHeight float64
		expected    string
	}{
		{12.0, "int16"},  // Small text
		{24.0, "int16"},  // Normal text
		{100.0, "int16"}, // Large text, still within int16 range
		{200.0, "int16"}, // Boundary case
		{201.0, "int32"}, // Just over boundary
		{500.0, "int32"}, // Very large text
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GlyphHeight%.1f", tc.glyphHeight), func(t *testing.T) {
			recommended := recommendedEngineForGlyphSize(tc.glyphHeight)
			if recommended != tc.expected {
				t.Errorf("For glyph height %.1f, expected %s, got %s",
					tc.glyphHeight, tc.expected, recommended)
			}
		})
	}
}

// TestFontManager tests the high-level font manager functionality.
// fontManager is a Go convenience wrapper, so these tests intentionally cover
// port-added behavior rather than direct AGG API parity.
func TestFontManager(t *testing.T) {
	fm, err := newFontManager()
	if err != nil {
		t.Fatalf("Failed to create font manager: %v", err)
	}
	defer fm.Close()

	// Test engine switching
	t.Run("EngineSwitching", func(t *testing.T) {
		// Start with default (int16)
		if fm.defaultEngine != "int16" {
			t.Error("Expected default engine to be int16")
		}

		// Switch to int32
		err := fm.SwitchEngine("int32")
		if err != nil {
			t.Fatalf("Failed to switch to int32 engine: %v", err)
		}

		if fm.defaultEngine != "int32" {
			t.Error("Expected current engine to be int32")
		}

		// Try invalid engine
		err = fm.SwitchEngine("invalid")
		if err == nil {
			t.Error("Expected error when switching to invalid engine")
		}
	})

	t.Run("SwitchEngineClearsCurrentFace", func(t *testing.T) {
		fm.currentFace = &LoadedFace{}
		fm.cacheManager.SetCurrentFont(&LoadedFace{})

		err := fm.SwitchEngine("int16")
		if err != nil {
			t.Fatalf("switching back to int16 failed: %v", err)
		}

		if fm.CurrentFace() != nil {
			t.Fatal("expected current face to be cleared after engine switch")
		}
		if fm.GetCacheManager().GetCurrentFont() != nil {
			t.Fatal("expected cache manager font context to be cleared after engine switch")
		}
	})
}

// TestLoadedFaceBasics tests basic LoadedFace functionality without actual font files.
func TestLoadedFaceBasics(t *testing.T) {
	// Create a mock engine for testing
	engine, err := NewFontEngineInt16Default()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Test encoding types
	t.Run("EncodingTypes", func(t *testing.T) {
		encodings := []CharEncoding{
			EncodingNone,
			EncodingMS,
			EncodingUnicode,
			EncodingSymbol,
			EncodingAdobeLatin1,
			EncodingAdobeCustom,
			EncodingAdobeExpert,
		}

		// Just test that the constants are defined and different from each other
		encodingSet := make(map[CharEncoding]bool)
		for _, enc := range encodings {
			if encodingSet[enc] {
				t.Errorf("Duplicate encoding value: %d", int(enc))
			}
			encodingSet[enc] = true
		}

		// Ensure we have the expected number of encodings
		if len(encodingSet) != len(encodings) {
			t.Errorf("Expected %d unique encodings, got %d", len(encodings), len(encodingSet))
		}
	})
}

// TestGlyphRendering tests glyph rendering mode functionality.
func TestGlyphRendering(t *testing.T) {
	testCases := []struct {
		rendering GlyphRendering
		expected  string
	}{
		{GlyphRenNativeMono, "native_mono"},
		{GlyphRenNativeGray8, "native_gray8"},
		{GlyphRenOutline, "outline"},
		{GlyphRenAggMono, "agg_mono"},
		{GlyphRenAggGray8, "agg_gray8"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.rendering.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestCacheManager2 tests the cache manager integration.
func TestCacheManager2(t *testing.T) {
	testCases := []struct {
		name string
		new  func() (FontEngineInterface, error)
	}{
		{
			name: "Int16",
			new: func() (FontEngineInterface, error) {
				return NewFontEngineInt16Default()
			},
		},
		{
			name: "Int32",
			new: func() (FontEngineInterface, error) {
				return NewFontEngineInt32Default()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine, err := tc.new()
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}
			defer engine.Close()

			cm := NewCacheManager2(engine)
			if cm == nil {
				t.Fatal("Failed to create cache manager")
			}
			defer cm.Close()

			if cm.fontEngine != engine {
				t.Error("Cache manager should reference the provided engine")
			}
			if cm.currentCachedFont != nil {
				t.Error("Cache manager should defer cached-font creation until a face is selected")
			}
			if cm.PathAdaptor() == nil {
				t.Error("Cache manager should initialize a path adaptor")
			}
			if cm.Gray8Adaptor() == nil {
				t.Error("Cache manager should initialize a gray8 adaptor")
			}
			if cm.MonoAdaptor() == nil {
				t.Error("Cache manager should initialize a mono adaptor")
			}
		})
	}
}

func TestFaceAndEngineCloseOwnership(t *testing.T) {
	fontPath := findTestFont()
	if fontPath == "" {
		t.Skip("No test font found on system")
		return
	}

	engine, err := NewFontEngineInt16Default()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	face, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		engine.Close()
		t.Fatalf("Failed to load font: %v", err)
	}

	if err := face.Close(); err != nil {
		engine.Close()
		t.Fatalf("face.Close() failed: %v", err)
	}
	if err := face.Close(); err != nil {
		engine.Close()
		t.Fatalf("face.Close() should be idempotent: %v", err)
	}
	if got := len(engine.loadedFaces); got != 0 {
		engine.Close()
		t.Fatalf("expected engine to release face after Close, got %d faces", got)
	}
	if err := engine.Close(); err != nil {
		t.Fatalf("engine.Close() failed after face.Close(): %v", err)
	}
	if err := engine.Close(); err != nil {
		t.Fatalf("engine.Close() should be idempotent: %v", err)
	}
}

func TestCacheManager2DoesNotOwnEngine(t *testing.T) {
	engine, err := NewFontEngineInt16Default()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	cm := NewCacheManager2(engine)
	if cm == nil {
		t.Fatal("Failed to create cache manager")
	}
	if err := cm.Close(); err != nil {
		t.Fatalf("cache manager close failed: %v", err)
	}
	if engine.library == nil {
		t.Fatal("cache manager close should not close the engine")
	}
}

func TestCacheManager2GlyphTracksFaceInstanceContext(t *testing.T) {
	fontPath := findTestFont()
	if fontPath == "" {
		t.Skip("No test font found on system")
		return
	}

	engine, err := NewFontEngineInt16Default()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	faceIface, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	defer faceIface.Close()

	face, ok := faceIface.(*LoadedFace)
	if !ok {
		t.Fatal("expected concrete LoadedFace")
	}

	cm := NewCacheManager2(engine)
	defer cm.Close()
	cm.SetCurrentFont(face)

	face.SelectInstance(24.0, 24.0, true, GlyphRenNativeGray8)
	grayGlyph := cm.Glyph(uint32('A'))
	if grayGlyph == nil {
		t.Fatal("expected gray glyph")
	}
	if grayGlyph.DataType != fonts.FmanGlyphDataGray8 {
		t.Fatalf("unexpected gray glyph data type: %v", grayGlyph.DataType)
	}
	if len(grayGlyph.Data) == 0 {
		t.Fatal("expected serialized gray glyph data")
	}
	cm.Gray8Adaptor().InitGlyph(grayGlyph.Data, grayGlyph.DataSize, 0, 0)
	cm.Gray8Adaptor().Rewind(0)
	if !cm.Gray8Adaptor().SweepScanline() {
		t.Fatal("expected serialized gray glyph data to be readable")
	}

	face.SelectInstance(24.0, 24.0, true, GlyphRenOutline)
	outlineGlyph := cm.Glyph(uint32('A'))
	if outlineGlyph == nil {
		t.Fatal("expected outline glyph after rendering-mode switch")
	}
	if outlineGlyph.DataType != fonts.FmanGlyphDataOutline {
		t.Fatalf("unexpected outline glyph data type: %v", outlineGlyph.DataType)
	}
	if grayGlyph.DataType == outlineGlyph.DataType {
		t.Fatal("expected face-instance change to produce a different cached glyph format")
	}
	if len(outlineGlyph.Data) == 0 {
		t.Fatal("expected serialized outline glyph data")
	}
	adaptor := cm.PathAdaptor()
	adaptor.Init(outlineGlyph.Data, 0, 0, 1.0, 6)
	adaptor.Rewind(0)
	_, _, cmd := adaptor.Vertex()
	if cmd == 0 {
		t.Fatal("expected serialized outline data to be readable")
	}
}

func TestMultiFaceUnloadCompactsLoadedFaces(t *testing.T) {
	fontPath := findTestFont()
	if fontPath == "" {
		t.Skip("No test font found on system")
		return
	}

	engine, err := NewFontEngineInt16(3, nil)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	face1, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		t.Fatalf("failed to load first face: %v", err)
	}
	face2, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		t.Fatalf("failed to load second face: %v", err)
	}
	face3, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		t.Fatalf("failed to load third face: %v", err)
	}

	if got := len(engine.loadedFaces); got != 3 {
		t.Fatalf("expected 3 loaded faces, got %d", got)
	}

	if err := face2.Close(); err != nil {
		t.Fatalf("failed to close middle face: %v", err)
	}

	if got := len(engine.loadedFaces); got != 2 {
		t.Fatalf("expected 2 loaded faces after close, got %d", got)
	}

	firstLoaded, ok := face1.(*LoadedFace)
	if !ok {
		t.Fatal("expected first face to be concrete LoadedFace")
	}
	thirdLoaded, ok := face3.(*LoadedFace)
	if !ok {
		t.Fatal("expected third face to be concrete LoadedFace")
	}

	if engine.loadedFaces[0] != firstLoaded {
		t.Fatal("expected first loaded face to remain in slot 0")
	}
	if engine.loadedFaces[1] != thirdLoaded {
		t.Fatal("expected third loaded face to compact into slot 1")
	}

	face4, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		t.Fatalf("failed to load replacement face: %v", err)
	}
	defer face4.Close()

	if got := len(engine.loadedFaces); got != 3 {
		t.Fatalf("expected replacement load to restore 3 faces, got %d", got)
	}
}

func TestEngineCloseReleasesAllLoadedFaces(t *testing.T) {
	fontPath := findTestFont()
	if fontPath == "" {
		t.Skip("No test font found on system")
		return
	}

	engine, err := NewFontEngineInt16(3, nil)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	faces := make([]LoadedFaceInterface, 0, 3)
	for i := 0; i < 3; i++ {
		face, loadErr := engine.LoadFaceFile(fontPath)
		if loadErr != nil {
			engine.Close()
			t.Fatalf("failed to load face %d: %v", i, loadErr)
		}
		faces = append(faces, face)
	}

	if got := len(engine.loadedFaces); got != 3 {
		engine.Close()
		t.Fatalf("expected 3 loaded faces before engine close, got %d", got)
	}

	if err := engine.Close(); err != nil {
		t.Fatalf("engine.Close() failed: %v", err)
	}

	if got := len(engine.loadedFaces); got != 0 {
		t.Fatalf("expected all loaded faces to be released, got %d", got)
	}

	for i, face := range faces {
		loaded, ok := face.(*LoadedFace)
		if !ok {
			t.Fatalf("expected concrete LoadedFace for face %d", i)
		}
		if loaded.engine != nil {
			t.Fatalf("expected engine reference cleared for face %d", i)
		}
		if loaded.ftFace != nil {
			t.Fatalf("expected FreeType face released for face %d", i)
		}
		if err := face.Close(); err != nil {
			t.Fatalf("face.Close() should remain idempotent after engine.Close() for face %d: %v", i, err)
		}
	}
}

// Benchmark tests for performance evaluation

func BenchmarkEngineCreation(b *testing.B) {
	b.Run("Int16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine, err := NewFontEngineInt16Default()
			if err != nil {
				b.Fatal(err)
			}
			engine.Close()
		}
	})

	b.Run("Int32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine, err := NewFontEngineInt32Default()
			if err != nil {
				b.Fatal(err)
			}
			engine.Close()
		}
	})
}

func BenchmarkFontManager(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fm, err := newFontManager()
		if err != nil {
			b.Fatal(err)
		}
		fm.Close()
	}
}

// Integration tests that would work with actual font files

// findTestFont tries to find a font file for testing.
// Returns empty string if no suitable font is found.
func findTestFont() string {
	// Common font locations on different systems
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",                 // Ubuntu/Debian
		"/usr/share/fonts/TTF/DejaVuSans.ttf",                             // Arch Linux
		"/System/Library/Fonts/Arial.ttf",                                 // macOS
		"C:/Windows/Fonts/arial.ttf",                                      // Windows
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf", // CentOS/RHEL
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Try to find any TTF font in common directories
	fontDirs := []string{
		"/usr/share/fonts",
		"/System/Library/Fonts",
		"C:/Windows/Fonts",
	}

	for _, dir := range fontDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue searching
			}
			if filepath.Ext(path) == ".ttf" {
				return filepath.SkipDir // Found one, stop searching
			}
			return nil
		})
		if err == filepath.SkipDir {
			return dir
		}
	}

	return ""
}

// TestWithActualFont tests font loading with a real font file (if available).
func TestWithActualFont(t *testing.T) {
	fontPath := findTestFont()
	if fontPath == "" {
		t.Skip("No test font found on system")
		return
	}

	t.Run("LoadFromFile", func(t *testing.T) {
		fm, err := newFontManager()
		if err != nil {
			t.Fatalf("Failed to create font manager: %v", err)
		}
		defer fm.Close()

		face, err := fm.LoadFont(fontPath, "")
		if err != nil {
			t.Fatalf("Failed to load font %s: %v", fontPath, err)
		}
		defer face.Close()

		// Test basic face properties
		if face.Name() == "" {
			t.Error("Font face should have a name")
		}

		if face.NumFaces() == 0 {
			t.Error("Font should have at least one face")
		}
		if fm.CurrentFace() != face {
			t.Error("font manager should track the currently loaded face")
		}
		if fm.GetCacheManager().GetCurrentFont() != face {
			t.Error("cache manager should be bound to the currently loaded face")
		}
	})

	t.Run("LoadFromMemory", func(t *testing.T) {
		// Read the font file into memory
		fontData, err := os.ReadFile(fontPath)
		if err != nil {
			t.Fatalf("Failed to read font file: %v", err)
		}

		fm, err := newFontManager()
		if err != nil {
			t.Fatalf("Failed to create font manager: %v", err)
		}
		defer fm.Close()

		face, err := fm.LoadFontFromMemory(fontData, "")
		if err != nil {
			t.Fatalf("Failed to load font from memory: %v", err)
		}
		defer face.Close()

		// Test basic face properties
		if face.Name() == "" {
			t.Error("Font face should have a name")
		}
		if fm.CurrentFace() != face {
			t.Error("font manager should track the currently loaded in-memory face")
		}
		if fm.GetCacheManager().GetCurrentFont() != face {
			t.Error("cache manager should be bound to the currently loaded in-memory face")
		}
	})
}

// TestFaceInstance tests face instance selection and configuration.
func TestFaceInstance(t *testing.T) {
	fontPath := findTestFont()
	if fontPath == "" {
		t.Skip("No test font found on system")
		return
	}

	engine, err := NewFontEngineInt16Default()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	face, err := engine.LoadFaceFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	defer face.Close()

	// Test instance selection
	t.Run("InstanceSelection", func(t *testing.T) {
		// Test different configurations
		configs := []struct {
			height    float64
			width     float64
			hinting   bool
			rendering GlyphRendering
		}{
			{12.0, 12.0, true, GlyphRenNativeGray8},
			{24.0, 24.0, false, GlyphRenOutline},
			{36.0, 36.0, true, GlyphRenAggGray8},
		}

		for i, config := range configs {
			t.Run(fmt.Sprintf("Config%d", i), func(t *testing.T) {
				face.SelectInstance(config.height, config.width, config.hinting, config.rendering)

				if face.Height() != config.height {
					t.Errorf("Expected height %.1f, got %.1f", config.height, face.Height())
				}

				if face.Width() != config.width {
					t.Errorf("Expected width %.1f, got %.1f", config.width, face.Width())
				}

				if face.Hinting() != config.hinting {
					t.Errorf("Expected hinting %t, got %t", config.hinting, face.Hinting())
				}
			})
		}
	})

	// Test glyph preparation
	t.Run("GlyphPreparation", func(t *testing.T) {
		face.SelectInstance(24.0, 24.0, true, GlyphRenNativeGray8)

		// Test common ASCII characters
		testChars := []uint32{'A', 'B', 'a', 'b', '1', '2', ' '}

		for _, char := range testChars {
			prepared, ok := face.PrepareGlyph(char)
			if !ok {
				t.Errorf("Failed to prepare glyph for character %c (U+%04X)", rune(char), char)
				continue
			}

			if prepared.GlyphCode != char {
				t.Errorf("Glyph code mismatch: expected %d, got %d", char, prepared.GlyphCode)
			}

			// Basic sanity checks
			if prepared.GlyphIndex == 0 && char != ' ' { // Space might have index 0
				t.Errorf("Invalid glyph index 0 for character %c", rune(char))
			}
		}
	})

	t.Run("SerializedGlyphCaches", func(t *testing.T) {
		concreteFace, ok := face.(*LoadedFace)
		if !ok {
			t.Fatal("expected concrete LoadedFace")
		}

		testCases := []struct {
			name      string
			rendering GlyphRendering
			validate  func(t *testing.T, engine *FontEngineInt16, data []byte, prepared *PreparedGlyph)
		}{
			{
				name:      "NativeGray8",
				rendering: GlyphRenNativeGray8,
				validate: func(t *testing.T, engine *FontEngineInt16, data []byte, prepared *PreparedGlyph) {
					if prepared.DataType != fonts.FmanGlyphDataGray8 {
						t.Fatalf("unexpected data type: %v", prepared.DataType)
					}
					if prepared.DataSize == 0 {
						t.Fatal("expected serialized native gray8 scanlines")
					}
					adaptor := engine.AdaptorTypes().Gray8Adaptor
					adaptor.Init(data, int(prepared.DataSize), 0, 0)
					if !adaptor.RewindScanlines() {
						t.Fatal("expected serialized native gray8 scanline data")
					}
				},
			},
			{
				name:      "NativeMono",
				rendering: GlyphRenNativeMono,
				validate: func(t *testing.T, engine *FontEngineInt16, data []byte, prepared *PreparedGlyph) {
					if prepared.DataType != fonts.FmanGlyphDataMono {
						t.Fatalf("unexpected data type: %v", prepared.DataType)
					}
					if prepared.DataSize == 0 {
						t.Fatal("expected serialized native mono scanlines")
					}
					adaptor := engine.AdaptorTypes().MonoAdaptor
					adaptor.Init(data, 0, 0)
					if !adaptor.RewindScanlines() {
						t.Fatal("expected serialized native mono scanline data")
					}
				},
			},
			{
				name:      "Outline",
				rendering: GlyphRenOutline,
				validate: func(t *testing.T, engine *FontEngineInt16, data []byte, prepared *PreparedGlyph) {
					if prepared.DataType != fonts.FmanGlyphDataOutline {
						t.Fatalf("unexpected data type: %v", prepared.DataType)
					}
					if prepared.DataSize == 0 {
						t.Fatal("expected serialized outline data")
					}
					adaptor := engine.PathAdaptor()
					adaptor.Init(data, 0, 0, 1.0, 6)
					adaptor.Rewind(0)
					_, _, cmd := adaptor.Vertex()
					if cmd == 0 {
						t.Fatal("expected serialized outline vertex data")
					}
				},
			},
			{
				name:      "AggGray8",
				rendering: GlyphRenAggGray8,
				validate: func(t *testing.T, engine *FontEngineInt16, data []byte, prepared *PreparedGlyph) {
					if prepared.DataType != fonts.FmanGlyphDataGray8 {
						t.Fatalf("unexpected data type: %v", prepared.DataType)
					}
					if prepared.DataSize == 0 {
						t.Fatal("expected serialized gray8 scanlines")
					}
					adaptor := engine.AdaptorTypes().Gray8Adaptor
					adaptor.Init(data, int(prepared.DataSize), 0, 0)
					if !adaptor.RewindScanlines() {
						t.Fatal("expected serialized gray8 scanline data")
					}
				},
			},
			{
				name:      "AggMono",
				rendering: GlyphRenAggMono,
				validate: func(t *testing.T, engine *FontEngineInt16, data []byte, prepared *PreparedGlyph) {
					if prepared.DataType != fonts.FmanGlyphDataMono {
						t.Fatalf("unexpected data type: %v", prepared.DataType)
					}
					if prepared.DataSize == 0 {
						t.Fatal("expected serialized mono scanlines")
					}
					adaptor := engine.AdaptorTypes().MonoAdaptor
					adaptor.Init(data, 0, 0)
					if !adaptor.RewindScanlines() {
						t.Fatal("expected serialized mono scanline data")
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				concreteFace.SelectInstance(24.0, 24.0, true, tc.rendering)
				prepared, ok := concreteFace.PrepareGlyph(uint32('A'))
				if !ok {
					t.Fatalf("failed to prepare glyph for rendering mode %v", tc.rendering)
				}

				data := make([]byte, prepared.DataSize)
				concreteFace.WriteGlyphTo(prepared, data)
				tc.validate(t, engine, data, prepared)
			})
		}
	})

	// Test kerning
	t.Run("Kerning", func(t *testing.T) {
		face.SelectInstance(24.0, 24.0, true, GlyphRenNativeGray8)

		// Test common kerning pairs
		pairs := []struct {
			first, second uint32
		}{
			{'A', 'V'},
			{'T', 'o'},
			{'W', 'a'},
		}

		for _, pair := range pairs {
			dx, dy := face.AddKerning(pair.first, pair.second)
			// We don't know if the font has kerning, but the call should not crash
			// and should return reasonable values
			if dx < -100 || dx > 100 || dy < -100 || dy > 100 {
				t.Errorf("Suspicious kerning values for %c%c: dx=%.2f, dy=%.2f",
					rune(pair.first), rune(pair.second), dx, dy)
			}
		}
	})
}
