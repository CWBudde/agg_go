package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// Packed RGB blenders for 16-bit formats
// These blenders work directly on packed pixel data for optimal performance

// BlenderRGB555 implements standard RGB555 blending
type BlenderRGB555 struct{}

// BlendPix blends an RGB pixel into an RGB555 packed pixel
func (bl BlenderRGB555) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixel555(*pixel)

		// Blend in 8-bit space
		nr := color.RGBA8Lerp(dr, r, blendAlpha)
		ng := color.RGBA8Lerp(dg, g, blendAlpha)
		nb := color.RGBA8Lerp(db, b, blendAlpha)

		// Repack
		*pixel = MakePixel555(nr, ng, nb)
	}
}

// MakePix creates an RGB555 pixel from 8-bit RGB values
func (bl BlenderRGB555) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixel555(r, g, b)
}

// BlenderRGB555Pre implements premultiplied RGB555 blending
type BlenderRGB555Pre struct{}

// BlendPix blends a premultiplied RGB pixel into an RGB555 packed pixel
func (bl BlenderRGB555Pre) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	cr := color.RGBA8MultCover(r, cover)
	cg := color.RGBA8MultCover(g, cover)
	cb := color.RGBA8MultCover(b, cover)
	ca := color.RGBA8MultCover(alpha, cover)

	// Extract current pixel values
	dr, dg, db := UnpackPixel555(*pixel)

	// Premultiplied blend in 8-bit space
	nr := color.RGBA8Prelerp(dr, cr, ca)
	ng := color.RGBA8Prelerp(dg, cg, ca)
	nb := color.RGBA8Prelerp(db, cb, ca)

	// Repack
	*pixel = MakePixel555(nr, ng, nb)
}

// MakePix creates an RGB555 pixel from 8-bit RGB values
func (bl BlenderRGB555Pre) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixel555(r, g, b)
}

// BlenderRGB555Gamma implements gamma-corrected RGB555 blending
type BlenderRGB555Gamma[G any] struct {
	gamma G
}

// NewBlenderRGB555Gamma creates a new gamma-corrected RGB555 blender
func NewBlenderRGB555Gamma[G any](gamma G) BlenderRGB555Gamma[G] {
	return BlenderRGB555Gamma[G]{gamma: gamma}
}

// BlendPix blends an RGB pixel with gamma correction into an RGB555 packed pixel
func (bl BlenderRGB555Gamma[G]) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixel555(*pixel)

		// Apply gamma correction if gamma interface is available
		if gamma, ok := interface{}(bl.gamma).(GammaCorrector); ok {
			// Convert to linear space
			dr = gamma.Dir(dr)
			dg = gamma.Dir(dg)
			db = gamma.Dir(db)
			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			// Blend in linear space
			nr := color.RGBA8Lerp(dr, sr, blendAlpha)
			ng := color.RGBA8Lerp(dg, sg, blendAlpha)
			nb := color.RGBA8Lerp(db, sb, blendAlpha)

			// Convert back to gamma space
			nr = gamma.Inv(nr)
			ng = gamma.Inv(ng)
			nb = gamma.Inv(nb)

			// Repack
			*pixel = MakePixel555(nr, ng, nb)
		} else {
			// Fallback to standard blending
			nr := color.RGBA8Lerp(dr, r, blendAlpha)
			ng := color.RGBA8Lerp(dg, g, blendAlpha)
			nb := color.RGBA8Lerp(db, b, blendAlpha)
			*pixel = MakePixel555(nr, ng, nb)
		}
	}
}

// MakePix creates an RGB555 pixel from 8-bit RGB values
func (bl BlenderRGB555Gamma[G]) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixel555(r, g, b)
}

// BlenderRGB565 implements standard RGB565 blending
type BlenderRGB565 struct{}

// BlendPix blends an RGB pixel into an RGB565 packed pixel
func (bl BlenderRGB565) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixel565(*pixel)

		// Blend in 8-bit space
		nr := color.RGBA8Lerp(dr, r, blendAlpha)
		ng := color.RGBA8Lerp(dg, g, blendAlpha)
		nb := color.RGBA8Lerp(db, b, blendAlpha)

		// Repack
		*pixel = MakePixel565(nr, ng, nb)
	}
}

// MakePix creates an RGB565 pixel from 8-bit RGB values
func (bl BlenderRGB565) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixel565(r, g, b)
}

