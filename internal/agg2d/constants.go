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
	// ApproxScale is C++ g_approxScale (agg2d.cpp:28). The curve and stroke
	// converters get worldToScreen(1.0) * ApproxScale whenever the world
	// transform is scaled or replaced (scale, affine, parallelogram, viewport,
	// transformations); until then they keep AGG's default scale of 1.0.
	ApproxScale = 2.0
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

	// ImageFilterBlackman144 is C++ Agg2D::Blackman144, i.e.
	// image_filter_blackman144 = image_filter_blackman with radius 6
	// (agg_image_filters.h:435). ImageFilterBlackman (radius 4) is a Go
	// extension and is not the same filter.
	ImageFilterBlackman144 ImageFilter = 17
)

// Additional ImageResample constants
const (
	ResampleAlways    ImageResample = 1
	ResampleOnZoomOut ImageResample = 2
)

// Affine image resample policy controls how affine image transforms pick their
// sampling generator when ImageResample is left at NoResample.
const (
	AffineImageResampleAgg2D AffineImageResamplePolicy = iota
	AffineImageResamplePreferFiltered
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
	Hamming                 = ImageFilterHamming
	Hermite                 = ImageFilterHermite
	Quadric                 = ImageFilterQuadric
	Bicubic                 = ImageFilterBicubic
	Catrom                  = ImageFilterCatrom
	Spline16                = ImageFilterSpline16
	Spline36                = ImageFilterSpline36
	Blackman                = ImageFilterBlackman
	Blackman144             = ImageFilterBlackman144
	Kaiser                  = ImageFilterKaiser
	Gaussian                = ImageFilterGaussian
	Bessel                  = ImageFilterBessel
	Mitchell                = ImageFilterMitchell
	Sinc                    = ImageFilterSinc
	Lanczos                 = ImageFilterLanczos
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
