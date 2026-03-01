package rasterizer

import (
	"agg_go/internal/basics"
)

// AA scale constants are defined in compound_aa.go

// Status enumeration for rasterizer state
type Status uint32

const (
	StatusInitial Status = iota
	StatusMoveTo
	StatusLineTo
	StatusClosed
)

// RasterizerScanlineAA is the main polygon rasterizer for high-quality anti-aliased rendering.
// It uses coordinates in the format specified by the converter's coordinate type.
// This is equivalent to AGG's rasterizer_scanline_aa<Clip> template class.
type RasterizerScanlineAA[C basics.CoordType, V Conv[C], Clip any] struct {
	outline *RasterizerCellsAASimple // Cell-based rasterizer
	clipper interface {
		ResetClipping()
		ClipBox(x1, y1, x2, y2 C)
		MoveTo(x1, y1 C)
		LineTo(sink LineSink, x2, y2 C)
	} // Clipping implementation
	conv        V                  // Conversion policy
	gamma       [AAScale]uint8     // Gamma correction table
	fillingRule basics.FillingRule // Filling rule (non-zero or even-odd)
	autoClose   bool               // Auto-close polygons flag
	startX      C                  // Starting X coordinate (in converter coord_type)
	startY      C                  // Starting Y coordinate (in converter coord_type)
	status      Status             // Current rasterizer status
	scanY       int                // Current scanline Y coordinate
}

// NewRasterizerScanlineAA creates a new anti-aliased scanline rasterizer
func NewRasterizerScanlineAA[C basics.CoordType, V Conv[C], Clip any](conv V, clipper interface {
	ResetClipping()
	ClipBox(x1, y1, x2, y2 C)
	MoveTo(x1, y1 C)
	LineTo(sink LineSink, x2, y2 C)
},
) *RasterizerScanlineAA[C, V, Clip] {
	r := &RasterizerScanlineAA[C, V, Clip]{
		outline:     NewRasterizerCellsAASimple(256), // Default cell block limit
		clipper:     clipper,
		conv:        conv,
		fillingRule: basics.FillNonZero,
		autoClose:   true,
		status:      StatusInitial,
		scanY:       0,
	}

	// Initialize linear gamma table
	for i := 0; i < AAScale; i++ {
		r.gamma[i] = uint8(i)
	}

	return r
}

// NewRasterizerScanlineAAWithGamma creates a new rasterizer with custom gamma function
func NewRasterizerScanlineAAWithGamma[C basics.CoordType, V Conv[C], Clip any](conv V, clipper interface {
	ResetClipping()
	ClipBox(x1, y1, x2, y2 C)
	MoveTo(x1, y1 C)
	LineTo(sink LineSink, x2, y2 C)
}, gammaFunc func(float64) float64,
) *RasterizerScanlineAA[C, V, Clip] {
	r := NewRasterizerScanlineAA[C, V, Clip](conv, clipper)
	r.SetGamma(gammaFunc)
	return r
}

// Reset clears the rasterizer and prepares it for new geometry
func (r *RasterizerScanlineAA[C, V, Clip]) Reset() {
	r.outline.Reset()
	r.status = StatusInitial
}

// ResetClipping resets the clipping settings
func (r *RasterizerScanlineAA[C, V, Clip]) ResetClipping() {
	r.Reset()
	r.clipper.ResetClipping()
}

// ClipBox sets the clipping rectangle
func (r *RasterizerScanlineAA[C, V, Clip]) ClipBox(x1, y1, x2, y2 float64) {
	r.Reset()
	r.clipper.ClipBox(r.conv.Upscale(x1), r.conv.Upscale(y1), r.conv.Upscale(x2), r.conv.Upscale(y2))
}

// FillingRule sets the polygon filling rule
func (r *RasterizerScanlineAA[C, V, Clip]) FillingRule(rule basics.FillingRule) {
	r.fillingRule = rule
}

// AutoClose sets whether polygons should be automatically closed
func (r *RasterizerScanlineAA[C, V, Clip]) AutoClose(flag bool) {
	r.autoClose = flag
}

