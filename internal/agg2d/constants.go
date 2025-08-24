// Package agg2d constants for AGG2D high-level interface.
// This file contains additional constant definitions that extend the C++ AGG2D interface.
package agg2d

import "math"

// Mathematical constants
const (
	Pi      = math.Pi
	Pi2     = math.Pi * 2
	PiHalf  = math.Pi / 2
	Deg2Rad = math.Pi / 180.0
	Rad2Deg = 180.0 / math.Pi
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
	ImageFilterBicubic  ImageFilter = 1
	ImageFilterSpline16 ImageFilter = 2
	ImageFilterSpline36 ImageFilter = 3
	ImageFilterHanning  ImageFilter = 4
	ImageFilterHamming  ImageFilter = 5
	ImageFilterHermite  ImageFilter = 6
	ImageFilterKaiser   ImageFilter = 7
	ImageFilterQuadric  ImageFilter = 8
	ImageFilterCatrom   ImageFilter = 9
	ImageFilterGaussian ImageFilter = 10
	ImageFilterBessel   ImageFilter = 11
	ImageFilterMitchell ImageFilter = 12
	ImageFilterSinc     ImageFilter = 13
	ImageFilterLanczos  ImageFilter = 14
	ImageFilterBlackman ImageFilter = 15
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
	NoFilter    = 0
	Bilinear    = ImageFilterBilinear
	Hanning     = ImageFilterHanning
	Hermite     = ImageFilterHermite
	Quadric     = ImageFilterQuadric
	Bicubic     = ImageFilterBicubic
	Catrom      = ImageFilterCatrom
	Spline16    = ImageFilterSpline16
	Spline36    = ImageFilterSpline36
	Blackman144 = ImageFilterBlackman
)
