//go:build freetype

package freetype2

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
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
		{12.0, "int16"},   // Small text
		{24.0, "int16"},   // Normal text
		{100.0, "int16"},  // Large text, still within int16 range
		{200.0, "int16"},  // Boundary case
		{201.0, "int32"},  // Just over boundary
		{500.0, "int32"},  // Very large text
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GlyphHeight%.1f", tc.glyphHeight), func(t *testing.T) {
			recommended := RecommendedEngineForGlyphSize(tc.glyphHeight)
			if recommended != tc.expected {
				t.Errorf("For glyph height %.1f, expected %s, got %s", 
					tc.glyphHeight, tc.expected, recommended)
			}
		})
	}
}

// TestFontManager tests the high-level font manager functionality.
func TestFontManager(t *testing.T) {
	fm, err := NewFontManager()
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
	engine, err := NewFontEngineInt16Default()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	
	cm := NewCacheManager2(engine)
	if cm == nil {
		t.Fatal("Failed to create cache manager")
	}
	defer cm.Close()
	
	// Test basic properties
	if cm.fontEngine != engine {
		t.Error("Cache manager should reference the provided engine")
	}
	
	if cm.cachedGlyphs == nil {
		t.Error("Cache manager should have initialized glyph cache")
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
		fm, err := NewFontManager()
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
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",           // Ubuntu/Debian
		"/usr/share/fonts/TTF/DejaVuSans.ttf",                      // Arch Linux
		"/System/Library/Fonts/Arial.ttf",                          // macOS
		"C:/Windows/Fonts/arial.ttf",                               // Windows
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
		fm, err := NewFontManager()
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
	})
	
	t.Run("LoadFromMemory", func(t *testing.T) {
		// Read the font file into memory
		fontData, err := os.ReadFile(fontPath)
		if err != nil {
			t.Fatalf("Failed to read font file: %v", err)
		}
		
		fm, err := NewFontManager()
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
			height   float64
			width    float64
			hinting  bool
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