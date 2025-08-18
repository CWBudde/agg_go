// Package scanline provides boolean algebra operations for scanline containers.
// This implements the functionality from AGG's agg_scanline_boolean_algebra.h
// allowing intersection, union, XOR, and subtraction operations on scanlines.
package scanline

import (
	"agg_go/internal/basics"
	"agg_go/internal/renderer/scanline"
)

// Constants for coverage calculations
const (
	CoverShift = basics.CoverShift
	CoverSize  = basics.CoverSize
	CoverMask  = basics.CoverMask
	CoverFull  = basics.CoverFull
)

// BooleanScanlineInterface represents any scanline container that can be used
// in boolean operations. This includes the core scanline interface methods
// and additional methods needed for boolean algebra.
type BooleanScanlineInterface interface {
	// Core scanline interface methods
	Y() int
	NumSpans() int
	Begin() scanline.ScanlineIterator

	// Boolean algebra specific methods
	ResetSpans()
	AddCell(x int, cover uint)
	AddCells(x, length int, covers []basics.Int8u)
	AddSpan(x, length int, cover basics.Int8u)
	Finalize(y int)
}

// RasterizerInterface represents scanline generators used in shape operations
type RasterizerInterface interface {
	RewindScanlines() bool
	SweepScanline(sl BooleanScanlineInterface) bool
	MinX() int
	MinY() int
	MaxX() int
	MaxY() int
}

// RendererInterface represents a renderer that can render scanlines
type RendererInterface interface {
	Prepare()
	Render(sl BooleanScanlineInterface)
}

// IteratorInterface represents a scanline span iterator
type IteratorInterface interface {
	X() int
	Len() int
	Covers() []basics.Int8u
}

// ==============================================================================
// Boolean Algebra Function Types (Functors)
// ==============================================================================

// CombineSpansBinFunc combines two binary encoded spans (no anti-aliasing)
type CombineSpansBinFunc func(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface)

// CombineSpansEmptyFunc is used for XOR binary spans (does nothing)
type CombineSpansEmptyFunc func(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface)

// AddSpanEmptyFunc adds nothing (used in combine_shapes_sub)
type AddSpanEmptyFunc func(span IteratorInterface, x int, length uint, sl BooleanScanlineInterface)

// AddSpanBinFunc adds a binary span
type AddSpanBinFunc func(span IteratorInterface, x int, length uint, sl BooleanScanlineInterface)

// AddSpanAAFunc adds an anti-aliased span
type AddSpanAAFunc func(span IteratorInterface, x int, length uint, sl BooleanScanlineInterface)

// CombineSpansAAFunc combines spans with anti-aliasing preservation
type CombineSpansAAFunc func(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface)

// ==============================================================================
// Default Functor Implementations
// ==============================================================================

// CombineSpansBin combines two binary encoded spans
func CombineSpansBin(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
}

// CombineSpansEmpty does nothing (used for XOR binary spans)
func CombineSpansEmpty(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	// Do nothing
}

// AddSpanEmpty adds nothing
func AddSpanEmpty(span IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	// Do nothing
}

// AddSpanBin adds a binary span
func AddSpanBin(span IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
}

// AddSpanAA adds an anti-aliased span
func AddSpanAA(span IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	if span.Len() < 0 {
		// Solid span
		if len(span.Covers()) > 0 {
			sl.AddSpan(x, int(length), span.Covers()[0])
		}
	} else if span.Len() > 0 {
		// Anti-aliased span
		covers := span.Covers()
		if span.X() < x {
			// Adjust covers array if span starts before our region
			offset := x - span.X()
			if offset < len(covers) {
				covers = covers[offset:]
			}
		}
		sl.AddCells(x, int(length), covers)
	}
}

