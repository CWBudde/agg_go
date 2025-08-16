package rasterizer

import (
	"agg_go/internal/basics"
)

// Status enumeration for rasterizer state
type Status uint32

const (
	StatusInitial Status = iota
	StatusMoveTo
	StatusLineTo
	StatusClosed
)

// AA scale constants for anti-aliasing
const (
	AAShift  = 8
	AAScale  = 1 << AAShift
	AAMask   = AAScale - 1
	AAScale2 = AAScale * 2
	AAMask2  = AAScale2 - 1
)

// RasterizerScanlineAA is the main polygon rasterizer for high-quality anti-aliased rendering.
// It uses integer coordinates in format 24.8 (24 bits integer, 8 bits fractional).
// This is equivalent to AGG's rasterizer_scanline_aa<Clip> template class.
type RasterizerScanlineAA[Clip ClipInterface] struct {
	outline     *RasterizerCellsAASimple // Cell-based rasterizer
	clipper     Clip                     // Clipping implementation
	gamma       [AAScale]int             // Gamma correction table
	fillingRule basics.FillingRule       // Filling rule (non-zero or even-odd)
	autoClose   bool                     // Auto-close polygons flag
	startX      int                      // Starting X coordinate
	startY      int                      // Starting Y coordinate
	status      Status                   // Current rasterizer status
	scanY       int                      // Current scanline Y coordinate
}

// NewRasterizerScanlineAA creates a new anti-aliased scanline rasterizer
func NewRasterizerScanlineAA[Clip ClipInterface](cellBlockLimit uint32) *RasterizerScanlineAA[Clip] {
	r := &RasterizerScanlineAA[Clip]{
		outline:     NewRasterizerCellsAASimple(cellBlockLimit),
		fillingRule: basics.FillNonZero,
		autoClose:   true,
		startX:      0,
		startY:      0,
		status:      StatusInitial,
		scanY:       0,
	}

	// Initialize linear gamma table
	for i := 0; i < AAScale; i++ {
		r.gamma[i] = i
	}

	return r
}

// NewRasterizerScanlineAAWithGamma creates a new rasterizer with custom gamma function
func NewRasterizerScanlineAAWithGamma[Clip ClipInterface](gammaFunc func(float64) float64, cellBlockLimit uint32) *RasterizerScanlineAA[Clip] {
	r := NewRasterizerScanlineAA[Clip](cellBlockLimit)
	r.SetGamma(gammaFunc)
	return r
}

// Reset clears the rasterizer and prepares it for new geometry
func (r *RasterizerScanlineAA[Clip]) Reset() {
	r.outline.Reset()
	r.status = StatusInitial
}

// ResetClipping resets the clipping settings
func (r *RasterizerScanlineAA[Clip]) ResetClipping() {
	r.Reset()
	r.clipper.ResetClipping()
}

// ClipBox sets the clipping rectangle
func (r *RasterizerScanlineAA[Clip]) ClipBox(x1, y1, x2, y2 float64) {
	r.Reset()
	r.clipper.ClipBox(x1, y1, x2, y2)
}

// FillingRule sets the polygon filling rule
func (r *RasterizerScanlineAA[Clip]) FillingRule(rule basics.FillingRule) {
	r.fillingRule = rule
}

// AutoClose sets whether polygons should be automatically closed
func (r *RasterizerScanlineAA[Clip]) AutoClose(flag bool) {
	r.autoClose = flag
}

// SetGamma sets the gamma correction function
func (r *RasterizerScanlineAA[Clip]) SetGamma(gammaFunc func(float64) float64) {
	for i := 0; i < AAScale; i++ {
		r.gamma[i] = int(basics.URound(gammaFunc(float64(i)/AAMask) * AAMask))
	}
}

// ApplyGamma applies gamma correction to a coverage value
func (r *RasterizerScanlineAA[Clip]) ApplyGamma(cover uint32) uint32 {
	return uint32(r.gamma[cover])
}

// MoveTo starts a new contour at the specified integer coordinates
func (r *RasterizerScanlineAA[Clip]) MoveTo(x, y int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}

	r.startX = x
	r.startY = y
	r.clipper.MoveTo(float64(x), float64(y))
	r.status = StatusMoveTo
}

// LineTo draws a line to the specified integer coordinates
func (r *RasterizerScanlineAA[Clip]) LineTo(x, y int) {
	r.clipper.LineTo(float64(x), float64(y))
	r.status = StatusLineTo
}

// MoveToD starts a new contour at the specified double coordinates
func (r *RasterizerScanlineAA[Clip]) MoveToD(x, y float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}

	r.startX = basics.IRound(x * basics.PolySubpixelScale)
	r.startY = basics.IRound(y * basics.PolySubpixelScale)
	r.clipper.MoveTo(x, y)
	r.status = StatusMoveTo
}

// LineToD draws a line to the specified double coordinates
func (r *RasterizerScanlineAA[Clip]) LineToD(x, y float64) {
	r.clipper.LineTo(x, y)
	r.status = StatusLineTo
}

// ClosePolygon closes the current polygon
func (r *RasterizerScanlineAA[Clip]) ClosePolygon() {
	if r.status == StatusLineTo {
		r.clipper.LineTo(float64(r.startX)/basics.PolySubpixelScale, float64(r.startY)/basics.PolySubpixelScale)
		r.status = StatusClosed
	}
}

