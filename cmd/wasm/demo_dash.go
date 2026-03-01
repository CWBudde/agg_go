// Based on the original AGG examples: conv_dash_marker.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/path"
)

var (
	dashTriangleX = [3]float64{157, 469, 243}
	dashTriangleY = [3]float64{160, 270, 410}
	dashWidth     = 3.0
	dashSmooth    = 1.0
	dashClosed    = true
	dashSelected  = -1
	dashDragDX    = 0.0
	dashDragDY    = 0.0
)

func drawDashDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Static patterns preview (at the top)
	agg2d.LineColor(agg.NewColor(100, 100, 100, 255))
	agg2d.NoFill()
	patterns := [][]float64{
		{10, 5},
		{20, 10, 5, 5},
		{5, 10},
	}
	for i, p := range patterns {
		y := 30.0 + float64(i)*30.0
		agg2d.LineWidth(2.0)
		agg2d.RemoveAllDashes()
		for j := 0; j < len(p); j += 2 {
			agg2d.AddDash(p[j], p[j+1])
		}
		agg2d.Line(50, y, 750, y)
	}

	// 2. Interactive Polygon
	ps := path.NewPathStorageStl()
	ps.MoveTo(dashTriangleX[0], dashTriangleY[0])
	ps.LineTo(dashTriangleX[1], dashTriangleY[1])
	// Add a middle point like in the original demo
	midX := (dashTriangleX[0] + dashTriangleX[1] + dashTriangleX[2]) / 3.0
	midY := (dashTriangleY[0] + dashTriangleY[1] + dashTriangleY[2]) / 3.0
	ps.LineTo(midX, midY)
	ps.LineTo(dashTriangleX[2], dashTriangleY[2])
	
	if dashClosed {
		ps.ClosePolygon(basics.PathFlagsNone)
	}

	// 3. Draw filled semi-transparent background
	agg2d.FillColor(agg.NewColor(200, 150, 50, 100))
	agg2d.NoLine()
	agg2d.ResetPath()
	// Manual copy from PathStorage to Agg2D path
	ps.Rewind(0)
	for {
		x, y, cmd := ps.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsMoveTo(basics.PathCommand(cmd)) {
			agg2d.MoveTo(x, y)
		} else {
			agg2d.LineTo(x, y)
		}
	}
	if dashClosed {
		agg2d.ClosePolygon()
	}
	agg2d.DrawPath(agg.FillOnly)

	// 4. Draw dashed outline
	agg2d.NoFill()
	agg2d.LineColor(agg.Black)
	agg2d.LineWidth(dashWidth)
	agg2d.RemoveAllDashes()
	agg2d.AddDash(20, 5)
	agg2d.AddDash(5, 5)
	
	agg2d.ResetPath()
	ps.Rewind(0)
	for {
		x, y, cmd := ps.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsMoveTo(basics.PathCommand(cmd)) {
			agg2d.MoveTo(x, y)
		} else {
			agg2d.LineTo(x, y)
		}
	}
	if dashClosed {
		agg2d.ClosePolygon()
	}
	agg2d.DrawPath(agg.StrokeOnly)

	// 5. Draw interactive handles
	for i := 0; i < 3; i++ {
		agg2d.FillColor(agg.NewColor(200, 50, 20, 150))
		agg2d.NoLine()
		agg2d.FillCircle(dashTriangleX[i], dashTriangleY[i], 6)
		agg2d.LineColor(agg.Black)
		agg2d.LineWidth(1.0)
		agg2d.DrawCircle(dashTriangleX[i], dashTriangleY[i], 6)
	}
}

func handleDashMouseDown(x, y float64) bool {
	dashSelected = -1
	for i := 0; i < 3; i++ {
		dist := math.Sqrt(math.Pow(x-dashTriangleX[i], 2) + math.Pow(y-dashTriangleY[i], 2))
		if dist < 15 {
			dashSelected = i
			dashDragDX = x - dashTriangleX[i]
			dashDragDY = y - dashTriangleY[i]
			return true
		}
	}
	return false
}

func handleDashMouseMove(x, y float64) bool {
	if dashSelected != -1 {
		dashTriangleX[dashSelected] = x - dashDragDX
		dashTriangleY[dashSelected] = y - dashDragDY
		return true
	}
	return false
}

func handleDashMouseUp() {
	dashSelected = -1
}