// IntersectSpansAA intersects two spans preserving anti-aliasing information
func IntersectSpansAA(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	len1 := span1.Len()
	len2 := span2.Len()
	covers1 := span1.Covers()
	covers2 := span2.Covers()

	// Calculate operation code based on span types
	// 0 = Both spans are AA type
	// 1 = span1 is solid, span2 is AA
	// 2 = span1 is AA, span2 is solid
	// 3 = Both spans are solid
	opCode := 0
	if len1 < 0 {
		opCode |= 1
	}
	if len2 < 0 {
		opCode |= 2
	}

	switch opCode {
	case 0: // Both are AA spans
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}

		for i := 0; i < int(length); i++ {
			cover1 := uint(0)
			cover2 := uint(0)
			if i < len(covers1) {
				cover1 = uint(covers1[i])
			}
			if i < len(covers2) {
				cover2 = uint(covers2[i])
			}

			cover := cover1 * cover2
			if cover == CoverFull*CoverFull {
				sl.AddCell(x+i, CoverFull)
			} else {
				sl.AddCell(x+i, cover>>CoverShift)
			}
		}

	case 1: // span1 is solid, span2 is AA
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}
		if len(covers1) > 0 && covers1[0] == CoverFull {
			sl.AddCells(x, int(length), covers2)
		} else {
			cover1 := uint(0)
			if len(covers1) > 0 {
				cover1 = uint(covers1[0])
			}
			for i := 0; i < int(length); i++ {
				cover2 := uint(0)
				if i < len(covers2) {
					cover2 = uint(covers2[i])
				}
				cover := cover1 * cover2
				if cover == CoverFull*CoverFull {
					sl.AddCell(x+i, CoverFull)
				} else {
					sl.AddCell(x+i, cover>>CoverShift)
				}
			}
		}

	case 2: // span1 is AA, span2 is solid
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		if len(covers2) > 0 && covers2[0] == CoverFull {
			sl.AddCells(x, int(length), covers1)
		} else {
			cover2 := uint(0)
			if len(covers2) > 0 {
				cover2 = uint(covers2[0])
			}
			for i := 0; i < int(length); i++ {
				cover1 := uint(0)
				if i < len(covers1) {
					cover1 = uint(covers1[i])
				}
				cover := cover1 * cover2
				if cover == CoverFull*CoverFull {
					sl.AddCell(x+i, CoverFull)
				} else {
					sl.AddCell(x+i, cover>>CoverShift)
				}
			}
		}

	case 3: // Both are solid spans
		cover1 := uint(0)
		cover2 := uint(0)
		if len(covers1) > 0 {
			cover1 = uint(covers1[0])
		}
		if len(covers2) > 0 {
			cover2 = uint(covers2[0])
		}
		cover := cover1 * cover2
		if cover == CoverFull*CoverFull {
			sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
		} else {
			sl.AddSpan(x, int(length), basics.Int8u(cover>>CoverShift))
		}
	}
}

// UniteSpansAA unites two spans preserving anti-aliasing information
func UniteSpansAA(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	len1 := span1.Len()
	len2 := span2.Len()
	covers1 := span1.Covers()
	covers2 := span2.Covers()

	// Calculate operation code based on span types
	opCode := 0
	if len1 < 0 {
		opCode |= 1
	}
	if len2 < 0 {
		opCode |= 2
	}

	switch opCode {
	case 0: // Both are AA spans
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}

		for i := 0; i < int(length); i++ {
			cover1 := uint(0)
			cover2 := uint(0)
			if i < len(covers1) {
				cover1 = uint(covers1[i])
			}
			if i < len(covers2) {
				cover2 = uint(covers2[i])
			}

			// Union formula: CoverMask * CoverMask - (CoverMask - a) * (CoverMask - b)
			cover := CoverMask*CoverMask - (CoverMask-cover1)*(CoverMask-cover2)
			if cover == CoverFull*CoverFull {
				sl.AddCell(x+i, CoverFull)
			} else {
				sl.AddCell(x+i, cover>>CoverShift)
			}
		}

	case 1: // span1 is solid, span2 is AA
		if len(covers1) > 0 && covers1[0] == CoverFull {
			sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
		} else {
			if span2.X() < x {
				offset := x - span2.X()
				if offset < len(covers2) {
					covers2 = covers2[offset:]
				}
			}
			cover1 := uint(0)
			if len(covers1) > 0 {
				cover1 = uint(covers1[0])
			}
			for i := 0; i < int(length); i++ {
				cover2 := uint(0)
				if i < len(covers2) {
					cover2 = uint(covers2[i])
				}
				cover := CoverMask*CoverMask - (CoverMask-cover1)*(CoverMask-cover2)
				if cover == CoverFull*CoverFull {
					sl.AddCell(x+i, CoverFull)
				} else {
					sl.AddCell(x+i, cover>>CoverShift)
				}
			}
		}

	case 2: // span1 is AA, span2 is solid
		if len(covers2) > 0 && covers2[0] == CoverFull {
			sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
		} else {
			if span1.X() < x {
				offset := x - span1.X()
				if offset < len(covers1) {
					covers1 = covers1[offset:]
				}
			}
			cover2 := uint(0)
			if len(covers2) > 0 {
				cover2 = uint(covers2[0])
			}
			for i := 0; i < int(length); i++ {
				cover1 := uint(0)
				if i < len(covers1) {
					cover1 = uint(covers1[i])
				}
				cover := CoverMask*CoverMask - (CoverMask-cover1)*(CoverMask-cover2)
				if cover == CoverFull*CoverFull {
					sl.AddCell(x+i, CoverFull)
				} else {
					sl.AddCell(x+i, cover>>CoverShift)
				}
			}
		}

	case 3: // Both are solid spans
		cover1 := uint(0)
		cover2 := uint(0)
		if len(covers1) > 0 {
			cover1 = uint(covers1[0])
		}
		if len(covers2) > 0 {
			cover2 = uint(covers2[0])
		}
		cover := CoverMask*CoverMask - (CoverMask-cover1)*(CoverMask-cover2)
		if cover == CoverFull*CoverFull {
			sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
		} else {
			sl.AddSpan(x, int(length), basics.Int8u(cover>>CoverShift))
		}
	}
}