// BlenderRGB565Pre implements premultiplied RGB565 blending
type BlenderRGB565Pre struct{}

// BlendPix blends a premultiplied RGB pixel into an RGB565 packed pixel
func (bl BlenderRGB565Pre) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	cr := color.RGBA8MultCover(r, cover)
	cg := color.RGBA8MultCover(g, cover)
	cb := color.RGBA8MultCover(b, cover)
	ca := color.RGBA8MultCover(alpha, cover)

	// Extract current pixel values
	dr, dg, db := UnpackPixel565(*pixel)

	// Premultiplied blend in 8-bit space
	nr := color.RGBA8Prelerp(dr, cr, ca)
	ng := color.RGBA8Prelerp(dg, cg, ca)
	nb := color.RGBA8Prelerp(db, cb, ca)

	// Repack
	*pixel = MakePixel565(nr, ng, nb)
}

// MakePix creates an RGB565 pixel from 8-bit RGB values
func (bl BlenderRGB565Pre) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixel565(r, g, b)
}

// BlenderRGB565Gamma implements gamma-corrected RGB565 blending
type BlenderRGB565Gamma[G any] struct {
	gamma G
}

// NewBlenderRGB565Gamma creates a new gamma-corrected RGB565 blender
func NewBlenderRGB565Gamma[G any](gamma G) BlenderRGB565Gamma[G] {
	return BlenderRGB565Gamma[G]{gamma: gamma}
}

// BlendPix blends an RGB pixel with gamma correction into an RGB565 packed pixel
func (bl BlenderRGB565Gamma[G]) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixel565(*pixel)

		// Apply gamma correction if gamma interface is available
		if gamma, ok := interface{}(bl.gamma).(GammaCorrector); ok {
			// Convert to linear space
			dr = gamma.Dir(dr)
			dg = gamma.Dir(dg)
			db = gamma.Dir(db)
			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			// Blend in linear space
			nr := color.RGBA8Lerp(dr, sr, blendAlpha)
			ng := color.RGBA8Lerp(dg, sg, blendAlpha)
			nb := color.RGBA8Lerp(db, sb, blendAlpha)

			// Convert back to gamma space
			nr = gamma.Inv(nr)
			ng = gamma.Inv(ng)
			nb = gamma.Inv(nb)

			// Repack
			*pixel = MakePixel565(nr, ng, nb)
		} else {
			// Fallback to standard blending
			nr := color.RGBA8Lerp(dr, r, blendAlpha)
			ng := color.RGBA8Lerp(dg, g, blendAlpha)
			nb := color.RGBA8Lerp(db, b, blendAlpha)
			*pixel = MakePixel565(nr, ng, nb)
		}
	}
}

// MakePix creates an RGB565 pixel from 8-bit RGB values
func (bl BlenderRGB565Gamma[G]) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixel565(r, g, b)
}

// BlenderBGR555 implements standard BGR555 blending
type BlenderBGR555 struct{}

// BlendPix blends an RGB pixel into a BGR555 packed pixel
func (bl BlenderBGR555) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixelBGR555(*pixel)

		// Blend in 8-bit space
		nr := color.RGBA8Lerp(dr, r, blendAlpha)
		ng := color.RGBA8Lerp(dg, g, blendAlpha)
		nb := color.RGBA8Lerp(db, b, blendAlpha)

		// Repack
		*pixel = MakePixelBGR555(nr, ng, nb)
	}
}

// MakePix creates a BGR555 pixel from 8-bit RGB values
func (bl BlenderBGR555) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixelBGR555(r, g, b)
}

// BlenderBGR565 implements standard BGR565 blending
type BlenderBGR565 struct{}

// BlendPix blends an RGB pixel into a BGR565 packed pixel
func (bl BlenderBGR565) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixelBGR565(*pixel)

		// Blend in 8-bit space
		nr := color.RGBA8Lerp(dr, r, blendAlpha)
		ng := color.RGBA8Lerp(dg, g, blendAlpha)
		nb := color.RGBA8Lerp(db, b, blendAlpha)

		// Repack
		*pixel = MakePixelBGR565(nr, ng, nb)
	}
}

// MakePix creates a BGR565 pixel from 8-bit RGB values
func (bl BlenderBGR565) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixelBGR565(r, g, b)
}

