package x11

import (
	"testing"
	"time"
)

// TestDelay verifies that the Delay method provides accurate timing
func TestDelay(t *testing.T) {
	// Create a backend instance (we don't need to initialize it for testing Delay)
	backend := &X11Backend{}

	testCases := []struct {
		name       string
		delayMs    uint32
		tolerance  time.Duration // Allow some tolerance for timing precision
	}{
		{
			name:      "short delay 10ms",
			delayMs:   10,
			tolerance: 5 * time.Millisecond,
		},
		{
			name:      "medium delay 50ms", 
			delayMs:   50,
			tolerance: 5 * time.Millisecond,
		},
		{
			name:      "frame rate delay 16ms (~60 FPS)",
			delayMs:   16,
			tolerance: 5 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			backend.Delay(tc.delayMs)
			elapsed := time.Since(start)

			expected := time.Duration(tc.delayMs) * time.Millisecond
			diff := elapsed - expected

			// Check if the elapsed time is within tolerance
			if diff < -tc.tolerance || diff > tc.tolerance {
				t.Errorf("Delay(%d ms) took %v, expected %v ± %v",
					tc.delayMs, elapsed, expected, tc.tolerance)
			}
		})
	}
}

// TestDelayZero verifies that zero delay works correctly
func TestDelayZero(t *testing.T) {
	backend := &X11Backend{}
	
	start := time.Now()
	backend.Delay(0)
	elapsed := time.Since(start)
	
	// Zero delay should complete very quickly (within 1ms)
	if elapsed > time.Millisecond {
		t.Errorf("Delay(0) took %v, expected to complete quickly", elapsed)
	}
}