// ==============================================================================
// XOR Formula Types
// ==============================================================================

// XorFormula defines the interface for XOR calculation formulas
type XorFormula interface {
	Calculate(a, b uint) uint
}

// XorFormulaLinear implements linear XOR formula
type XorFormulaLinear struct{}

func (XorFormulaLinear) Calculate(a, b uint) uint {
	cover := a + b
	if cover > CoverMask {
		cover = CoverMask + CoverMask - cover
	}
	return cover
}

// XorFormulaSaddle implements saddle XOR formula
type XorFormulaSaddle struct{}

func (XorFormulaSaddle) Calculate(a, b uint) uint {
	k := a * b
	if k == CoverMask*CoverMask {
		return 0
	}

	// Handle the edge cases where inputs are 0 or 255
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}

	// Use the original formula with proper integer handling
	newA := (CoverMask*CoverMask - (a << CoverShift) + k) >> CoverShift
	newB := (CoverMask*CoverMask - (b << CoverShift) + k) >> CoverShift
	return CoverMask - ((newA * newB) >> CoverShift)
}

// XorFormulaAbsDiff implements absolute difference XOR formula
type XorFormulaAbsDiff struct{}

func (XorFormulaAbsDiff) Calculate(a, b uint) uint {
	if a > b {
		return a - b
	}
	return b - a
}

// XorSpansAA applies XOR to two spans using the specified formula
func XorSpansAA(formula XorFormula, span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	len1 := span1.Len()
	len2 := span2.Len()
	covers1 := span1.Covers()
	covers2 := span2.Covers()

	// Calculate operation code based on span types
	opCode := 0
	if len1 < 0 {
		opCode |= 1
	}
	if len2 < 0 {
		opCode |= 2
	}

	switch opCode {
	case 0: // Both are AA spans
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}

		for i := 0; i < int(length); i++ {
			cover1 := uint(0)
			cover2 := uint(0)
			if i < len(covers1) {
				cover1 = uint(covers1[i])
			}
			if i < len(covers2) {
				cover2 = uint(covers2[i])
			}

			cover := formula.Calculate(cover1, cover2)
			if cover > 0 {
				sl.AddCell(x+i, cover)
			}
		}

	case 1: // span1 is solid, span2 is AA
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}
		cover1 := uint(0)
		if len(covers1) > 0 {
			cover1 = uint(covers1[0])
		}
		for i := 0; i < int(length); i++ {
			cover2 := uint(0)
			if i < len(covers2) {
				cover2 = uint(covers2[i])
			}
			cover := formula.Calculate(cover1, cover2)
			if cover > 0 {
				sl.AddCell(x+i, cover)
			}
		}

	case 2: // span1 is AA, span2 is solid
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		cover2 := uint(0)
		if len(covers2) > 0 {
			cover2 = uint(covers2[0])
		}
		for i := 0; i < int(length); i++ {
			cover1 := uint(0)
			if i < len(covers1) {
				cover1 = uint(covers1[i])
			}
			cover := formula.Calculate(cover1, cover2)
			if cover > 0 {
				sl.AddCell(x+i, cover)
			}
		}

	case 3: // Both are solid spans
		cover1 := uint(0)
		cover2 := uint(0)
		if len(covers1) > 0 {
			cover1 = uint(covers1[0])
		}
		if len(covers2) > 0 {
			cover2 = uint(covers2[0])
		}
		cover := formula.Calculate(cover1, cover2)
		if cover > 0 {
			sl.AddSpan(x, int(length), basics.Int8u(cover))
		}
	}
}

