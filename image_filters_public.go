package agg

import "github.com/cwbudde/agg_go/internal/agg2d"

// Canonical image-filter names. These constants supplement the historical
// Filter* aliases without changing their values or behaviour.
const (
	ImageFilterHamming  ImageFilter = agg2d.ImageFilterHamming
	ImageFilterKaiser   ImageFilter = agg2d.ImageFilterKaiser
	ImageFilterGaussian ImageFilter = agg2d.ImageFilterGaussian
	ImageFilterBessel   ImageFilter = agg2d.ImageFilterBessel
	ImageFilterMitchell ImageFilter = agg2d.ImageFilterMitchell
	ImageFilterSinc     ImageFilter = agg2d.ImageFilterSinc
	ImageFilterLanczos  ImageFilter = agg2d.ImageFilterLanczos
)

// Short canonical aliases matching the original AGG filter names.
const (
	Hamming  = ImageFilterHamming
	Kaiser   = ImageFilterKaiser
	Gaussian = ImageFilterGaussian
	Bessel   = ImageFilterBessel
	Mitchell = ImageFilterMitchell
	Sinc     = ImageFilterSinc
	Lanczos  = ImageFilterLanczos
)
