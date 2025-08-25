package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// BlenderRGBA16 implements standard RGBA blending for 16-bit values with byte order support
type BlenderRGBA16[CS any, O any] struct{}

// BlendPix blends a 16-bit RGBA pixel respecting byte order
func (bl BlenderRGBA16[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGBA16MultCover(a, cover)
	if alpha > 0 {
		order := GetColorOrder[O]()

		// Read destination as 16-bit values based on order
		dr := basics.Int16u(dst[order.R*2]) | (basics.Int16u(dst[order.R*2+1]) << 8)
		dg := basics.Int16u(dst[order.G*2]) | (basics.Int16u(dst[order.G*2+1]) << 8)
		db := basics.Int16u(dst[order.B*2]) | (basics.Int16u(dst[order.B*2+1]) << 8)
		da := basics.Int16u(dst[order.A*2]) | (basics.Int16u(dst[order.A*2+1]) << 8)

		// Blend
		dr = color.RGBA16Lerp(dr, r, alpha)
		dg = color.RGBA16Lerp(dg, g, alpha)
		db = color.RGBA16Lerp(db, b, alpha)
		da = color.RGBA16Prelerp(da, alpha, alpha)

		// Write back as little-endian bytes based on order
		dst[order.R*2] = basics.Int8u(dr)
		dst[order.R*2+1] = basics.Int8u(dr >> 8)
		dst[order.G*2] = basics.Int8u(dg)
		dst[order.G*2+1] = basics.Int8u(dg >> 8)
		dst[order.B*2] = basics.Int8u(db)
		dst[order.B*2+1] = basics.Int8u(db >> 8)
		dst[order.A*2] = basics.Int8u(da)
		dst[order.A*2+1] = basics.Int8u(da >> 8)
	}
}

// BlenderRGBA16Pre implements premultiplied RGBA blending for 16-bit values with byte order support
type BlenderRGBA16Pre[CS any, O any] struct{}

// BlendPix blends a premultiplied 16-bit RGBA pixel respecting byte order
func (bl BlenderRGBA16Pre[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	cr := color.RGBA16MultCover(r, cover)
	cg := color.RGBA16MultCover(g, cover)
	cb := color.RGBA16MultCover(b, cover)
	ca := color.RGBA16MultCover(a, cover)

	order := GetColorOrder[O]()

	// Read destination as 16-bit values based on order
	dr := basics.Int16u(dst[order.R*2]) | (basics.Int16u(dst[order.R*2+1]) << 8)
	dg := basics.Int16u(dst[order.G*2]) | (basics.Int16u(dst[order.G*2+1]) << 8)
	db := basics.Int16u(dst[order.B*2]) | (basics.Int16u(dst[order.B*2+1]) << 8)
	da := basics.Int16u(dst[order.A*2]) | (basics.Int16u(dst[order.A*2+1]) << 8)

	// Blend
	dr = color.RGBA16Prelerp(dr, cr, ca)
	dg = color.RGBA16Prelerp(dg, cg, ca)
	db = color.RGBA16Prelerp(db, cb, ca)
	da = color.RGBA16Prelerp(da, ca, ca)

	// Write back as little-endian bytes based on order
	dst[order.R*2] = basics.Int8u(dr)
	dst[order.R*2+1] = basics.Int8u(dr >> 8)
	dst[order.G*2] = basics.Int8u(dg)
	dst[order.G*2+1] = basics.Int8u(dg >> 8)
	dst[order.B*2] = basics.Int8u(db)
	dst[order.B*2+1] = basics.Int8u(db >> 8)
	dst[order.A*2] = basics.Int8u(da)
	dst[order.A*2+1] = basics.Int8u(da >> 8)
}

// BlenderRGBA16Plain implements plain RGBA blending for 16-bit values with byte order support
type BlenderRGBA16Plain[CS any, O any] struct{}

// BlendPix blends a plain 16-bit RGBA pixel respecting byte order
func (bl BlenderRGBA16Plain[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGBA16MultCover(a, cover)
	if alpha > 0 {
		order := GetColorOrder[O]()

		// Read destination as 16-bit values based on order
		dr := basics.Int16u(dst[order.R*2]) | (basics.Int16u(dst[order.R*2+1]) << 8)
		dg := basics.Int16u(dst[order.G*2]) | (basics.Int16u(dst[order.G*2+1]) << 8)
		db := basics.Int16u(dst[order.B*2]) | (basics.Int16u(dst[order.B*2+1]) << 8)
		da := basics.Int16u(dst[order.A*2]) | (basics.Int16u(dst[order.A*2+1]) << 8)

		// Premultiply destination for calculation
		pdr := color.RGBA16Multiply(dr, da)
		pdg := color.RGBA16Multiply(dg, da)
		pdb := color.RGBA16Multiply(db, da)

		// Blend in premultiplied space
		pdr = color.RGBA16Lerp(pdr, r, alpha)
		pdg = color.RGBA16Lerp(pdg, g, alpha)
		pdb = color.RGBA16Lerp(pdb, b, alpha)
		da = color.RGBA16Prelerp(da, alpha, alpha)

		// Demultiply result back to plain space
		if da > 0 {
			dr = basics.Int16u((uint32(pdr)*65535 + uint32(da)/2) / uint32(da))
			dg = basics.Int16u((uint32(pdg)*65535 + uint32(da)/2) / uint32(da))
			db = basics.Int16u((uint32(pdb)*65535 + uint32(da)/2) / uint32(da))
		} else {
			dr, dg, db = 0, 0, 0
		}

		// Write back as little-endian bytes based on order
		dst[order.R*2] = basics.Int8u(dr)
		dst[order.R*2+1] = basics.Int8u(dr >> 8)
		dst[order.G*2] = basics.Int8u(dg)
		dst[order.G*2+1] = basics.Int8u(dg >> 8)
		dst[order.B*2] = basics.Int8u(db)
		dst[order.B*2+1] = basics.Int8u(db >> 8)
		dst[order.A*2] = basics.Int8u(da)
		dst[order.A*2+1] = basics.Int8u(da >> 8)
	}
}

// Helper function to blend a single RGBA16 pixel
func BlendRGBA16Pixel[B interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u)
}](dst []basics.Int8u, src color.RGBA16[color.Linear], cover basics.Int16u, blender B) {
	if !src.IsTransparent() {
		blender.BlendPix(dst, src.R, src.G, src.B, src.A, cover)
	}
}

// Concrete 16-bit blender types for different byte orders
type (
	// Standard RGBA16 blenders with Linear color space
	BlenderRGBA16Linear = BlenderRGBA16[color.Linear, RGBAOrder]
	BlenderARGB16Linear = BlenderRGBA16[color.Linear, ARGBOrder]
	BlenderABGR16Linear = BlenderRGBA16[color.Linear, ABGROrder]
	BlenderBGRA16Linear = BlenderRGBA16[color.Linear, BGRAOrder]

	// Premultiplied RGBA16 blenders
	BlenderRGBA16PreLinear = BlenderRGBA16Pre[color.Linear, RGBAOrder]
	BlenderARGB16PreLinear = BlenderRGBA16Pre[color.Linear, ARGBOrder]
	BlenderABGR16PreLinear = BlenderRGBA16Pre[color.Linear, ABGROrder]
	BlenderBGRA16PreLinear = BlenderRGBA16Pre[color.Linear, BGRAOrder]

	// Plain RGBA16 blenders
	BlenderRGBA16PlainLinear = BlenderRGBA16Plain[color.Linear, RGBAOrder]
	BlenderARGB16PlainLinear = BlenderRGBA16Plain[color.Linear, ARGBOrder]
	BlenderABGR16PlainLinear = BlenderRGBA16Plain[color.Linear, ABGROrder]
	BlenderBGRA16PlainLinear = BlenderRGBA16Plain[color.Linear, BGRAOrder]
)
