// Package agg2d constants for AGG2D high-level interface.
// This file contains additional constant definitions that extend the C++ AGG2D interface.
package agg2d

import "math"

// Mathematical constants
const (
	Pi            = math.Pi
	Pi2           = math.Pi * 2
	PiHalf        = math.Pi / 2
	Deg2RadFactor = math.Pi / 180.0
	Rad2DegFactor = 180.0 / math.Pi
)

// Curve approximation constants
const (
	// ApproxScale is the global approximation scale factor for curves.
	// This matches g_approxScale from the original C++ AGG library.
	ApproxScale = 1.0
)

// Additional LineCap constants not yet defined in agg2d.go
const (
	CapSquare LineCap = 2
)

// Additional LineJoin constants not yet defined in agg2d.go
const (
	JoinBevel LineJoin = 2
)

// Additional ImageFilter constants
const (
	ImageFilterHanning  ImageFilter = 2
	ImageFilterHermite  ImageFilter = 3
	ImageFilterQuadric  ImageFilter = 4
	ImageFilterBicubic  ImageFilter = 5
	ImageFilterCatrom   ImageFilter = 6
	ImageFilterSpline16 ImageFilter = 7
	ImageFilterSpline36 ImageFilter = 8
	ImageFilterBlackman ImageFilter = 9
	ImageFilterHamming  ImageFilter = 10
	ImageFilterKaiser   ImageFilter = 11
	ImageFilterGaussian ImageFilter = 12
	ImageFilterBessel   ImageFilter = 13
	ImageFilterMitchell ImageFilter = 14
	ImageFilterSinc     ImageFilter = 15
	ImageFilterLanczos  ImageFilter = 16
)

// Additional ImageResample constants
const (
	ResampleAlways    ImageResample = 1
	ResampleOnZoomOut ImageResample = 2
)

// Additional TextAlignment constants
const (
	AlignRight  TextAlignment = 1
	AlignCenter TextAlignment = 2
	AlignTop    TextAlignment = AlignRight
)

// FontCacheType constants
const (
	VectorFontCache FontCacheType = 1
)

// RectD represents a double-precision rectangle
type RectD struct {
	X1, Y1, X2, Y2 float64
}

// Image filter constants for testing
const (
	NoFilter    ImageFilter = 0
	Bilinear                = ImageFilterBilinear
	Hanning                 = ImageFilterHanning
	Hermite                 = ImageFilterHermite
	Quadric                 = ImageFilterQuadric
	Bicubic                 = ImageFilterBicubic
	Catrom                  = ImageFilterCatrom
	Spline16                = ImageFilterSpline16
	Spline36                = ImageFilterSpline36
	Blackman144             = ImageFilterBlackman
)

// DrawPathFlag represents different path drawing modes
type DrawPathFlag int

const (
	FillOnly          DrawPathFlag = iota // Fill the path only
	StrokeOnly                            // Stroke the path only
	FillAndStroke                         // Both fill and stroke the path
	FillWithLineColor                     // Fill the path using line color
)

// Direction represents path direction
type Direction int

const (
	CW  Direction = iota // Clockwise
	CCW                  // Counter-clockwise
)
