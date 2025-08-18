package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// PixFmtTransposer wraps another pixel format and transposes x/y coordinates
// This effectively rotates the coordinate system by 90 degrees
type PixFmtTransposer struct {
	pixfmt interface {
		Width() int
		Height() int
		GetPixel(x, y int) color.RGBA8[color.Linear]
		CopyPixel(x, y int, c color.RGBA8[color.Linear])
		BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u)
		CopyHline(x, y, length int, c color.RGBA8[color.Linear])
		CopyVline(x, y, length int, c color.RGBA8[color.Linear])
		BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u)
		BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u)
		BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
		BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
	}
}

// NewPixFmtTransposer creates a new transposed pixel format wrapper
func NewPixFmtTransposer(pixfmt interface {
	Width() int
	Height() int
	GetPixel(x, y int) color.RGBA8[color.Linear]
	CopyPixel(x, y int, c color.RGBA8[color.Linear])
	BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u)
	CopyHline(x, y, length int, c color.RGBA8[color.Linear])
	CopyVline(x, y, length int, c color.RGBA8[color.Linear])
	BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u)
	BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u)
	BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
	BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
},
) *PixFmtTransposer {
	return &PixFmtTransposer{
		pixfmt: pixfmt,
	}
}

// Attach attaches a new pixel format to the transposer
func (pt *PixFmtTransposer) Attach(pixfmt interface {
	Width() int
	Height() int
	GetPixel(x, y int) color.RGBA8[color.Linear]
	CopyPixel(x, y int, c color.RGBA8[color.Linear])
	BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u)
	CopyHline(x, y, length int, c color.RGBA8[color.Linear])
	CopyVline(x, y, length int, c color.RGBA8[color.Linear])
	BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u)
	BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u)
	BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
	BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
},
) {
	pt.pixfmt = pixfmt
}

// Width returns the height of the underlying format (transposed)
func (pt *PixFmtTransposer) Width() int {
	return pt.pixfmt.Height()
}

// Height returns the width of the underlying format (transposed)
func (pt *PixFmtTransposer) Height() int {
	return pt.pixfmt.Width()
}

// GetPixel returns the color at the transposed coordinates
func (pt *PixFmtTransposer) GetPixel(x, y int) color.RGBA8[color.Linear] {
	return pt.pixfmt.GetPixel(y, x)
}

// CopyPixel copies a pixel at the transposed coordinates
func (pt *PixFmtTransposer) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	pt.pixfmt.CopyPixel(y, x, c)
}

// BlendPixel blends a pixel at the transposed coordinates
func (pt *PixFmtTransposer) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	pt.pixfmt.BlendPixel(y, x, c, cover)
}

// CopyHline copies a horizontal line (becomes vertical line in underlying format)
func (pt *PixFmtTransposer) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	pt.pixfmt.CopyVline(y, x, length, c)
}

// CopyVline copies a vertical line (becomes horizontal line in underlying format)
func (pt *PixFmtTransposer) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	pt.pixfmt.CopyHline(y, x, length, c)
}

// BlendHline blends a horizontal line (becomes vertical line in underlying format)
func (pt *PixFmtTransposer) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	pt.pixfmt.BlendVline(y, x, length, c, cover)
}

// BlendVline blends a vertical line (becomes horizontal line in underlying format)
func (pt *PixFmtTransposer) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	pt.pixfmt.BlendHline(y, x, length, c, cover)
}

// BlendSolidHspan blends a horizontal span (becomes vertical span in underlying format)
func (pt *PixFmtTransposer) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	pt.pixfmt.BlendSolidVspan(y, x, length, c, covers)
}

// BlendSolidVspan blends a vertical span (becomes horizontal span in underlying format)
func (pt *PixFmtTransposer) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	pt.pixfmt.BlendSolidHspan(y, x, length, c, covers)
}

// BlendColorHspan blends a horizontal span with varying colors
func (pt *PixFmtTransposer) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if colorSpanner, ok := pt.pixfmt.(interface {
		BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u)
	}); ok {
		colorSpanner.BlendColorVspan(y, x, length, colors, covers, cover)
	} else {
		// Fallback: blend pixel by pixel
		for i := 0; i < length; i++ {
			if covers != nil && covers[i] > 0 {
				actualCover := basics.Int8u((uint32(covers[i]) * uint32(cover)) / 255)
				pt.BlendPixel(x+i, y, colors[i], actualCover)
			} else if covers == nil {
				pt.BlendPixel(x+i, y, colors[i], cover)
			}
		}
	}
}

// BlendColorVspan blends a vertical span with varying colors
func (pt *PixFmtTransposer) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if colorSpanner, ok := pt.pixfmt.(interface {
		BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u)
	}); ok {
		colorSpanner.BlendColorHspan(y, x, length, colors, covers, cover)
	} else {
		// Fallback: blend pixel by pixel
		for i := 0; i < length; i++ {
			if covers != nil && covers[i] > 0 {
				actualCover := basics.Int8u((uint32(covers[i]) * uint32(cover)) / 255)
				pt.BlendPixel(x, y+i, colors[i], actualCover)
			} else if covers == nil {
				pt.BlendPixel(x, y+i, colors[i], cover)
			}
		}
	}
}

// CopyFrom copies from another rendering buffer
func (pt *PixFmtTransposer) CopyFrom(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int,
) {
	if copier, ok := pt.pixfmt.(interface {
		CopyFrom(src interface {
			RowData(y int) []basics.Int8u
			Width() int
			Height() int
		}, xdst, ydst, xsrc, ysrc, length int)
	}); ok {
		// Transpose coordinates for underlying format
		copier.CopyFrom(src, ydst, xdst, ysrc, xsrc, length)
	} else {
		// Fallback: copy pixel by pixel
		for i := 0; i < length; i++ {
			if xsrc+i >= 0 && xsrc+i < src.Width() && ysrc >= 0 && ysrc < src.Height() {
				// This is a simplified fallback - a real implementation would need
				// to handle the specific pixel format of the source
				pt.CopyPixel(xdst+i, ydst, color.RGBA8[color.Linear]{})
			}
		}
	}
}

// BlendFrom blends from another pixel format
func (pt *PixFmtTransposer) BlendFrom(src interface {
	GetPixel(x, y int) color.RGBA8[color.Linear]
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	if blender, ok := pt.pixfmt.(interface {
		BlendFrom(src interface {
			GetPixel(x, y int) color.RGBA8[color.Linear]
			Width() int
			Height() int
		}, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u)
	}); ok {
		// Transpose coordinates for underlying format
		blender.BlendFrom(src, ydst, xdst, ysrc, xsrc, length, cover)
	} else {
		// Fallback: blend pixel by pixel
		for i := 0; i < length; i++ {
			if xsrc+i >= 0 && xsrc+i < src.Width() && ysrc >= 0 && ysrc < src.Height() {
				srcPixel := src.GetPixel(xsrc+i, ysrc)
				pt.BlendPixel(xdst+i, ydst, srcPixel, cover)
			}
		}
	}
}

// Example usage and testing helpers

// TransposeCoords is a utility function that transposes coordinates
func TransposeCoords(x, y, width, height int) (newX, newY, newWidth, newHeight int) {
	return y, x, height, width
}

// IsTransposed checks if coordinates have been transposed
func IsTransposed(originalWidth, originalHeight, currentWidth, currentHeight int) bool {
	return originalWidth == currentHeight && originalHeight == currentWidth
}