// SubtractSpansAA subtracts spans preserving anti-aliasing information
func SubtractSpansAA(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
	len1 := span1.Len()
	len2 := span2.Len()
	covers1 := span1.Covers()
	covers2 := span2.Covers()

	// Calculate operation code based on span types
	opCode := 0
	if len1 < 0 {
		opCode |= 1
	}
	if len2 < 0 {
		opCode |= 2
	}

	switch opCode {
	case 0: // Both are AA spans
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}

		for i := 0; i < int(length); i++ {
			cover1 := uint(0)
			cover2 := uint(0)
			if i < len(covers1) {
				cover1 = uint(covers1[i])
			}
			if i < len(covers2) {
				cover2 = uint(covers2[i])
			}

			cover := cover1 * (CoverMask - cover2)
			if cover > 0 {
				if cover == CoverFull*CoverFull {
					sl.AddCell(x+i, CoverFull)
				} else {
					sl.AddCell(x+i, cover>>CoverShift)
				}
			}
		}

	case 1: // span1 is solid, span2 is AA
		if span2.X() < x {
			offset := x - span2.X()
			if offset < len(covers2) {
				covers2 = covers2[offset:]
			}
		}
		cover1 := uint(0)
		if len(covers1) > 0 {
			cover1 = uint(covers1[0])
		}
		for i := 0; i < int(length); i++ {
			cover2 := uint(0)
			if i < len(covers2) {
				cover2 = uint(covers2[i])
			}
			cover := cover1 * (CoverMask - cover2)
			if cover > 0 {
				if cover == CoverFull*CoverFull {
					sl.AddCell(x+i, CoverFull)
				} else {
					sl.AddCell(x+i, cover>>CoverShift)
				}
			}
		}

	case 2: // span1 is AA, span2 is solid
		if span1.X() < x {
			offset := x - span1.X()
			if offset < len(covers1) {
				covers1 = covers1[offset:]
			}
		}
		cover2 := uint(0)
		if len(covers2) > 0 {
			cover2 = uint(covers2[0])
		}
		if cover2 != CoverFull {
			for i := 0; i < int(length); i++ {
				cover1 := uint(0)
				if i < len(covers1) {
					cover1 = uint(covers1[i])
				}
				cover := cover1 * (CoverMask - cover2)
				if cover > 0 {
					if cover == CoverFull*CoverFull {
						sl.AddCell(x+i, CoverFull)
					} else {
						sl.AddCell(x+i, cover>>CoverShift)
					}
				}
			}
		}

	case 3: // Both are solid spans
		cover1 := uint(0)
		cover2 := uint(0)
		if len(covers1) > 0 {
			cover1 = uint(covers1[0])
		}
		if len(covers2) > 0 {
			cover2 = uint(covers2[0])
		}
		cover := cover1 * (CoverMask - cover2)
		if cover > 0 {
			if cover == CoverFull*CoverFull {
				sl.AddSpan(x, int(length), basics.Int8u(CoverFull))
			} else {
				sl.AddSpan(x, int(length), basics.Int8u(cover>>CoverShift))
			}
		}
	}
}

// ==============================================================================
// Main Boolean Algebra Algorithms
// ==============================================================================

// AddSpansAndRender adds spans from a scanline to output using the provided functor
func AddSpansAndRender(sl1 BooleanScanlineInterface, sl BooleanScanlineInterface, ren RendererInterface, addSpanFunc AddSpanAAFunc) {
	sl.ResetSpans()
	spanIterator := sl1.Begin()
	numSpans := sl1.NumSpans()

	for i := 0; i < numSpans; i++ {
		spanData := spanIterator.GetSpan()
		length := spanData.Len
		if length < 0 {
			length = -length
		}

		// Create a simple iterator wrapper for the span data
		spanIter := &simpleIterator{
			x:      spanData.X,
			len:    spanData.Len,
			covers: spanData.Covers,
		}

		addSpanFunc(spanIter, spanData.X, uint(length), sl)
		spanIterator.Next()
	}
	sl.Finalize(sl1.Y())
	ren.Render(sl)
}

// simpleIterator implements IteratorInterface for span data
type simpleIterator struct {
	x      int
	len    int
	covers []basics.Int8u
}

func (it *simpleIterator) X() int                 { return it.x }
func (it *simpleIterator) Len() int               { return it.len }
func (it *simpleIterator) Covers() []basics.Int8u { return it.covers }