// SetGamma sets the gamma correction function
func (r *RasterizerScanlineAA[C, V, Clip]) SetGamma(gammaFunc func(float64) float64) {
	for i := 0; i < AAScale; i++ {
		val := gammaFunc(float64(i)/float64(AAMask)) * float64(AAMask)
		if val < 0 {
			val = 0
		}
		if val > AAMask {
			val = AAMask
		}
		r.gamma[i] = uint8(val)
	}
}

// ApplyGamma applies gamma correction to a coverage value
func (r *RasterizerScanlineAA[C, V, Clip]) ApplyGamma(cover int) uint8 {
	if cover > AAMask {
		cover = AAMask
	}
	return r.gamma[cover]
}

// MoveTo starts a new contour at the specified integer coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) MoveTo(x, y int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}

	r.startX = r.conv.Downscale(x * basics.PolySubpixelScale)
	r.startY = r.conv.Downscale(y * basics.PolySubpixelScale)
	r.clipper.MoveTo(r.startX, r.startY)
	r.status = StatusMoveTo
}

// LineTo draws a line to the specified integer coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) LineTo(x, y int) {
	xCoord := r.conv.Downscale(x * basics.PolySubpixelScale)
	yCoord := r.conv.Downscale(y * basics.PolySubpixelScale)
	r.clipper.LineTo(r.outline, xCoord, yCoord)
	r.status = StatusLineTo
}

// MoveToD starts a new contour at the specified double coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) MoveToD(x, y float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}

	r.startX = r.conv.Upscale(x)
	r.startY = r.conv.Upscale(y)
	r.clipper.MoveTo(r.startX, r.startY)
	r.status = StatusMoveTo
}

// LineToD draws a line to the specified double coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) LineToD(x, y float64) {
	xCoord := r.conv.Upscale(x)
	yCoord := r.conv.Upscale(y)
	r.clipper.LineTo(r.outline, xCoord, yCoord)
	r.status = StatusLineTo
}

// ClosePolygon closes the current polygon
func (r *RasterizerScanlineAA[C, V, Clip]) ClosePolygon() {
	if r.status == StatusLineTo || r.status == StatusMoveTo {
		r.clipper.LineTo(r.outline, r.startX, r.startY)
		r.status = StatusClosed
	}
}

// AddVertex adds a vertex with the specified command
func (r *RasterizerScanlineAA[C, V, Clip]) AddVertex(x, y float64, cmd uint32) {
	pathCmd := basics.PathCommand(cmd & uint32(basics.PathCmdMask))

	if basics.IsMoveTo(pathCmd) {
		r.MoveToD(x, y)
	} else if basics.IsVertex(pathCmd) {
		r.LineToD(x, y)
	} else if basics.IsClose(cmd) {
		r.ClosePolygon()
	}
}

// Edge adds a single edge from (x1,y1) to (x2,y2) with integer coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) Edge(x1, y1, x2, y2 int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	x1Coord := r.conv.Downscale(x1 * basics.PolySubpixelScale)
	y1Coord := r.conv.Downscale(y1 * basics.PolySubpixelScale)
	x2Coord := r.conv.Downscale(x2 * basics.PolySubpixelScale)
	y2Coord := r.conv.Downscale(y2 * basics.PolySubpixelScale)
	r.clipper.MoveTo(x1Coord, y1Coord)
	r.clipper.LineTo(r.outline, x2Coord, y2Coord)
}

// EdgeD adds a single edge from (x1,y1) to (x2,y2) with double coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) EdgeD(x1, y1, x2, y2 float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	x1Coord := r.conv.Upscale(x1)
	y1Coord := r.conv.Upscale(y1)
	x2Coord := r.conv.Upscale(x2)
	y2Coord := r.conv.Upscale(y2)
	r.clipper.MoveTo(x1Coord, y1Coord)
	r.clipper.LineTo(r.outline, x2Coord, y2Coord)
}

// AddPath adds all vertices from a vertex source
func (r *RasterizerScanlineAA[C, V, Clip]) AddPath(vs VertexSource, pathID uint32) {
	var x, y float64

	vs.Rewind(pathID)
	if r.outline.Sorted() {
		r.Reset()
	}

	for {
		cmd := vs.Vertex(&x, &y)
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		r.AddVertex(x, y, cmd)
	}
}

// MinX returns the minimum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[C, V, Clip]) MinX() int {
	return r.outline.MinX()
}

