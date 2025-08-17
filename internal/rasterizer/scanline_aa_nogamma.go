package rasterizer

import (
	"agg_go/internal/basics"
)

// RasterizerScanlineAANoGamma is a polygon rasterizer optimized for high-quality
// anti-aliased rendering without gamma correction. It provides the same interface
// as RasterizerScanlineAA but with simplified coverage calculation for better performance.
// This is equivalent to AGG's rasterizer_scanline_aa_nogamma<Clip> template class.
type RasterizerScanlineAANoGamma[Clip ClipInterface] struct {
	outline     *RasterizerCellsAASimple // Cell-based rasterizer
	clipper     Clip                     // Clipping implementation
	fillingRule basics.FillingRule       // Filling rule (non-zero or even-odd)
	autoClose   bool                     // Auto-close polygons flag
	startX      int                      // Starting X coordinate
	startY      int                      // Starting Y coordinate
	status      Status                   // Current rasterizer status
	scanY       int                      // Current scanline Y coordinate
}

// NewRasterizerScanlineAANoGamma creates a new anti-aliased scanline rasterizer without gamma correction
func NewRasterizerScanlineAANoGamma[Clip ClipInterface](cellBlockLimit uint32) *RasterizerScanlineAANoGamma[Clip] {
	return &RasterizerScanlineAANoGamma[Clip]{
		outline:     NewRasterizerCellsAASimple(cellBlockLimit),
		fillingRule: basics.FillNonZero,
		autoClose:   true,
		startX:      0,
		startY:      0,
		status:      StatusInitial,
		scanY:       0,
	}
}

// Reset clears the rasterizer and prepares it for new geometry
func (r *RasterizerScanlineAANoGamma[Clip]) Reset() {
	r.outline.Reset()
	r.status = StatusInitial
}

// ResetClipping resets the clipping settings
func (r *RasterizerScanlineAANoGamma[Clip]) ResetClipping() {
	r.Reset()
	r.clipper.ResetClipping()
}

// ClipBox sets the clipping rectangle
func (r *RasterizerScanlineAANoGamma[Clip]) ClipBox(x1, y1, x2, y2 float64) {
	r.Reset()
	r.clipper.ClipBox(x1, y1, x2, y2)
}

// FillingRule sets the polygon filling rule
func (r *RasterizerScanlineAANoGamma[Clip]) FillingRule(rule basics.FillingRule) {
	r.fillingRule = rule
}

// AutoClose sets whether polygons should be automatically closed
func (r *RasterizerScanlineAANoGamma[Clip]) AutoClose(flag bool) {
	r.autoClose = flag
}

// ApplyGamma applies gamma correction to the coverage value.
// In the no-gamma variant, this simply returns the cover value unchanged.
func (r *RasterizerScanlineAANoGamma[Clip]) ApplyGamma(cover uint32) uint32 {
	return cover
}

// MoveTo starts a new contour at the specified integer coordinates
func (r *RasterizerScanlineAANoGamma[Clip]) MoveTo(x, y int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}
	r.startX = x
	r.startY = y
	r.clipper.MoveTo(float64(x)/basics.PolySubpixelScale, float64(y)/basics.PolySubpixelScale)
	r.status = StatusMoveTo
}

// LineTo adds a line segment to the current contour
func (r *RasterizerScanlineAANoGamma[Clip]) LineTo(x, y int) {
	r.clipper.LineTo(float64(x)/basics.PolySubpixelScale, float64(y)/basics.PolySubpixelScale)
	r.status = StatusLineTo
}

// MoveToD starts a new contour at the specified floating-point coordinates
func (r *RasterizerScanlineAANoGamma[Clip]) MoveToD(x, y float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}
	r.startX = int(x * basics.PolySubpixelScale)
	r.startY = int(y * basics.PolySubpixelScale)
	r.clipper.MoveTo(x, y)
	r.status = StatusMoveTo
}

// LineToD adds a line segment to the current contour using floating-point coordinates
func (r *RasterizerScanlineAANoGamma[Clip]) LineToD(x, y float64) {
	r.clipper.LineTo(x, y)
	r.status = StatusLineTo
}

// ClosePolygon closes the current polygon contour
func (r *RasterizerScanlineAANoGamma[Clip]) ClosePolygon() {
	if r.status == StatusLineTo {
		r.clipper.LineTo(float64(r.startX)/basics.PolySubpixelScale, float64(r.startY)/basics.PolySubpixelScale)
		r.status = StatusClosed
	}
}

// AddVertex adds a vertex to the path based on the path command
func (r *RasterizerScanlineAANoGamma[Clip]) AddVertex(x, y float64, cmd uint32) {
	pathCmd := basics.PathCommand(cmd & uint32(basics.PathCmdMask))

	switch {
	case basics.IsMoveTo(pathCmd):
		r.MoveToD(x, y)
	case basics.IsVertex(pathCmd):
		r.LineToD(x, y)
	case basics.IsClose(cmd):
		r.ClosePolygon()
	}
}

