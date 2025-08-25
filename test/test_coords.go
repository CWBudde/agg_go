package main

import (
	"fmt"

	"agg_go/internal/basics"
	"agg_go/internal/path"
	"agg_go/internal/rasterizer"
)

func main() {
	fmt.Println("Testing coordinate conversion...")

	// Create rasterizer
	clipper := &rasterizer.RasterizerSlNoClip{}
	ras := rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl](1000, clipper)

	// Create a simple triangle path
	pathStorage := path.NewPathStorage()
	pathStorage.MoveTo(10.0, 10.0)
	pathStorage.LineTo(50.0, 10.0)
	pathStorage.LineTo(30.0, 50.0)
	pathStorage.ClosePolygon(basics.PathFlag(basics.PathFlagClose))

	// Reset rasterizer and add path
	ras.Reset()

	// Manually add vertices to rasterizer
	pathStorage.Rewind(0)
	vertexCount := 0
	for {
		x, y, cmd := pathStorage.NextVertex()
		if cmd == 0 { // PathCmdStop
			break
		}

		fmt.Printf("Vertex %d: (%.2f, %.2f) cmd=%d\n", vertexCount, x, y, cmd)
		vertexCount++

		switch cmd {
		case 1: // PathCmdMoveTo
			ras.MoveToD(x, y)
		case 2: // PathCmdLineTo
			ras.LineToD(x, y)
		case 6: // PathCmdEndPoly | PathFlagClose
			ras.ClosePolygon()
		}

		// Safety check to prevent infinite loops
		if vertexCount > 100 {
			fmt.Printf("ERROR: too many vertices, possible infinite loop\n")
			return
		}
	}

	fmt.Printf("Added %d vertices to rasterizer\n", vertexCount)

	// Test basic properties - these should now be reasonable
	fmt.Printf("Rasterizer bounds: minX=%d, maxX=%d, minY=%d, maxY=%d\n",
		ras.MinX(), ras.MaxX(), ras.MinY(), ras.MaxY())

	// Check if bounds are reasonable
	if ras.MinX() > ras.MaxX() || ras.MinY() > ras.MaxY() {
		fmt.Printf("ERROR: Invalid bounds (min > max)\n")
		return
	}

	if ras.MinX() < -10000 || ras.MaxX() > 50000 || ras.MinY() < -10000 || ras.MaxY() > 50000 {
		fmt.Printf("ERROR: Bounds seem unreasonable\n")
		return
	}

	// Test rewind scanlines
	fmt.Println("Testing RewindScanlines...")
	if !ras.RewindScanlines() {
		fmt.Printf("ERROR: RewindScanlines returned false - no cells to process\n")
		return
	}

	fmt.Println("RewindScanlines succeeded")

	// Test bounds after processing
	fmt.Printf("Bounds after processing: minX=%d, maxX=%d, minY=%d, maxY=%d\n",
		ras.MinX(), ras.MaxX(), ras.MinY(), ras.MaxY())

	// Test a simple hit test
	hit := ras.HitTest(25, 25) // Should be inside the triangle
	fmt.Printf("Hit test at (25, 25): %t\n", hit)

	// Test pixel coordinates that should be inside (in scaled coordinates)
	scaledX := int(25 * basics.PolySubpixelScale)
	scaledY := int(25 * basics.PolySubpixelScale)
	scaledHit := ras.HitTest(scaledX, scaledY)
	fmt.Printf("Hit test at scaled coordinates (%d, %d): %t\n", scaledX, scaledY, scaledHit)

	fmt.Println("Coordinate test completed successfully!")
}