package rasterizer

import (
	"math"
	"sort"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// AA scale constants for compound rasterization
const (
	AAShift  = 8
	AAScale  = 1 << AAShift // 256
	AAMask   = AAScale - 1  // 255
	AAScale2 = AAScale * 2  // 512
	AAMask2  = AAScale2 - 1 // 511
)

// StyleInfo contains information about a rendering style during compound rasterization
type StyleInfo struct {
	StartCell uint32 // Starting index in the cell array
	NumCells  uint32 // Number of cells for this style
	LastX     int    // Last X coordinate processed for this style
}

// CellInfo represents a simplified cell for compound rasterization processing
type CellInfo struct {
	X     int // X coordinate
	Area  int // Area value
	Cover int // Coverage value
}

// CompoundScanlineInterface defines the interface for scanlines used with compound rasterization
type CompoundScanlineInterface interface {
	ResetSpans()
	AddCell(x int, cover basics.Int8u)
	AddSpan(x, len int, cover basics.Int8u)
	Finalize(y int)
	NumSpans() int
}

// ScanlineHitTest is a simple scanline implementation for hit testing.
// It checks if a specific X coordinate is covered by any rendered spans.
type ScanlineHitTest struct {
	x   int  // X coordinate to test for hit
	hit bool // Whether the test coordinate was hit
}

// NewScanlineHitTest creates a new hit test scanline for the given X coordinate
func NewScanlineHitTest(x int) *ScanlineHitTest {
	return &ScanlineHitTest{
		x:   x,
		hit: false,
	}
}

// ResetSpans resets the scanline for a new test, clearing the hit flag
func (s *ScanlineHitTest) ResetSpans() {
	s.hit = false
}

// AddCell adds a single cell to the scanline.
// If the cell's X coordinate matches our test coordinate, mark as hit.
func (s *ScanlineHitTest) AddCell(x int, cover basics.Int8u) {
	if x == s.x && cover > 0 {
		s.hit = true
	}
}

// AddSpan adds a span of consecutive pixels to the scanline.
// If our test X coordinate falls within this span, mark as hit.
func (s *ScanlineHitTest) AddSpan(x, len int, cover basics.Int8u) {
	if cover > 0 && s.x >= x && s.x < x+len {
		s.hit = true
	}
}

// Finalize completes the scanline processing (no-op for hit test)
func (s *ScanlineHitTest) Finalize(y int) {
	// No operation needed for hit testing
}

// Hit returns whether the test coordinate was hit by any rendered span
func (s *ScanlineHitTest) Hit() bool {
	return s.hit
}

// NumSpans returns the number of spans (always 0 or 1 for hit test)
func (s *ScanlineHitTest) NumSpans() int {
	if s.hit {
		return 1
	}
	return 0
}

// CompoundClipInterface defines the interface for clipping implementations used with compound rasterization
type CompoundClipInterface interface {
	ResetClipping()
	ClipBox(x1, y1, x2, y2 float64)
	MoveTo(x, y float64)
	LineTo(outline *RasterizerCellsAAStyled, x, y float64)
}

// RasterizerCompoundAA implements compound anti-aliased rasterization.
// This allows rendering multiple styles (fill colors, patterns, etc.) in a single pass.
// The Clip type parameter allows for different clipping implementations.
type RasterizerCompoundAA[Clip CompoundClipInterface] struct {
	// Core rasterization components
	outline     *RasterizerCellsAAStyled // Cell rasterizer with style support
	clipper     Clip                     // Clipping implementation
	fillingRule basics.FillingRule       // Polygon filling rule
	layerOrder  basics.LayerOrder        // Layer rendering order

	// Style management
	styles   *array.PodVector[StyleInfo] // Active styles information
	ast      *array.PodVector[uint32]    // Active Style Table (unique style IDs)
	asm      *array.PodVector[uint8]     // Active Style Mask (bitfield)
	cells    *array.PodVector[CellInfo]  // Processed cells for current scanline
	coverBuf *array.PodVector[uint8]     // Coverage buffer for spans

	// Style tracking
	minStyle int // Minimum style ID encountered
	maxStyle int // Maximum style ID encountered

	// Current processing state
	startX  int    // Starting X coordinate for current path
	startY  int    // Starting Y coordinate for current path
	scanY   uint32 // Current scanline Y coordinate
	slStart int    // Scanline start X coordinate
	slLen   uint32 // Scanline length
}

// NewRasterizerCompoundAA creates a new compound anti-aliased rasterizer
func NewRasterizerCompoundAA[Clip CompoundClipInterface](clipper Clip) *RasterizerCompoundAA[Clip] {
	return &RasterizerCompoundAA[Clip]{
		outline:     NewRasterizerCellsAAStyled(1024),
		clipper:     clipper,
		fillingRule: basics.FillNonZero,
		layerOrder:  basics.LayerDirect,
		styles:      array.NewPodVector[StyleInfo](),
		ast:         array.NewPodVector[uint32](),
		asm:         array.NewPodVector[uint8](),
		cells:       array.NewPodVector[CellInfo](),
		coverBuf:    array.NewPodVector[uint8](),
		minStyle:    math.MaxInt32,
		maxStyle:    math.MinInt32,
		startX:      0,
		startY:      0,
		scanY:       math.MaxUint32,
		slStart:     0,
		slLen:       0,
	}
}

// Reset clears all accumulated path data and resets the rasterizer state
func (r *RasterizerCompoundAA[Clip]) Reset() {
	r.outline.Reset()
	r.minStyle = math.MaxInt32
	r.maxStyle = math.MinInt32
	r.scanY = math.MaxUint32
	r.slStart = 0
	r.slLen = 0
}

// ResetClipping resets the clipping rectangle to unlimited
func (r *RasterizerCompoundAA[Clip]) ResetClipping() {
	r.Reset()
	r.clipper.ResetClipping()
}

// ClipBox sets the clipping rectangle
func (r *RasterizerCompoundAA[Clip]) ClipBox(x1, y1, x2, y2 float64) {
	r.Reset()
	r.clipper.ClipBox(x1, y1, x2, y2)
}

// FillingRule sets the polygon filling rule
func (r *RasterizerCompoundAA[Clip]) FillingRule(rule basics.FillingRule) {
	r.fillingRule = rule
}

// LayerOrder sets the layer rendering order
func (r *RasterizerCompoundAA[Clip]) LayerOrder(order basics.LayerOrder) {
	r.layerOrder = order
}

// Styles sets the left and right style identifiers for subsequent path operations
func (r *RasterizerCompoundAA[Clip]) Styles(left, right int) {
	cell := CellStyleAA{}
	cell.Initial()
	cell.Left = int16(left)
	cell.Right = int16(right)
	r.outline.Style(cell)

	// Update style range
	if left >= 0 && left < r.minStyle {
		r.minStyle = left
	}
	if left >= 0 && left > r.maxStyle {
		r.maxStyle = left
	}
	if right >= 0 && right < r.minStyle {
		r.minStyle = right
	}
	if right >= 0 && right > r.maxStyle {
		r.maxStyle = right
	}
}

// MoveTo starts a new subpath at the given coordinates
func (r *RasterizerCompoundAA[Clip]) MoveTo(x, y int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.startX = x
	r.startY = y
	r.clipper.MoveTo(float64(x), float64(y))
}

// LineTo adds a line segment to the current subpath
func (r *RasterizerCompoundAA[Clip]) LineTo(x, y int) {
	r.clipper.LineTo(r.outline, float64(x), float64(y))
}

// MoveToD starts a new subpath at the given double coordinates
func (r *RasterizerCompoundAA[Clip]) MoveToD(x, y float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.startX = basics.IRound(x * basics.PolySubpixelScale)
	r.startY = basics.IRound(y * basics.PolySubpixelScale)
	r.clipper.MoveTo(x, y)
}

// LineToD adds a line segment to the current subpath using double coordinates
func (r *RasterizerCompoundAA[Clip]) LineToD(x, y float64) {
	r.clipper.LineTo(r.outline, x, y)
}

// Edge adds a single edge from (x1,y1) to (x2,y2)
func (r *RasterizerCompoundAA[Clip]) Edge(x1, y1, x2, y2 int) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.clipper.MoveTo(float64(x1), float64(y1))
	r.clipper.LineTo(r.outline, float64(x2), float64(y2))
}