// IntersectScanlines intersects two scanlines and generates a new one
func IntersectScanlines(sl1, sl2, sl BooleanScanlineInterface, combineSpansFunc CombineSpansAAFunc) {
	sl.ResetSpans()

	num1 := sl1.NumSpans()
	if num1 == 0 {
		return
	}

	num2 := sl2.NumSpans()
	if num2 == 0 {
		return
	}

	span1 := sl1.Begin()
	span2 := sl2.Begin()

	for num1 > 0 && num2 > 0 {
		span1Data := span1.GetSpan()
		span2Data := span2.GetSpan()

		xb1 := span1Data.X
		xb2 := span2Data.X

		len1 := span1Data.Len
		len2 := span2Data.Len
		if len1 < 0 {
			len1 = -len1
		}
		if len2 < 0 {
			len2 = -len2
		}

		xe1 := xb1 + len1 - 1
		xe2 := xb2 + len2 - 1

		// Determine which spans to advance
		advanceSpan1 := xe1 < xe2
		advanceBoth := xe1 == xe2

		// Find intersection
		if xb1 < xb2 {
			xb1 = xb2
		}
		if xe1 > xe2 {
			xe1 = xe2
		}

		if xb1 <= xe1 {
			span1Iter := &simpleIterator{
				x:      span1Data.X,
				len:    span1Data.Len,
				covers: span1Data.Covers,
			}
			span2Iter := &simpleIterator{
				x:      span2Data.X,
				len:    span2Data.Len,
				covers: span2Data.Covers,
			}
			combineSpansFunc(span1Iter, span2Iter, xb1, uint(xe1-xb1+1), sl)
		}

		// Advance spans
		if advanceBoth {
			num1--
			num2--
			if num1 > 0 {
				span1.Next()
			}
			if num2 > 0 {
				span2.Next()
			}
		} else if advanceSpan1 {
			num1--
			if num1 > 0 {
				span1.Next()
			}
		} else {
			num2--
			if num2 > 0 {
				span2.Next()
			}
		}
	}
}

// UniteScanlines unites two scanlines and generates a new one
func UniteScanlines(sl1, sl2, sl BooleanScanlineInterface,
	addSpan1Func, addSpan2Func AddSpanAAFunc,
	combineSpansFunc CombineSpansAAFunc,
) {
	sl.ResetSpans()

	num1 := sl1.NumSpans()
	num2 := sl2.NumSpans()

	var span1, span2 scanline.ScanlineIterator

	const (
		invalidB = 0x0FFFFFFF
		invalidE = invalidB - 1
	)

	// Initialize as invalid
	xb1 := invalidB
	xb2 := invalidB
	xe1 := invalidE
	xe2 := invalidE

	// Initialize span1 if there are spans
	if num1 > 0 {
		span1 = sl1.Begin()
		span1Data := span1.GetSpan()
		xb1 = span1Data.X
		len1 := span1Data.Len
		if len1 < 0 {
			len1 = -len1
		}
		xe1 = xb1 + len1 - 1
		num1--
	}

	// Initialize span2 if there are spans
	if num2 > 0 {
		span2 = sl2.Begin()
		span2Data := span2.GetSpan()
		xb2 = span2Data.X
		len2 := span2Data.Len
		if len2 < 0 {
			len2 = -len2
		}
		xe2 = xb2 + len2 - 1
		num2--
	}

	for {
		// Retrieve new span1 if invalid
		if num1 > 0 && xb1 > xe1 {
			num1--
			span1.Next()
			span1Data := span1.GetSpan()
			xb1 = span1Data.X
			len1 := span1Data.Len
			if len1 < 0 {
				len1 = -len1
			}
			xe1 = xb1 + len1 - 1
		}

		// Retrieve new span2 if invalid
		if num2 > 0 && xb2 > xe2 {
			num2--
			span2.Next()
			span2Data := span2.GetSpan()
			xb2 = span2Data.X
			len2 := span2Data.Len
			if len2 < 0 {
				len2 = -len2
			}
			xe2 = xb2 + len2 - 1
		}

		if xb1 > xe1 && xb2 > xe2 {
			break
		}

		// Calculate intersection
		xb := xb1
		xe := xe1
		if xb < xb2 {
			xb = xb2
		}
		if xe > xe2 {
			xe = xe2
		}
		length := xe - xb + 1

		if length > 0 {
			// Spans intersect - add beginning parts
			if xb1 < xb2 {
				span1Data := span1.GetSpan()
				span1Iter := &simpleIterator{
					x:      span1Data.X,
					len:    span1Data.Len,
					covers: span1Data.Covers,
				}
				addSpan1Func(span1Iter, xb1, uint(xb2-xb1), sl)
				xb1 = xb2
			} else if xb2 < xb1 {
				span2Data := span2.GetSpan()
				span2Iter := &simpleIterator{
					x:      span2Data.X,
					len:    span2Data.Len,
					covers: span2Data.Covers,
				}
				addSpan2Func(span2Iter, xb2, uint(xb1-xb2), sl)
				xb2 = xb1
			}

			// Add combination part
			span1Data := span1.GetSpan()
			span2Data := span2.GetSpan()
			span1Iter := &simpleIterator{
				x:      span1Data.X,
				len:    span1Data.Len,
				covers: span1Data.Covers,
			}
			span2Iter := &simpleIterator{
				x:      span2Data.X,
				len:    span2Data.Len,
				covers: span2Data.Covers,
			}
			combineSpansFunc(span1Iter, span2Iter, xb, uint(length), sl)

			// Invalidate processed spans
			if xe1 < xe2 {
				xb1 = invalidB
				xe1 = invalidE
				xb2 += length
			} else if xe2 < xe1 {
				xb2 = invalidB
				xe2 = invalidE
				xb1 += length
			} else {
				xb1 = invalidB
				xb2 = invalidB
				xe1 = invalidE
				xe2 = invalidE
			}
		} else {
			// Spans don't intersect
			if xb1 < xb2 {
				if xb1 <= xe1 {
					span1Data := span1.GetSpan()
					span1Iter := &simpleIterator{
						x:      span1Data.X,
						len:    span1Data.Len,
						covers: span1Data.Covers,
					}
					addSpan1Func(span1Iter, xb1, uint(xe1-xb1+1), sl)
				}
				xb1 = invalidB
				xe1 = invalidE
			} else {
				if xb2 <= xe2 {
					span2Data := span2.GetSpan()
					span2Iter := &simpleIterator{
						x:      span2Data.X,
						len:    span2Data.Len,
						covers: span2Data.Covers,
					}
					addSpan2Func(span2Iter, xb2, uint(xe2-xb2+1), sl)
				}
				xb2 = invalidB
				xe2 = invalidE
			}
		}
	}
}