// Edge adds a single edge (line segment) to the rasterizer
func (r *RasterizerScanlineAANoGamma[Clip]) Edge(x1, y1, x2, y2 int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.clipper.MoveTo(float64(x1)/basics.PolySubpixelScale, float64(y1)/basics.PolySubpixelScale)
	r.clipper.LineTo(float64(x2)/basics.PolySubpixelScale, float64(y2)/basics.PolySubpixelScale)
	r.status = StatusMoveTo
}

// EdgeD adds a single edge using floating-point coordinates
func (r *RasterizerScanlineAANoGamma[Clip]) EdgeD(x1, y1, x2, y2 float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.clipper.MoveTo(x1, y1)
	r.clipper.LineTo(x2, y2)
	r.status = StatusMoveTo
}

// AddPath adds a complete path from a vertex source
func (r *RasterizerScanlineAANoGamma[Clip]) AddPath(vs VertexSource, pathID uint32) {
	var x, y float64
	var cmd uint32

	vs.Rewind(pathID)
	if r.outline.Sorted() {
		r.Reset()
	}

	for {
		cmd = vs.Vertex(&x, &y)
		if basics.IsStop(basics.PathCommand(cmd & uint32(basics.PathCmdMask))) {
			break
		}
		r.AddVertex(x, y, cmd)
	}
}

// MinX returns the minimum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAANoGamma[Clip]) MinX() int {
	return r.outline.MinX()
}

// MinY returns the minimum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAANoGamma[Clip]) MinY() int {
	return r.outline.MinY()
}

// MaxX returns the maximum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAANoGamma[Clip]) MaxX() int {
	return r.outline.MaxX()
}

// MaxY returns the maximum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAANoGamma[Clip]) MaxY() int {
	return r.outline.MaxY()
}

// Sort sorts the cells for scanline processing
func (r *RasterizerScanlineAANoGamma[Clip]) Sort() {
	if r.autoClose {
		r.ClosePolygon()
	}
	r.outline.SortCells()
}

// RewindScanlines prepares for scanline iteration, returns true if there are scanlines to process
func (r *RasterizerScanlineAANoGamma[Clip]) RewindScanlines() bool {
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

// NavigateScanline positions the rasterizer at a specific scanline
func (r *RasterizerScanlineAANoGamma[Clip]) NavigateScanline(y int) bool {
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

// CalculateAlpha calculates the alpha (coverage) value for the given area
// without gamma correction, following the original AGG algorithm
func (r *RasterizerScanlineAANoGamma[Clip]) CalculateAlpha(area int) uint32 {
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

	return uint32(cover)
}

// SweepScanline generates the next scanline of anti-aliased spans
func (r *RasterizerScanlineAANoGamma[Clip]) SweepScanline(sl ScanlineInterface) bool {
	for {
		if r.scanY > r.outline.MaxY() {
			return false
		}

		sl.ResetSpans()
		numCells := r.outline.ScanlineNumCells(uint32(r.scanY))
		cells := r.outline.ScanlineCells(uint32(r.scanY))
		cover := 0

		cellIndex := uint32(0)
		for cellIndex < numCells {
			curCell := cells[cellIndex]
			x := curCell.X
			area := curCell.Area
			var alpha uint32

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
				alpha = r.CalculateAlpha((cover << (basics.PolySubpixelShift + 1)) - area)
				if alpha != 0 {
					sl.AddCell(x, alpha)
				}
				x++
			}

			if cellIndex < numCells && cells[cellIndex].X > x {
				alpha = r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
				if alpha != 0 {
					sl.AddSpan(x, cells[cellIndex].X-x, alpha)
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

// HitTest checks if the specified point is inside the rasterized geometry
func (r *RasterizerScanlineAANoGamma[Clip]) HitTest(tx, ty int) bool {
	if !r.NavigateScanline(ty) {
		return false
	}

	// Use a simple hit test scanline that tracks whether we're inside
	hitTest := &HitTestScanline{targetX: tx, hit: false}
	r.SweepScanline(hitTest)
	return hitTest.hit
}

// HitTestScanline is a specialized scanline for hit testing
type HitTestScanline struct {
	targetX int
	hit     bool
}

func (ht *HitTestScanline) ResetSpans() {
	ht.hit = false
}

func (ht *HitTestScanline) AddCell(x int, cover uint32) {
	if x == ht.targetX && cover > 0 {
		ht.hit = true
	}
}

func (ht *HitTestScanline) AddSpan(x, length int, cover uint32) {
	if cover > 0 && ht.targetX >= x && ht.targetX < x+length {
		ht.hit = true
	}
}

func (ht *HitTestScanline) Finalize(y int) {
	// Nothing to do
}

func (ht *HitTestScanline) NumSpans() int {
	if ht.hit {
		return 1
	}
	return 0
}