// MinY returns the minimum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[C, V, Clip]) MinY() int {
	return r.outline.MinY()
}

// MaxX returns the maximum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[C, V, Clip]) MaxX() int {
	return r.outline.MaxX()
}

// MaxY returns the maximum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[C, V, Clip]) MaxY() int {
	return r.outline.MaxY()
}

// Sort sorts the cells in preparation for scanline rendering
func (r *RasterizerScanlineAA[C, V, Clip]) Sort() {
	if r.autoClose {
		r.ClosePolygon()
	}
	r.outline.SortCells()
}

// RewindScanlines resets the scanline iterator
func (r *RasterizerScanlineAA[C, V, Clip]) RewindScanlines() bool {
	if r.autoClose {
		r.ClosePolygon()
	}
	r.outline.SortCells()
	if r.outline.TotalCells() == 0 {
		return false
	}
	r.scanY = r.outline.MinY()
	return true
}

// NavigateScanline moves to the specified scanline Y coordinate
func (r *RasterizerScanlineAA[C, V, Clip]) NavigateScanline(y int) bool {
	if r.autoClose {
		r.ClosePolygon()
	}
	r.outline.SortCells()
	if r.outline.TotalCells() == 0 ||
		y < r.outline.MinY() ||
		y > r.outline.MaxY() {
		return false
	}
	r.scanY = y
	return true
}

// CalculateAlpha calculates the alpha value for a given area
func (r *RasterizerScanlineAA[C, V, Clip]) CalculateAlpha(area int) uint8 {
	cover := area >> (basics.PolySubpixelShift*2 + 1 - AAShift)

	if cover < 0 {
		cover = -cover
	}

	if r.fillingRule == basics.FillEvenOdd {
		cover &= AAMask2
		if cover > AAScale {
			cover = AAScale2 - cover
		}
	}

	if cover > AAMask {
		cover = AAMask
	}

	return r.gamma[cover]
}

// SweepScanline generates the next scanline and stores it in the provided scanline object
func (r *RasterizerScanlineAA[C, V, Clip]) SweepScanline(sl ScanlineInterface) bool {
	for {
		if r.scanY > r.outline.MaxY() {
			return false
		}

		sl.ResetSpans()
		numCells := r.outline.ScanlineNumCells(r.scanY)
		cells := r.outline.ScanlineCells(r.scanY)
		cover := 0

		cellIndex := uint32(0)
		for cellIndex < numCells {
			curCell := cells[cellIndex]
			x := curCell.X
			area := curCell.Area

			cover += curCell.Cover

			// Accumulate all cells with the same X coordinate
			cellIndex++
			for cellIndex < numCells {
				nextCell := cells[cellIndex]
				if nextCell.X != x {
					break
				}
				area += nextCell.Area
				cover += nextCell.Cover
				cellIndex++
			}

			if area != 0 {
				alpha := r.CalculateAlpha((cover << (basics.PolySubpixelShift + 1)) - area)
				if alpha != 0 {
					sl.AddCell(x, uint32(alpha))
				}
				x++
			}

			if cellIndex < numCells && cells[cellIndex].X > x {
				alpha := r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
				if alpha != 0 {
					sl.AddSpan(x, cells[cellIndex].X-x, uint32(alpha))
				}
			}
		}

		if sl.NumSpans() > 0 {
			break
		}
		r.scanY++
	}

	sl.Finalize(r.scanY)
	r.scanY++
	return true
}

// HitTest performs a hit test at the specified coordinates
func (r *RasterizerScanlineAA[C, V, Clip]) HitTest(tx, ty int) bool {
	if !r.NavigateScanline(ty) {
		return false
	}

	numCells := r.outline.ScanlineNumCells(ty)
	cells := r.outline.ScanlineCells(ty)
	cover := 0

	for i := uint32(0); i < numCells; i++ {
		curCell := cells[i]
		x := curCell.X

		if x > tx {
			break
		}

		cover += curCell.Cover

		if x == tx {
			area := curCell.Area
			for i++; i < numCells && cells[i].X == x; i++ {
				area += cells[i].Area
				cover += cells[i].Cover
			}

			alpha := r.CalculateAlpha((cover << (basics.PolySubpixelShift + 1)) - area)
			return alpha != 0
		}
	}

	alpha := r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
	return alpha != 0
}