// ==============================================================================
// Shape Operations
// ==============================================================================

// IntersectShapes intersects two shape generators
func IntersectShapes(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface,
	ren RendererInterface, combineSpansFunc CombineSpansAAFunc,
) {
	// Prepare the scanline generators
	if !sg1.RewindScanlines() {
		return
	}
	if !sg2.RewindScanlines() {
		return
	}

	// Get bounding boxes
	r1 := basics.Rect[int]{X1: sg1.MinX(), Y1: sg1.MinY(), X2: sg1.MaxX(), Y2: sg1.MaxY()}
	r2 := basics.Rect[int]{X1: sg2.MinX(), Y1: sg2.MinY(), X2: sg2.MaxX(), Y2: sg2.MaxY()}

	// Calculate intersection of bounding boxes
	_, valid := basics.IntersectRectangles(r1, r2)
	if !valid {
		return
	}

	// Reset scanlines
	sl.ResetSpans()
	sl1.ResetSpans()
	sl2.ResetSpans()

	if !sg1.SweepScanline(sl1) {
		return
	}
	if !sg2.SweepScanline(sl2) {
		return
	}

	ren.Prepare()

	// Main loop - synchronize scanlines with same Y coordinate
	for {
		for sl1.Y() < sl2.Y() {
			if !sg1.SweepScanline(sl1) {
				return
			}
		}
		for sl2.Y() < sl1.Y() {
			if !sg2.SweepScanline(sl2) {
				return
			}
		}

		if sl1.Y() == sl2.Y() {
			// Combine scanlines with same Y coordinate
			IntersectScanlines(sl1, sl2, sl, combineSpansFunc)
			if sl.NumSpans() > 0 {
				sl.Finalize(sl1.Y())
				ren.Render(sl)
			}
			if !sg1.SweepScanline(sl1) {
				return
			}
			if !sg2.SweepScanline(sl2) {
				return
			}
		}
	}
}

