// Package span provides gradient contour generation functionality for AGG.
// This implements a port of AGG's span_gradient_contour classes and functions.
package span

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/conv"
	"agg_go/internal/path"
	"agg_go/internal/transform"
)

// PathVertexAdapter adapts PathBase to the VertexSource interface used by conv
type PathVertexAdapter struct {
	path *path.PathStorage
}

// NewPathVertexAdapter creates a new adapter
func NewPathVertexAdapter(p *path.PathStorage) *PathVertexAdapter {
	return &PathVertexAdapter{path: p}
}

// Rewind rewinds the path to the specified path ID
func (pva *PathVertexAdapter) Rewind(pathID uint) {
	pva.path.Rewind(pathID)
}

// Vertex returns the next vertex using the NextVertex method
func (pva *PathVertexAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmdUint := pva.path.NextVertex()
	return x, y, basics.PathCommand(cmdUint)
}

// GradientContour generates contour-based gradients using distance transforms.
// This is a port of AGG's gradient_contour class.
type GradientContour struct {
	buffer []uint8
	width  int
	height int
	frame  int
	d1     float64
	d2     float64
}

// NewGradientContour creates a new gradient contour generator.
func NewGradientContour() *GradientContour {
	return &GradientContour{
		frame: 10,
		d1:    0.0,
		d2:    100.0,
	}
}

// NewGradientContourWithDistances creates a gradient contour generator with specific distances.
func NewGradientContourWithDistances(d1, d2 float64) *GradientContour {
	return &GradientContour{
		frame: 10,
		d1:    d1,
		d2:    d2,
	}
}

// boundingRectSingle calculates the bounding rectangle of a vertex source for a single path
func boundingRectSingle(vs conv.VertexSource, pathID uint) (x1, y1, x2, y2 float64, ok bool) {
	vs.Rewind(pathID)
	first := true

	x1, y1, x2, y2 = 1, 1, 0, 0 // Invalid initial state

	for {
		x, y, cmd := vs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		if basics.IsVertex(cmd) {
			if first {
				x1, y1, x2, y2 = x, y, x, y
				first = false
			} else {
				if x < x1 {
					x1 = x
				}
				if y < y1 {
					y1 = y
				}
				if x > x2 {
					x2 = x
				}
				if y > y2 {
					y2 = y
				}
			}
		}
	}

	ok = x1 <= x2 && y1 <= y2
	return
}

// ContourCreate generates a contour gradient buffer from a path.
// Returns the buffer data or nil if creation failed.
func (gc *GradientContour) ContourCreate(ps *path.PathStorage) []uint8 {
	if ps == nil {
		return nil
	}

	// Create adapter and convert path to curves
	adapter := NewPathVertexAdapter(ps)
	convCurve := conv.NewConvCurve(adapter)

	// Get bounding rectangle
	x1, y1, x2, y2, ok := boundingRectSingle(convCurve, 0)
	if !ok {
		return nil
	}

	// Create rendering surface with frame
	width := int(math.Ceil(x2-x1)) + gc.frame*2 + 1
	height := int(math.Ceil(y2-y1)) + gc.frame*2 + 1

	// For now, create a simplified black and white buffer by rasterizing the path manually
	// This is a simplified implementation that will be enhanced later
	bwBuffer := make([]uint8, width*height)

	// Initialize to white (255)
	for i := range bwBuffer {
		bwBuffer[i] = 255
	}

	// Setup transformation matrix
	mtx := transform.NewTransAffine()
	mtx = mtx.Multiply(transform.NewTransAffineTranslation(-x1+float64(gc.frame), -y1+float64(gc.frame)))

	// Transform and rasterize the curve manually (simplified approach)
	trans := conv.NewConvTransform(convCurve, mtx)
	gc.rasterizePathSimple(trans, bwBuffer, width, height)

	// Now perform distance transform
	return gc.performDistanceTransform(bwBuffer, width, height)
}

// rasterizePathSimple performs simple path rasterization by drawing lines between vertices
func (gc *GradientContour) rasterizePathSimple(vs conv.VertexSource, buffer []uint8, width, height int) {
	vs.Rewind(0)

	var startX, startY, lastX, lastY float64
	first := true

	for {
		x, y, cmd := vs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		switch cmd {
		case basics.PathCmdMoveTo:
			startX, startY = x, y
			lastX, lastY = x, y
			first = false

		case basics.PathCmdLineTo:
			if !first {
				gc.drawLine(lastX, lastY, x, y, buffer, width, height)
			}
			lastX, lastY = x, y

		case basics.PathCmdEndPoly:
			if !first {
				gc.drawLine(lastX, lastY, startX, startY, buffer, width, height)
			}
		}
	}
}