// Convenience function to get the correct color from a packed pixel
func MakeColorRGB555(pixel basics.Int16u) color.RGB8[color.Linear] {
	r, g, b := UnpackPixel555(pixel)
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

func MakeColorRGB565(pixel basics.Int16u) color.RGB8[color.Linear] {
	r, g, b := UnpackPixel565(pixel)
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

func MakeColorBGR555(pixel basics.Int16u) color.RGB8[color.Linear] {
	r, g, b := UnpackPixelBGR555(pixel)
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

func MakeColorBGR565(pixel basics.Int16u) color.RGB8[color.Linear] {
	r, g, b := UnpackPixelBGR565(pixel)
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

// BlenderBGR555Gamma implements gamma-corrected BGR555 blending
type BlenderBGR555Gamma[G any] struct {
	gamma G
}

// NewBlenderBGR555Gamma creates a new gamma-corrected BGR555 blender
func NewBlenderBGR555Gamma[G any](gamma G) BlenderBGR555Gamma[G] {
	return BlenderBGR555Gamma[G]{gamma: gamma}
}

// BlendPix blends an RGB pixel with gamma correction into a BGR555 packed pixel
func (bl BlenderBGR555Gamma[G]) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixelBGR555(*pixel)

		// Apply gamma correction if gamma interface is available
		if gamma, ok := interface{}(bl.gamma).(GammaCorrector); ok {
			// Convert to linear space for blending
			dr = gamma.Dir(dr)
			dg = gamma.Dir(dg)
			db = gamma.Dir(db)
			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			// Blend in linear space
			nr := color.RGBA8Lerp(dr, sr, blendAlpha)
			ng := color.RGBA8Lerp(dg, sg, blendAlpha)
			nb := color.RGBA8Lerp(db, sb, blendAlpha)

			// Convert back to gamma space
			nr = gamma.Inv(nr)
			ng = gamma.Inv(ng)
			nb = gamma.Inv(nb)

			// Repack
			*pixel = MakePixelBGR555(nr, ng, nb)
		} else {
			// Fallback to standard blend without gamma correction
			nr := color.RGBA8Lerp(dr, r, blendAlpha)
			ng := color.RGBA8Lerp(dg, g, blendAlpha)
			nb := color.RGBA8Lerp(db, b, blendAlpha)
			*pixel = MakePixelBGR555(nr, ng, nb)
		}
	}
}

// MakePix creates a BGR555 pixel from 8-bit RGB values
func (bl BlenderBGR555Gamma[G]) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixelBGR555(r, g, b)
}

// BlenderBGR565Gamma implements gamma-corrected BGR565 blending
type BlenderBGR565Gamma[G any] struct {
	gamma G
}

// NewBlenderBGR565Gamma creates a new gamma-corrected BGR565 blender
func NewBlenderBGR565Gamma[G any](gamma G) BlenderBGR565Gamma[G] {
	return BlenderBGR565Gamma[G]{gamma: gamma}
}

// BlendPix blends an RGB pixel with gamma correction into a BGR565 packed pixel
func (bl BlenderBGR565Gamma[G]) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		// Extract current pixel values
		dr, dg, db := UnpackPixelBGR565(*pixel)

		// Apply gamma correction if gamma interface is available
		if gamma, ok := interface{}(bl.gamma).(GammaCorrector); ok {
			// Convert to linear space for blending
			dr = gamma.Dir(dr)
			dg = gamma.Dir(dg)
			db = gamma.Dir(db)
			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			// Blend in linear space
			nr := color.RGBA8Lerp(dr, sr, blendAlpha)
			ng := color.RGBA8Lerp(dg, sg, blendAlpha)
			nb := color.RGBA8Lerp(db, sb, blendAlpha)

			// Convert back to gamma space
			nr = gamma.Inv(nr)
			ng = gamma.Inv(ng)
			nb = gamma.Inv(nb)

			// Repack
			*pixel = MakePixelBGR565(nr, ng, nb)
		} else {
			// Fallback to standard blend without gamma correction
			nr := color.RGBA8Lerp(dr, r, blendAlpha)
			ng := color.RGBA8Lerp(dg, g, blendAlpha)
			nb := color.RGBA8Lerp(db, b, blendAlpha)
			*pixel = MakePixelBGR565(nr, ng, nb)
		}
	}
}

// MakePix creates a BGR565 pixel from 8-bit RGB values
func (bl BlenderBGR565Gamma[G]) MakePix(r, g, b basics.Int8u) basics.Int16u {
	return MakePixelBGR565(r, g, b)
}
