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
// It uses integer coordinates in format 24.8 (24 bits integer, 8 bits fractional).
// This is equivalent to AGG's rasterizer_scanline_aa<Clip> template class.
type RasterizerScanlineAA[Clip ClipInterface, Conv ConverterInterface] struct {
	outline     *RasterizerCellsAASimple // Cell-based rasterizer
	clipper     Clip                     // Clipping implementation
	gamma       [AAScale]int             // Gamma correction table
	fillingRule basics.FillingRule       // Filling rule (non-zero or even-odd)
	autoClose   bool                     // Auto-close polygons flag
	startX      int                      // Starting X coordinate (in converter coord_type)
	startY      int                      // Starting Y coordinate (in converter coord_type)
	status      Status                   // Current rasterizer status
	scanY       int                      // Current scanline Y coordinate
}

// NewRasterizerScanlineAA creates a new anti-aliased scanline rasterizer
func NewRasterizerScanlineAA[Clip ClipInterface, Conv ConverterInterface](cellBlockLimit uint32) *RasterizerScanlineAA[Clip, Conv] {
	r := &RasterizerScanlineAA[Clip, Conv]{
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
func NewRasterizerScanlineAAWithGamma[Clip ClipInterface, Conv ConverterInterface](gammaFunc func(float64) float64, cellBlockLimit uint32) *RasterizerScanlineAA[Clip, Conv] {
	r := NewRasterizerScanlineAA[Clip, Conv](cellBlockLimit)
	r.SetGamma(gammaFunc)
	return r
}

// Reset clears the rasterizer and prepares it for new geometry
func (r *RasterizerScanlineAA[Clip, Conv]) Reset() {
	r.outline.Reset()
	r.status = StatusInitial
}

// ResetClipping resets the clipping settings
func (r *RasterizerScanlineAA[Clip, Conv]) ResetClipping() {
	r.Reset()
	r.clipper.ResetClipping()
}

// ClipBox sets the clipping rectangle
func (r *RasterizerScanlineAA[Clip, Conv]) ClipBox(x1, y1, x2, y2 float64) {
	r.Reset()
	// Pass coordinates directly to clipper without converter transformation
	// The clipper interface expects float64 coordinates
	r.clipper.ClipBox(x1, y1, x2, y2)
}

// FillingRule sets the polygon filling rule
func (r *RasterizerScanlineAA[Clip, Conv]) FillingRule(rule basics.FillingRule) {
	r.fillingRule = rule
}

// AutoClose sets whether polygons should be automatically closed
func (r *RasterizerScanlineAA[Clip, Conv]) AutoClose(flag bool) {
	r.autoClose = flag
}

// SetGamma sets the gamma correction function
func (r *RasterizerScanlineAA[Clip, Conv]) SetGamma(gammaFunc func(float64) float64) {
	for i := 0; i < AAScale; i++ {
		r.gamma[i] = int(basics.URound(gammaFunc(float64(i)/AAMask) * AAMask))
	}
}

// ApplyGamma applies gamma correction to a coverage value
func (r *RasterizerScanlineAA[Clip, Conv]) ApplyGamma(cover uint32) uint32 {
	return uint32(r.gamma[cover])
}

// MoveTo starts a new contour at the specified integer coordinates
func (r *RasterizerScanlineAA[Clip, Conv]) MoveTo(x, y int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}

	var conv Conv
	// Handle different converter return types
	if startX, ok := conv.Downscale(x).(int); ok {
		r.startX = startX
	} else if startXF, ok := conv.Downscale(x).(float64); ok {
		r.startX = int(startXF)
	}
	if startY, ok := conv.Downscale(y).(int); ok {
		r.startY = startY
	} else if startYF, ok := conv.Downscale(y).(float64); ok {
		r.startY = int(startYF)
	}
	r.clipper.MoveTo(float64(r.startX), float64(r.startY))
	r.status = StatusMoveTo
}

// LineTo draws a line to the specified integer coordinates
func (r *RasterizerScanlineAA[Clip, Conv]) LineTo(x, y int) {
	var conv Conv
	// Handle different converter return types
	var xDown, yDown float64
	if xDownI, ok := conv.Downscale(x).(int); ok {
		xDown = float64(xDownI)
	} else if xDownF, ok := conv.Downscale(x).(float64); ok {
		xDown = xDownF
	}
	if yDownI, ok := conv.Downscale(y).(int); ok {
		yDown = float64(yDownI)
	} else if yDownF, ok := conv.Downscale(y).(float64); ok {
		yDown = yDownF
	}
	r.clipper.LineTo(r.outline, xDown, yDown)
	r.status = StatusLineTo
}

// MoveToD starts a new contour at the specified double coordinates
func (r *RasterizerScanlineAA[Clip, Conv]) MoveToD(x, y float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	if r.autoClose {
		r.ClosePolygon()
	}

	var conv Conv
	// Handle different converter return types
	if startX, ok := conv.Upscale(x).(int); ok {
		r.startX = startX
	} else if startXF, ok := conv.Upscale(x).(float64); ok {
		r.startX = int(startXF)
	}
	if startY, ok := conv.Upscale(y).(int); ok {
		r.startY = startY
	} else if startYF, ok := conv.Upscale(y).(float64); ok {
		r.startY = int(startYF)
	}
	r.clipper.MoveTo(x, y)
	r.status = StatusMoveTo
}

// LineToD draws a line to the specified double coordinates
func (r *RasterizerScanlineAA[Clip, Conv]) LineToD(x, y float64) {
	r.clipper.LineTo(r.outline, x, y)
	r.status = StatusLineTo
}

// ClosePolygon closes the current polygon
func (r *RasterizerScanlineAA[Clip, Conv]) ClosePolygon() {
	if r.status == StatusLineTo {
		r.clipper.LineTo(r.outline, float64(r.startX)/basics.PolySubpixelScale, float64(r.startY)/basics.PolySubpixelScale)
		r.status = StatusClosed
	}
}

// AddVertex adds a vertex with the specified command
func (r *RasterizerScanlineAA[Clip, Conv]) AddVertex(x, y float64, cmd uint32) {
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
func (r *RasterizerScanlineAA[Clip, Conv]) Edge(x1, y1, x2, y2 int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	var conv Conv
	// Handle different converter return types
	var x1Down, y1Down, x2Down, y2Down float64
	if x1DownI, ok := conv.Downscale(x1).(int); ok {
		x1Down = float64(x1DownI)
	} else if x1DownF, ok := conv.Downscale(x1).(float64); ok {
		x1Down = x1DownF
	}
	if y1DownI, ok := conv.Downscale(y1).(int); ok {
		y1Down = float64(y1DownI)
	} else if y1DownF, ok := conv.Downscale(y1).(float64); ok {
		y1Down = y1DownF
	}
	if x2DownI, ok := conv.Downscale(x2).(int); ok {
		x2Down = float64(x2DownI)
	} else if x2DownF, ok := conv.Downscale(x2).(float64); ok {
		x2Down = x2DownF
	}
	if y2DownI, ok := conv.Downscale(y2).(int); ok {
		y2Down = float64(y2DownI)
	} else if y2DownF, ok := conv.Downscale(y2).(float64); ok {
		y2Down = y2DownF
	}
	r.clipper.MoveTo(x1Down, y1Down)
	r.clipper.LineTo(r.outline, x2Down, y2Down)
}

// EdgeD adds a single edge from (x1,y1) to (x2,y2) with double coordinates
func (r *RasterizerScanlineAA[Clip, Conv]) EdgeD(x1, y1, x2, y2 float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	var conv Conv
	// Handle different converter return types
	var x1Up, y1Up, x2Up, y2Up float64
	if x1UpI, ok := conv.Upscale(x1).(int); ok {
		x1Up = float64(x1UpI)
	} else if x1UpF, ok := conv.Upscale(x1).(float64); ok {
		x1Up = x1UpF
	}
	if y1UpI, ok := conv.Upscale(y1).(int); ok {
		y1Up = float64(y1UpI)
	} else if y1UpF, ok := conv.Upscale(y1).(float64); ok {
		y1Up = y1UpF
	}
	if x2UpI, ok := conv.Upscale(x2).(int); ok {
		x2Up = float64(x2UpI)
	} else if x2UpF, ok := conv.Upscale(x2).(float64); ok {
		x2Up = x2UpF
	}
	if y2UpI, ok := conv.Upscale(y2).(int); ok {
		y2Up = float64(y2UpI)
	} else if y2UpF, ok := conv.Upscale(y2).(float64); ok {
		y2Up = y2UpF
	}
	r.clipper.MoveTo(x1Up, y1Up)
	r.clipper.LineTo(r.outline, x2Up, y2Up)
}

// AddPath adds all vertices from a vertex source
func (r *RasterizerScanlineAA[Clip, Conv]) AddPath(vs VertexSource, pathID uint32) {
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

// MinX returns the minimum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip, Conv]) MinX() int {
	return r.outline.MinX()
}

// MinY returns the minimum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip, Conv]) MinY() int {
	return r.outline.MinY()
}

// MaxX returns the maximum X coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip, Conv]) MaxX() int {
	return r.outline.MaxX()
}