// drawLine draws a simple line using Bresenham's algorithm
func (gc *GradientContour) drawLine(x0, y0, x1, y1 float64, buffer []uint8, width, height int) {
	ix0 := int(x0 + 0.5)
	iy0 := int(y0 + 0.5)
	ix1 := int(x1 + 0.5)
	iy1 := int(y1 + 0.5)

	dx := basics.Abs(ix1 - ix0)
	dy := basics.Abs(iy1 - iy0)

	sx := 1
	if ix0 > ix1 {
		sx = -1
	}

	sy := 1
	if iy0 > iy1 {
		sy = -1
	}

	err := dx - dy

	x, y := ix0, iy0

	for {
		// Set pixel to black if within bounds
		if x >= 0 && x < width && y >= 0 && y < height {
			buffer[y*width+x] = 0
		}

		if x == ix1 && y == iy1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// performDistanceTransform applies the distance transform algorithm
func (gc *GradientContour) performDistanceTransform(bwBuffer []uint8, width, height int) []uint8 {
	// Create float buffer for distance transform
	image := make([]float32, width*height)

	// Initialize: 0 for black pixels, infinity for white pixels
	for i, pixel := range bwBuffer {
		if pixel == 0 {
			image[i] = 0.0
		} else {
			image[i] = math.MaxFloat32
		}
	}

	// Determine maximum dimension for working arrays
	length := width
	if height > length {
		length = height
	}

	// Allocate working arrays for DT algorithm
	spanf := make([]float32, length)
	spang := make([]float32, length+1)
	spanr := make([]float32, length)
	spann := make([]int, length)

	// Transform along columns
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			spanf[y] = image[y*width+x]
		}

		// Apply 1D distance transform
		gc.dt(spanf, spang, spanr, spann, height)

		for y := 0; y < height; y++ {
			image[y*width+x] = spanr[y]
		}
	}

	// Transform along rows
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			spanf[x] = image[y*width+x]
		}

		// Apply 1D distance transform
		gc.dt(spanf, spang, spanr, spann, width)

		for x := 0; x < width; x++ {
			image[y*width+x] = spanr[x]
		}
	}

	// Take square roots and find min/max
	var min, max float32
	for i := range image {
		image[i] = float32(math.Sqrt(float64(image[i])))
		if i == 0 {
			min = image[i]
			max = image[i]
		} else {
			if image[i] < min {
				min = image[i]
			}
			if image[i] > max {
				max = image[i]
			}
		}
	}

	// Convert to grayscale
	result := make([]uint8, width*height)
	if min == max {
		// All values are the same, set to black
		for i := range result {
			result[i] = 0
		}
	} else {
		scale := 255.0 / (max - min)
		for i := range image {
			result[i] = uint8(int((image[i]-min)*scale + 0.5))
		}
	}

	// Store buffer dimensions
	gc.buffer = result
	gc.width = width
	gc.height = height

	return result
}

// dt performs the 1D distance transform algorithm by Pedro Felzenszwalb
func (gc *GradientContour) dt(spanf, spang, spanr []float32, spann []int, length int) {
	square := func(x int) int { return x * x }

	k := 0
	spann[0] = 0
	spang[0] = -math.MaxFloat32
	spang[1] = math.MaxFloat32

	for q := 1; q <= length-1; q++ {
		s := ((spanf[q] + float32(square(q))) - (spanf[spann[k]] + float32(square(spann[k])))) / float32(2*q-2*spann[k])

		for s <= spang[k] {
			k--
			s = ((spanf[q] + float32(square(q))) - (spanf[spann[k]] + float32(square(spann[k])))) / float32(2*q-2*spann[k])
		}

		k++
		spann[k] = q
		spang[k] = s
		spang[k+1] = math.MaxFloat32
	}

	k = 0
	for q := 0; q <= length-1; q++ {
		for spang[k+1] < float32(q) {
			k++
		}
		spanr[q] = float32(square(q-spann[k])) + spanf[spann[k]]
	}
}

// ContourWidth returns the width of the generated contour buffer.
func (gc *GradientContour) ContourWidth() int {
	return gc.width
}

// ContourHeight returns the height of the generated contour buffer.
func (gc *GradientContour) ContourHeight() int {
	return gc.height
}

// SetD1 sets the start distance parameter.
func (gc *GradientContour) SetD1(d float64) {
	gc.d1 = d
}

// SetD2 sets the end distance parameter.
func (gc *GradientContour) SetD2(d float64) {
	gc.d2 = d
}

// SetFrame sets the frame size around the path.
func (gc *GradientContour) SetFrame(f int) {
	gc.frame = f
}

// Frame returns the current frame size.
func (gc *GradientContour) Frame() int {
	return gc.frame
}

// Calculate computes the gradient value at the given coordinates.
// This implements the GradientFunction interface.
func (gc *GradientContour) Calculate(x, y, d int) int {
	if gc.buffer == nil {
		return 0
	}

	// Convert from subpixel to pixel coordinates
	px := x >> GradientSubpixelShift
	py := y >> GradientSubpixelShift

	// Wrap coordinates to buffer dimensions
	px %= gc.width
	if px < 0 {
		px += gc.width
	}

	py %= gc.height
	if py < 0 {
		py += gc.height
	}

	// Sample buffer and scale to gradient range
	sample := float64(gc.buffer[py*gc.width+px])
	result := gc.d1 + (sample/255.0)*(gc.d2-gc.d1)

	return basics.IRound(result) << GradientSubpixelShift
}
