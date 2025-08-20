// Package vertex_source provides vertex source implementations for AGG.
package vertex_source

// Square represents a square shape that can be drawn with a rasterizer.
// This is used in the aa_demo example for pixel magnification.
type Square struct {
	size float64 // Size of the square
}

// NewSquare creates a new square with the specified size.
func NewSquare(size float64) *Square {
	return &Square{
		size: size,
	}
}

// DrawSquare is a generic function to render a square using the provided rasterizer.
// This function mimics the C++ template method in the original AGG example.
func DrawSquare[RR RasterizerInterface](s *Square, ras RR, x, y float64) {
	ras.Reset()
	ras.MoveToD(x*s.size, y*s.size)
	ras.LineToD(x*s.size+s.size, y*s.size)
	ras.LineToD(x*s.size+s.size, y*s.size+s.size)
	ras.LineToD(x*s.size, y*s.size+s.size)
}

// RasterizerInterface defines the methods needed by DrawSquare
type RasterizerInterface interface {
	Reset()
	MoveToD(x, y float64)
	LineToD(x, y float64)
}
