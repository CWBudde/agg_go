package main

import (
	"fmt"
	"agg_go/internal/demo/shapesdata"
)

func main() {
	shapes := shapesdata.LoadShapes()
	s := &shapes[0]
	fmt.Printf("Shape 0: %d paths, styles %d..%d\n\n", len(s.Paths), s.MinStyle, s.MaxStyle)

	for style := s.MinStyle; style <= s.MaxStyle; style++ {
		fwd, inv, skip := 0, 0, 0
		for _, p := range s.Paths {
			if p.LeftFill == p.RightFill { skip++; continue }
			if p.LeftFill == style { fwd++ }
			if p.RightFill == style { inv++ }
		}
		fmt.Printf("Style %d: %d forward, %d inverted, %d skipped\n", style, fwd, inv, skip)
	}

	// Check: does any style have inverted paths first?
	fmt.Println()
	for style := s.MinStyle; style <= s.MaxStyle; style++ {
		firstIsInverted := false
		for _, p := range s.Paths {
			if p.LeftFill == p.RightFill { continue }
			if p.LeftFill == style {
				break // forward comes first
			}
			if p.RightFill == style {
				firstIsInverted = true
				break
			}
		}
		if firstIsInverted {
			fmt.Printf("  WARNING: Style %d has inverted path first!\n", style)
		}
	}
	
	// Show paths where lf==rf (skipped) - what are their values?
	fmt.Println("\nPaths with lf==rf:")
	for i, p := range s.Paths {
		if p.LeftFill == p.RightFill {
			fmt.Printf("  Path %d: lf=rf=%d ln=%d\n", i, p.LeftFill, p.Line)
		}
	}
}