// AddVertex adds a vertex with the specified command
func (r *RasterizerScanlineAA[Clip]) AddVertex(x, y float64, cmd uint32) {
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
func (r *RasterizerScanlineAA[Clip]) Edge(x1, y1, x2, y2 int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.clipper.MoveTo(float64(x1), float64(y1))
	r.clipper.LineTo(float64(x2), float64(y2))
}

// EdgeD adds a single edge from (x1,y1) to (x2,y2) with double coordinates
func (r *RasterizerScanlineAA[Clip]) EdgeD(x1, y1, x2, y2 float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.clipper.MoveTo(x1, y1)
	r.clipper.LineTo(x2, y2)
}

// AddPath adds all vertices from a vertex source
func (r *RasterizerScanlineAA[Clip]) AddPath(vs VertexSource, pathID uint32) {
	var x, y float64
	var cmd uint32

	vs.Rewind(pathID)
	if r.outline.Sorted() {
		r.Reset()
	}

	for !basics.IsStop(basics.PathCommand(cmd)) {
		cmd = vs.Vertex(&x, &y)
		r.AddVertex(x, y, cmd)
	}
}

// VertexSource interface for vertex generators
type VertexSource interface {
	Rewind(pathID uint32)
	Vertex(x, y *float64) uint32
}

// MinX returns the minimum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip]) MinX() int {
	return r.outline.MinX()
}

// MinY returns the minimum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip]) MinY() int {
	return r.outline.MinY()
}

// MaxX returns the maximum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip]) MaxX() int {
	return r.outline.MaxX()
}

// MaxY returns the maximum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip]) MaxY() int {
	return r.outline.MaxY()
}

// Sort sorts the cells in preparation for scanline rendering
func (r *RasterizerScanlineAA[Clip]) Sort() {
	r.outline.SortCells()
}

// RewindScanlines resets the scanline iterator
func (r *RasterizerScanlineAA[Clip]) RewindScanlines() bool {
	if r.outline.Sorted() {
		r.scanY = r.outline.MinY()
		return true
	}
	return false
}

// NavigateScanline moves to the specified scanline Y coordinate
func (r *RasterizerScanlineAA[Clip]) NavigateScanline(y int) bool {
	if r.outline.Sorted() && y >= r.outline.MinY() && y <= r.outline.MaxY() {
		r.scanY = y
		return true
	}
	return false
}

// CalculateAlpha calculates the alpha value for a given area
func (r *RasterizerScanlineAA[Clip]) CalculateAlpha(area int) uint32 {
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

	return uint32(r.gamma[cover])
}

// SweepScanline generates the next scanline and stores it in the provided scanline object
func (r *RasterizerScanlineAA[Clip]) SweepScanline(sl ScanlineInterface) bool {
	for {
		if r.scanY > r.outline.MaxY() {
			return false
		}

		sl.ResetSpans()
		numCells := r.outline.ScanlineNumCells(uint32(r.scanY))
		cells := r.outline.ScanlineCells(uint32(r.scanY))
		cover := 0

		for i := uint32(0); i < numCells; {
			curCell := cells[i]
			x := curCell.GetX()
			area := curCell.GetArea()

			cover += curCell.GetCover()

			// Accumulate all cells with the same X coordinate
			for i++; i < numCells; i++ {
				if cells[i].GetX() != x {
					break
				}
				area += cells[i].GetArea()
				cover += cells[i].GetCover()
			}

			if area != 0 {
				alpha := r.CalculateAlpha((cover << (basics.PolySubpixelShift + 1)) - area)
				if alpha != 0 {
					sl.AddCell(x, alpha)
				}
				x++
			}

			if i < numCells && cells[i].GetX() > x {
				alpha := r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
				if alpha != 0 {
					sl.AddSpan(x, cells[i].GetX()-x, alpha)
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

// ScanlineInterface defines the interface for scanline objects
type ScanlineInterface interface {
	ResetSpans()
	AddCell(x int, cover uint32)
	AddSpan(x, len int, cover uint32)
	Finalize(y int)
	NumSpans() int
}

// HitTest performs a hit test at the specified coordinates
func (r *RasterizerScanlineAA[Clip]) HitTest(tx, ty int) bool {
	if !r.outline.Sorted() {
		return false
	}

	if ty < r.outline.MinY() || ty > r.outline.MaxY() {
		return false
	}

	numCells := r.outline.ScanlineNumCells(uint32(ty))
	cells := r.outline.ScanlineCells(uint32(ty))
	cover := 0

	for i := uint32(0); i < numCells; i++ {
		curCell := cells[i]
		x := curCell.GetX()

		if x > tx {
			break
		}

		cover += curCell.GetCover()

		if x == tx {
			area := curCell.GetArea()
			for i++; i < numCells && cells[i].GetX() == x; i++ {
				area += cells[i].GetArea()
				cover += cells[i].GetCover()
			}

			alpha := r.CalculateAlpha((cover << (basics.PolySubpixelShift + 1)) - area)
			return alpha != 0
		}
	}

	alpha := r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
	return alpha != 0
}
