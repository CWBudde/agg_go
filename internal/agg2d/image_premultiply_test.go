package agg2d

import (
	"testing"

	"agg_go/internal/buffer"
)

func TestImagePremultiplyDemultiplyEnhanced(t *testing.T) {
	// Create test image data (RGBA format)
	testData := []uint8{
		128, 64, 32, 128, // Half-transparent pixel
		255, 128, 64, 255, // Opaque pixel
		100, 50, 25, 0, // Transparent pixel
		200, 150, 100, 200, // Another semi-transparent pixel
	}

	img := &Image{
		Data:   make([]uint8, len(testData)),
		width:  2,
		height: 2,
	}
	copy(img.Data, testData)
	img.renBuf = buffer.NewRenderingBuffer[uint8]()

	// Store original data for comparison
	originalData := make([]uint8, len(testData))
	copy(originalData, testData)

	// Test Premultiply
	err := img.Premultiply()
	if err != nil {
		t.Fatalf("Premultiply failed: %v", err)
	}

	// Check premultiplied values
	// First pixel: R=128*128/255=64, G=64*128/255=32, B=32*128/255=16, A=128
	expectedR := uint8(128 * 128 / 255) // Should be 64
	if img.Data[0] != expectedR {
		t.Errorf("Premultiply R: expected %d, got %d", expectedR, img.Data[0])
	}

	// Transparent pixel should have RGB = 0
	if img.Data[8] != 0 || img.Data[9] != 0 || img.Data[10] != 0 {
		t.Errorf("Premultiply transparent pixel: expected RGB=0, got RGB=(%d,%d,%d)",
			img.Data[8], img.Data[9], img.Data[10])
	}
	if img.Data[11] != 0 { // Alpha should remain 0
		t.Errorf("Premultiply transparent pixel alpha: expected 0, got %d", img.Data[11])
	}

	// Test Demultiply
	err = img.Demultiply()
	if err != nil {
		t.Fatalf("Demultiply failed: %v", err)
	}

	// Check that we're close to original values (allowing for rounding errors)
	tolerance := uint8(2) // Allow small rounding errors
	for i := 0; i < len(originalData); i += 4 {
		// Skip transparent pixels (they may not round-trip exactly)
		if originalData[i+3] == 0 {
			continue
		}

		for j := 0; j < 3; j++ { // Check R, G, B (not A)
			diff := int(img.Data[i+j]) - int(originalData[i+j])
			if diff < 0 {
				diff = -diff
			}
			if diff > int(tolerance) {
				t.Errorf("Demultiply pixel %d component %d: expected ~%d, got %d (diff=%d)",
					i/4, j, originalData[i+j], img.Data[i+j], diff)
			}
		}
	}
}

func TestImagePremultiplyErrors(t *testing.T) {
	// Test with nil buffer
	img := &Image{}
	err := img.Premultiply()
	if err == nil {
		t.Errorf("Expected error for nil buffer")
	}

	// Test with nil data
	img.renBuf = buffer.NewRenderingBuffer[uint8]()
	err = img.Premultiply()
	if err == nil {
		t.Errorf("Expected error for nil data")
	}
}

func TestImageDemultiplyErrors(t *testing.T) {
	// Test with nil buffer
	img := &Image{}
	err := img.Demultiply()
	if err == nil {
		t.Errorf("Expected error for nil buffer")
	}

	// Test with nil data
	img.renBuf = buffer.NewRenderingBuffer[uint8]()
	err = img.Demultiply()
	if err == nil {
		t.Errorf("Expected error for nil data")
	}
}

// Benchmark tests
func BenchmarkImagePremultiplyEnhanced(b *testing.B) {
	// Create 100x100 RGBA image
	testData := make([]uint8, 100*100*4)
	for i := 0; i < len(testData); i++ {
		testData[i] = uint8(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := &Image{
			Data:   make([]uint8, len(testData)),
			renBuf: buffer.NewRenderingBuffer[uint8](),
		}
		copy(img.Data, testData)

		_ = img.Premultiply()
	}
}

func BenchmarkImageDemultiplyEnhanced(b *testing.B) {
	// Create 100x100 RGBA image
	testData := make([]uint8, 100*100*4)
	for i := 0; i < len(testData); i++ {
		testData[i] = uint8(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := &Image{
			Data:   make([]uint8, len(testData)),
			renBuf: buffer.NewRenderingBuffer[uint8](),
		}
		copy(img.Data, testData)

		_ = img.Demultiply()
	}
}