// MaxY returns the maximum Y coordinate of the rasterized geometry
func (r *RasterizerScanlineAA[Clip, Conv]) MaxY() int {
	return r.outline.MaxY()
}

// Sort sorts the cells in preparation for scanline rendering
func (r *RasterizerScanlineAA[Clip, Conv]) Sort() {
	if r.autoClose {
		r.ClosePolygon()
	}
	r.outline.SortCells()
}

// RewindScanlines resets the scanline iterator
func (r *RasterizerScanlineAA[Clip, Conv]) RewindScanlines() bool {
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
func (r *RasterizerScanlineAA[Clip, Conv]) NavigateScanline(y int) bool {
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
func (r *RasterizerScanlineAA[Clip, Conv]) CalculateAlpha(area int) uint32 {
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
func (r *RasterizerScanlineAA[Clip, Conv]) SweepScanline(sl ScanlineInterface) bool {
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
					sl.AddCell(x, alpha)
				}
				x++
			}

			if cellIndex < numCells && cells[cellIndex].X > x {
				alpha := r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
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

// HitTest performs a hit test at the specified coordinates
func (r *RasterizerScanlineAA[Clip, Conv]) HitTest(tx, ty int) bool {
	if !r.NavigateScanline(ty) {
		return false
	}

	numCells := r.outline.ScanlineNumCells(uint32(ty))
	cells := r.outline.ScanlineCells(uint32(ty))
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