// UniteShapes unites two shape generators
func UniteShapes(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface,
	ren RendererInterface, addSpan1Func, addSpan2Func AddSpanAAFunc,
	combineSpansFunc CombineSpansAAFunc,
) {
	// Prepare scanline generators
	flag1 := sg1.RewindScanlines()
	flag2 := sg2.RewindScanlines()
	if !flag1 && !flag2 {
		return
	}

	// Get bounding boxes
	r1 := basics.Rect[int]{X1: sg1.MinX(), Y1: sg1.MinY(), X2: sg1.MaxX(), Y2: sg1.MaxY()}
	r2 := basics.Rect[int]{X1: sg2.MinX(), Y1: sg2.MinY(), X2: sg2.MaxX(), Y2: sg2.MaxY()}

	// Calculate union of bounding boxes (not used in the actual calculation)
	if flag1 && flag2 {
		_ = basics.Rect[int]{
			X1: basics.IMin(r1.X1, r2.X1),
			Y1: basics.IMin(r1.Y1, r2.Y1),
			X2: basics.IMax(r1.X2, r2.X2),
			Y2: basics.IMax(r1.Y2, r2.Y2),
		}
	}

	ren.Prepare()

	// Reset scanlines
	sl.ResetSpans()
	if flag1 {
		sl1.ResetSpans()
		flag1 = sg1.SweepScanline(sl1)
	}
	if flag2 {
		sl2.ResetSpans()
		flag2 = sg2.SweepScanline(sl2)
	}

	// Main loop
	for flag1 || flag2 {
		if flag1 && flag2 {
			if sl1.Y() == sl2.Y() {
				// Same Y coordinate - unite scanlines
				UniteScanlines(sl1, sl2, sl, addSpan1Func, addSpan2Func, combineSpansFunc)
				if sl.NumSpans() > 0 {
					sl.Finalize(sl1.Y())
					ren.Render(sl)
				}
				flag1 = sg1.SweepScanline(sl1)
				flag2 = sg2.SweepScanline(sl2)
			} else if sl1.Y() < sl2.Y() {
				AddSpansAndRender(sl1, sl, ren, addSpan1Func)
				flag1 = sg1.SweepScanline(sl1)
			} else {
				AddSpansAndRender(sl2, sl, ren, addSpan2Func)
				flag2 = sg2.SweepScanline(sl2)
			}
		} else {
			if flag1 {
				AddSpansAndRender(sl1, sl, ren, addSpan1Func)
				flag1 = sg1.SweepScanline(sl1)
			}
			if flag2 {
				AddSpansAndRender(sl2, sl, ren, addSpan2Func)
				flag2 = sg2.SweepScanline(sl2)
			}
		}
	}
}

// SubtractShapes subtracts shapes "sg1-sg2"
func SubtractShapes(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface,
	ren RendererInterface, addSpan1Func AddSpanAAFunc, combineSpansFunc CombineSpansAAFunc,
) {
	// Prepare scanline generators - sg1 is master, sg2 is slave
	if !sg1.RewindScanlines() {
		return
	}
	flag2 := sg2.RewindScanlines()

	// Reset scanlines
	sl.ResetSpans()
	sl1.ResetSpans()
	sl2.ResetSpans()

	if !sg1.SweepScanline(sl1) {
		return
	}
	if flag2 {
		flag2 = sg2.SweepScanline(sl2)
	}

	ren.Prepare()

	// Fake span2 processor
	addSpan2 := AddSpanEmpty

	flag1 := true
	for flag1 {
		// Synchronize slave with master
		for flag2 && sl2.Y() < sl1.Y() {
			flag2 = sg2.SweepScanline(sl2)
		}

		if flag2 && sl2.Y() == sl1.Y() {
			// Same Y coordinate - combine scanlines
			UniteScanlines(sl1, sl2, sl, addSpan1Func, addSpan2, combineSpansFunc)
			if sl.NumSpans() > 0 {
				sl.Finalize(sl1.Y())
				ren.Render(sl)
			}
		} else {
			AddSpansAndRender(sl1, sl, ren, addSpan1Func)
		}

		// Advance master
		flag1 = sg1.SweepScanline(sl1)
	}
}

// ==============================================================================
// Convenience Functions for Different Boolean Operations
// ==============================================================================

// IntersectShapesAA intersects two anti-aliased scanline shapes
func IntersectShapesAA(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	IntersectShapes(sg1, sg2, sl1, sl2, sl, ren, IntersectSpansAA)
}

// IntersectShapesBin intersects two binary scanline shapes
func IntersectShapesBin(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	IntersectShapes(sg1, sg2, sl1, sl2, sl, ren, CombineSpansBin)
}

// UniteShapesAA unites two anti-aliased scanline shapes
func UniteShapesAA(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	UniteShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanAA, AddSpanAA, UniteSpansAA)
}

// UniteShapesBin unites two binary scanline shapes
func UniteShapesBin(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	UniteShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanBin, AddSpanBin, CombineSpansBin)
}

