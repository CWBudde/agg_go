package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// TestBlenderInterfaceCompliance tests that all blender types implement the BlenderBase interface
func TestBlenderInterfaceCompliance(t *testing.T) {
	// Test RGBA blenders
	var _ BlenderBase[basics.Int8u, blender.RGBAOrder] = blender.BlenderRGBA[color.Linear, blender.RGBAOrder]{}
	var _ BlenderBase[basics.Int8u, blender.RGBAOrder] = blender.BlenderRGBAPre[color.Linear, blender.RGBAOrder]{}
	var _ BlenderBase[basics.Int8u, blender.RGBAOrder] = blender.BlenderRGBAPlain[color.Linear, blender.RGBAOrder]{}

	// Test RGB blenders
	var _ BlenderBase[basics.Int8u, blender.RGBOrder] = blender.BlenderRGB[color.Linear, blender.RGBOrder]{}
	var _ BlenderBase[basics.Int8u, blender.RGBOrder] = blender.BlenderRGBPre[color.Linear, blender.RGBOrder]{}

	// Test Gray blenders
	var _ BlenderBase[basics.Int8u, blender.GrayOrder] = blender.BlenderGray[color.Linear]{}
	var _ BlenderBase[basics.Int8u, blender.GrayOrder] = blender.BlenderGrayPre[color.Linear]{}

	// Test 16-bit blenders
	var _ BlenderBase[basics.Int16u, blender.RGBAOrder] = blender.BlenderRGBA16{}
	var _ BlenderBase[basics.Int16u, blender.RGBAOrder] = blender.BlenderRGBA16Pre{}
	var _ BlenderBase[basics.Int16u, blender.RGBAOrder] = blender.BlenderRGBA16Plain{}

	var _ BlenderBase[basics.Int16u, blender.GrayOrder] = blender.BlenderGray16[color.Linear]{}
	var _ BlenderBase[basics.Int16u, blender.GrayOrder] = blender.BlenderGray16Pre[color.Linear]{}

	// Test 32-bit gray blenders  
	var _ BlenderBase[basics.Int32u, blender.GrayOrder] = blender.BlenderGray32[color.Linear]{}
	var _ BlenderBase[basics.Int32u, blender.GrayOrder] = blender.BlenderGray32Pre[color.Linear]{}

	// Test packed RGB blenders
	var _ BlenderBase[basics.Int16u, blender.PackedRGB555Order] = blender.BlenderRGB555{}
	var _ BlenderBase[basics.Int16u, blender.PackedRGB555Order] = blender.BlenderRGB555Pre{}
	var _ BlenderBase[basics.Int16u, blender.PackedRGB565Order] = blender.BlenderRGB565{}
	var _ BlenderBase[basics.Int16u, blender.PackedRGB565Order] = blender.BlenderRGB565Pre{}
	var _ BlenderBase[basics.Int16u, blender.PackedBGR555Order] = blender.BlenderBGR555{}
	var _ BlenderBase[basics.Int16u, blender.PackedBGR565Order] = blender.BlenderBGR565{}
}

// TestBlenderInterfaceBehavior tests actual interface method behavior
func TestBlenderInterfaceBehavior(t *testing.T) {
	// Create test data
	pixel := []basics.Int8u{128, 64, 192, 255}
	testColor := color.RGBA{R: 0.5, G: 0.25, B: 0.75, A: 1.0}
	cover := basics.Int8u(255)

	// Test RGBA blender
	blender := blender.BlenderRGBA[color.Linear, blender.RGBAOrder]{}
	var blenderInterface BlenderBase[basics.Int8u, blender.RGBAOrder] = blender

	// Test Get method
	retrievedColor := blenderInterface.Get(pixel, cover)
	expectedR := 128.0 / 255.0
	expectedG := 64.0 / 255.0
	expectedB := 192.0 / 255.0
	expectedA := 255.0 / 255.0

	tolerance := 0.01
	if abs(retrievedColor.R-expectedR) > tolerance ||
		abs(retrievedColor.G-expectedG) > tolerance ||
		abs(retrievedColor.B-expectedB) > tolerance ||
		abs(retrievedColor.A-expectedA) > tolerance {
		t.Errorf("Get method failed: expected (%f,%f,%f,%f), got (%f,%f,%f,%f)",
			expectedR, expectedG, expectedB, expectedA,
			retrievedColor.R, retrievedColor.G, retrievedColor.B, retrievedColor.A)
	}

	// Test GetRaw method
	r, g, b, a := blenderInterface.GetRaw(pixel)
	if r != 128 || g != 64 || b != 192 || a != 255 {
		t.Errorf("GetRaw failed: expected (128,64,192,255), got (%d,%d,%d,%d)", r, g, b, a)
	}

	// Test Set method
	testPixel := make([]basics.Int8u, 4)
	blenderInterface.Set(testPixel, testColor)

	// Verify the set operation
	setR, setG, setB, setA := blenderInterface.GetRaw(testPixel)
	if setR != 127 || setG != 63 || setB != 191 || setA != 255 { // Allow for rounding
		t.Errorf("Set failed: expected approximately (127,63,191,255), got (%d,%d,%d,%d)", setR, setG, setB, setA)
	}

	// Test SetRaw method
	testPixel2 := make([]basics.Int8u, 4)
	blenderInterface.SetRaw(testPixel2, 100, 150, 200, 255)
	checkR, checkG, checkB, checkA := blenderInterface.GetRaw(testPixel2)
	if checkR != 100 || checkG != 150 || checkB != 200 || checkA != 255 {
		t.Errorf("SetRaw failed: expected (100,150,200,255), got (%d,%d,%d,%d)", checkR, checkG, checkB, checkA)
	}
}

// Helper function for absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}