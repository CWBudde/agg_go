package main

import (
	"fmt"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

func main() {
	ld := liondemo.Parse()
	minX, minY, maxX, maxY := 1e9, 1e9, -1e9, -1e9
	for idx := uint(0); idx < ld.Path.TotalVertices(); idx++ {
		x, y, cmd := ld.Path.Vertex(idx)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
		}
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	fmt.Printf("Lion bounds: X=[%.1f, %.1f] Y=[%.1f, %.1f]\n", minX, maxX, minY, maxY)
	fmt.Printf("Center: (%.1f, %.1f)\n", (minX+maxX)/2, (minY+maxY)/2)
	fmt.Printf("base_dx=%.1f, base_dy=%.1f\n", (maxX-minX)/2, (maxY-minY)/2)
}