// XorShapesAA applies XOR to two anti-aliased scanline shapes using linear formula
func XorShapesAA(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	formula := XorFormulaLinear{}
	combineFunc := func(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
		XorSpansAA(formula, span1, span2, x, length, sl)
	}
	UniteShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanAA, AddSpanAA, combineFunc)
}

// XorShapesSaddleAA applies XOR to two anti-aliased scanline shapes using saddle formula
func XorShapesSaddleAA(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	formula := XorFormulaSaddle{}
	combineFunc := func(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
		XorSpansAA(formula, span1, span2, x, length, sl)
	}
	UniteShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanAA, AddSpanAA, combineFunc)
}

// XorShapesAbsDiffAA applies XOR to two anti-aliased scanline shapes using absolute difference formula
func XorShapesAbsDiffAA(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	formula := XorFormulaAbsDiff{}
	combineFunc := func(span1, span2 IteratorInterface, x int, length uint, sl BooleanScanlineInterface) {
		XorSpansAA(formula, span1, span2, x, length, sl)
	}
	UniteShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanAA, AddSpanAA, combineFunc)
}

// XorShapesBin applies XOR to two binary scanline shapes
func XorShapesBin(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	UniteShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanBin, AddSpanBin, CombineSpansEmpty)
}

// SubtractShapesAA subtracts shapes "sg1-sg2" with anti-aliasing
func SubtractShapesAA(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	SubtractShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanAA, SubtractSpansAA)
}

// SubtractShapesBin subtracts binary shapes "sg1-sg2" without anti-aliasing
func SubtractShapesBin(sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	SubtractShapes(sg1, sg2, sl1, sl2, sl, ren, AddSpanBin, CombineSpansEmpty)
}

// ==============================================================================
// Boolean Operation Enumeration and Dispatch Functions
// ==============================================================================

// BoolOp represents the type of boolean operation to perform
type BoolOp int

const (
	BoolOr         BoolOp = iota // Union operation
	BoolAnd                      // Intersection operation
	BoolXor                      // XOR operation (linear formula)
	BoolXorSaddle                // XOR operation (saddle formula)
	BoolXorAbsDiff               // XOR operation (absolute difference formula)
	BoolAMinusB                  // A - B subtraction
	BoolBMinusA                  // B - A subtraction
)

// String returns the string representation of the boolean operation
func (op BoolOp) String() string {
	switch op {
	case BoolOr:
		return "Union"
	case BoolAnd:
		return "Intersection"
	case BoolXor:
		return "XOR (Linear)"
	case BoolXorSaddle:
		return "XOR (Saddle)"
	case BoolXorAbsDiff:
		return "XOR (Absolute Difference)"
	case BoolAMinusB:
		return "A - B"
	case BoolBMinusA:
		return "B - A"
	default:
		return "Unknown"
	}
}

// CombineShapesAA combines two anti-aliased scanline shapes using the specified operation
func CombineShapesAA(op BoolOp, sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	switch op {
	case BoolOr:
		UniteShapesAA(sg1, sg2, sl1, sl2, sl, ren)
	case BoolAnd:
		IntersectShapesAA(sg1, sg2, sl1, sl2, sl, ren)
	case BoolXor:
		XorShapesAA(sg1, sg2, sl1, sl2, sl, ren)
	case BoolXorSaddle:
		XorShapesSaddleAA(sg1, sg2, sl1, sl2, sl, ren)
	case BoolXorAbsDiff:
		XorShapesAbsDiffAA(sg1, sg2, sl1, sl2, sl, ren)
	case BoolAMinusB:
		SubtractShapesAA(sg1, sg2, sl1, sl2, sl, ren)
	case BoolBMinusA:
		SubtractShapesAA(sg2, sg1, sl2, sl1, sl, ren)
	}
}

// CombineShapesBin combines two binary scanline shapes using the specified operation
func CombineShapesBin(op BoolOp, sg1, sg2 RasterizerInterface, sl1, sl2, sl BooleanScanlineInterface, ren RendererInterface) {
	switch op {
	case BoolOr:
		UniteShapesBin(sg1, sg2, sl1, sl2, sl, ren)
	case BoolAnd:
		IntersectShapesBin(sg1, sg2, sl1, sl2, sl, ren)
	case BoolXor, BoolXorSaddle, BoolXorAbsDiff: // All XOR variants are the same for binary
		XorShapesBin(sg1, sg2, sl1, sl2, sl, ren)
	case BoolAMinusB:
		SubtractShapesBin(sg1, sg2, sl1, sl2, sl, ren)
	case BoolBMinusA:
		SubtractShapesBin(sg2, sg1, sl2, sl1, sl, ren)
	}
}