// EdgeD adds a single edge using double coordinates
func (r *RasterizerCompoundAA[Clip]) EdgeD(x1, y1, x2, y2 float64) {
	if r.outline.Sorted() {
		r.Reset()
	}
	r.clipper.MoveTo(x1, y1)
	r.clipper.LineTo(r.outline, x2, y2)
}

// AddVertex adds a vertex to the path based on the path command
func (r *RasterizerCompoundAA[Clip]) AddVertex(x, y float64, cmd uint32) {
	pathCmd := basics.PathCommand(cmd & uint32(basics.PathCmdMask))

	switch {
	case basics.IsMoveTo(pathCmd):
		r.MoveToD(x, y)
	case basics.IsVertex(pathCmd):
		r.LineToD(x, y)
	case basics.IsClose(cmd):
		r.LineToD(float64(r.startX)/basics.PolySubpixelScale, float64(r.startY)/basics.PolySubpixelScale)
	}
}

// AddPath adds a complete path from a vertex source
func (r *RasterizerCompoundAA[Clip]) AddPath(vs VertexSource, pathID uint32) {
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

// Bounds returns the bounding box of the rasterized geometry
func (r *RasterizerCompoundAA[Clip]) MinX() int     { return r.outline.MinX() }
func (r *RasterizerCompoundAA[Clip]) MinY() int     { return r.outline.MinY() }
func (r *RasterizerCompoundAA[Clip]) MaxX() int     { return r.outline.MaxX() }
func (r *RasterizerCompoundAA[Clip]) MaxY() int     { return r.outline.MaxY() }
func (r *RasterizerCompoundAA[Clip]) MinStyle() int { return r.minStyle }
func (r *RasterizerCompoundAA[Clip]) MaxStyle() int { return r.maxStyle }

// Sort sorts the accumulated cells for efficient scanline processing
func (r *RasterizerCompoundAA[Clip]) Sort() {
	r.outline.SortCells()
}

// HitTest checks if the given point is covered by any rendered geometry
func (r *RasterizerCompoundAA[Clip]) HitTest(tx, ty int) bool {
	if !r.NavigateScanline(uint32(ty)) {
		return false
	}

	numStyles := r.SweepStyles()
	if numStyles <= 0 {
		return false
	}

	sl := NewScanlineHitTest(tx)
	r.SweepScanline(sl, -1)
	return sl.Hit()
}

// NavigateScanline prepares the rasterizer for processing a specific scanline
func (r *RasterizerCompoundAA[Clip]) NavigateScanline(y uint32) bool {
	r.outline.SortCells()
	if r.outline.TotalCells() == 0 {
		return false
	}
	if r.maxStyle < r.minStyle {
		return false
	}
	if y < uint32(r.outline.MinY()) || y > uint32(r.outline.MaxY()) {
		return false
	}
	r.scanY = y
	r.styles.Allocate(r.maxStyle-r.minStyle+2, 128)
	return true
}

// RewindScanlines prepares for scanline iteration from the beginning
func (r *RasterizerCompoundAA[Clip]) RewindScanlines() bool {
	r.outline.SortCells()
	if r.outline.TotalCells() == 0 {
		return false
	}
	if r.maxStyle < r.minStyle {
		return false
	}
	r.scanY = uint32(r.outline.MinY())
	r.styles.Allocate(r.maxStyle-r.minStyle+2, 128)
	return true
}

// addStyle adds a style to the active style tracking
func (r *RasterizerCompoundAA[Clip]) addStyle(styleID int) {
	if styleID < 0 {
		styleID = 0
	} else {
		styleID -= r.minStyle - 1
	}

	nbyte := styleID >> 3
	mask := uint8(1 << (styleID & 7))

	style := &r.styles.Data()[styleID]
	if r.asm.Size() <= nbyte {
		r.asm.Resize(nbyte + 1)
	}
	if (r.asm.Data()[nbyte] & mask) == 0 {
		r.ast.Add(uint32(styleID))
		r.asm.Data()[nbyte] |= mask
		style.StartCell = 0
		style.NumCells = 0
		style.LastX = math.MinInt32
	}
	style.StartCell++
}

// SweepStyles processes all styles for the current scanline and returns the count
func (r *RasterizerCompoundAA[Clip]) SweepStyles() uint32 {
	for {
		if r.scanY > uint32(r.outline.MaxY()) {
			return 0
		}

		numCells := r.outline.ScanlineNumCells(r.scanY)
		cells := r.outline.ScanlineCells(r.scanY)
		numStyles := r.maxStyle - r.minStyle + 2

		r.cells.Allocate(int(numCells*2), 256) // Each cell can have two styles
		r.ast.SetCapacity(numStyles, 64)
		r.asm.Allocate((numStyles+7)>>3, 8)
		r.asm.Zero()

		if numCells > 0 {
			// Pre-add zero style for no-fill (-1 style)
			r.asm.Data()[0] |= 1
			r.ast.Add(0)
			style := &r.styles.Data()[0]
			style.StartCell = 0
			style.NumCells = 0
			style.LastX = math.MinInt32

			r.slStart = (*cells[0]).GetX()
			r.slLen = uint32((*cells[numCells-1]).GetX() - r.slStart + 1)

			// Add all styles from cells
			for i := 0; i < int(numCells); i++ {
				currCell := cells[i]
				r.addStyle(int((*currCell).Left))
				r.addStyle(int((*currCell).Right))
			}

			// Convert Y-histogram into array of starting indexes
			startCell := uint32(0)
			for i := 0; i < r.ast.Size(); i++ {
				st := &r.styles.Data()[r.ast.Data()[i]]
				v := st.StartCell
				st.StartCell = startCell
				startCell += v
			}

			// Process cells and distribute to styles
			for i := 0; i < int(numCells); i++ {
				currCell := cells[i]

				// Process left style
				styleID := 0
				if (*currCell).Left >= 0 {
					styleID = int((*currCell).Left) - r.minStyle + 1
				}

				style := &r.styles.Data()[styleID]
				if (*currCell).GetX() == style.LastX {
					cell := &r.cells.Data()[style.StartCell+style.NumCells-1]
					cell.Area += (*currCell).GetArea()
					cell.Cover += (*currCell).GetCover()
				} else {
					cell := &r.cells.Data()[style.StartCell+style.NumCells]
					cell.X = (*currCell).GetX()
					cell.Area = (*currCell).GetArea()
					cell.Cover = (*currCell).GetCover()
					style.LastX = (*currCell).GetX()
					style.NumCells++
				}

				// Process right style
				styleID = 0
				if (*currCell).Right >= 0 {
					styleID = int((*currCell).Right) - r.minStyle + 1
				}

				style = &r.styles.Data()[styleID]
				if (*currCell).GetX() == style.LastX {
					cell := &r.cells.Data()[style.StartCell+style.NumCells-1]
					cell.Area -= (*currCell).GetArea()
					cell.Cover -= (*currCell).GetCover()
				} else {
					cell := &r.cells.Data()[style.StartCell+style.NumCells]
					cell.X = (*currCell).GetX()
					cell.Area = -(*currCell).GetArea()
					cell.Cover = -(*currCell).GetCover()
					style.LastX = (*currCell).GetX()
					style.NumCells++
				}
			}
		}

		if r.ast.Size() > 1 {
			break
		}
		r.scanY++
	}
	r.scanY++

	// Sort styles according to layer order
	if r.layerOrder != basics.LayerUnsorted {
		astData := r.ast.Data()[1:] // Skip the first element (style 0)
		if r.layerOrder == basics.LayerDirect {
			sort.Slice(astData, func(i, j int) bool { return astData[i] > astData[j] })
		} else {
			sort.Slice(astData, func(i, j int) bool { return astData[i] < astData[j] })
		}
	}

	return uint32(r.ast.Size() - 1)
}

// ScanlineStart returns the starting X coordinate of the current scanline
func (r *RasterizerCompoundAA[Clip]) ScanlineStart() int {
	return r.slStart
}

// ScanlineLength returns the length of the current scanline
func (r *RasterizerCompoundAA[Clip]) ScanlineLength() uint32 {
	return r.slLen
}

// Style returns the style ID for the given style index
func (r *RasterizerCompoundAA[Clip]) Style(styleIdx uint32) uint32 {
	return r.ast.Data()[styleIdx+1] + uint32(r.minStyle) - 1
}

// CalculateAlpha calculates the alpha value from the area
func (r *RasterizerCompoundAA[Clip]) CalculateAlpha(area int) uint32 {
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

// SweepScanline sweeps one scanline with the specified style index
func (r *RasterizerCompoundAA[Clip]) SweepScanline(sl CompoundScanlineInterface, styleIdx int) bool {
	scanY := r.scanY - 1
	if scanY > uint32(r.outline.MaxY()) {
		return false
	}

	sl.ResetSpans()

	if styleIdx < 0 {
		styleIdx = 0
	} else {
		styleIdx++
	}

	st := &r.styles.Data()[r.ast.Data()[styleIdx]]
	numCells := st.NumCells
	cellPtr := st.StartCell

	cover := 0
	for i := uint32(0); i < numCells; i++ {
		cell := &r.cells.Data()[cellPtr+i]
		x := cell.X
		area := cell.Area

		cover += cell.Cover

		if area != 0 {
			alpha := r.CalculateAlpha((cover << (basics.PolySubpixelShift + 1)) - area)
			sl.AddCell(x, basics.Int8u(alpha))
			x++
		}

		if i+1 < numCells {
			nextCell := &r.cells.Data()[cellPtr+i+1]
			if nextCell.X > x {
				alpha := r.CalculateAlpha(cover << (basics.PolySubpixelShift + 1))
				if alpha > 0 {
					sl.AddSpan(x, nextCell.X-x, basics.Int8u(alpha))
				}
			}
		}
	}

	if sl.NumSpans() == 0 {
		return false
	}
	sl.Finalize(int(scanY))
	return true
}

// AllocateCoverBuffer allocates a cover buffer of the specified length
func (r *RasterizerCompoundAA[Clip]) AllocateCoverBuffer(len int) []uint8 {
	r.coverBuf.Allocate(len, 256)
	return r.coverBuf.Data()
}
