package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// AlphaMaskInterface defines the interface for alpha masks
type AlphaMaskInterface interface {
	// Width returns the width of the alpha mask
	Width() int
	// Height returns the height of the alpha mask
	Height() int
	// Pixel returns the alpha value at the given coordinates
	Pixel(x, y int) basics.Int8u
	// CombinePixel combines the given coverage with the mask's alpha at the coordinates
	CombinePixel(x, y int, cover basics.Int8u) basics.Int8u
	// FillHspan fills a horizontal span with mask alpha values
	FillHspan(x, y int, dst []basics.Int8u, length int)
	// CombineHspan combines coverage values with mask alpha for a horizontal span
	CombineHspan(x, y int, dst []basics.Int8u, length int)
	// FillVspan fills a vertical span with mask alpha values
	FillVspan(x, y int, dst []basics.Int8u, length int)
	// CombineVspan combines coverage values with mask alpha for a vertical span
	CombineVspan(x, y int, dst []basics.Int8u, length int)
}

// PixFmtAMaskAdaptor wraps a pixel format and applies an alpha mask to all operations
type PixFmtAMaskAdaptor struct {
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
	amask AlphaMaskInterface
	span  []basics.Int8u // Reusable span buffer
}

// SpanExtraTail is the extra space allocated for span buffers to avoid frequent reallocations
const SpanExtraTail = 256

// NewPixFmtAMaskAdaptor creates a new alpha mask adaptor
func NewPixFmtAMaskAdaptor(pixfmt interface {
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
}, amask AlphaMaskInterface) *PixFmtAMaskAdaptor {
	return &PixFmtAMaskAdaptor{
		pixfmt: pixfmt,
		amask:  amask,
		span:   make([]basics.Int8u, 0),
	}
}

// AttachPixfmt attaches a new pixel format
func (pa *PixFmtAMaskAdaptor) AttachPixfmt(pixfmt interface {
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
}) {
	pa.pixfmt = pixfmt
}

// AttachAlphaMask attaches a new alpha mask
func (pa *PixFmtAMaskAdaptor) AttachAlphaMask(amask AlphaMaskInterface) {
	pa.amask = amask
}

// reallocSpan reallocates the span buffer if needed
func (pa *PixFmtAMaskAdaptor) reallocSpan(length int) {
	if length > len(pa.span) {
		pa.span = make([]basics.Int8u, length+SpanExtraTail)
	}
}

// initSpan initializes the span buffer with full coverage
func (pa *PixFmtAMaskAdaptor) initSpan(length int) {
	pa.reallocSpan(length)
	for i := 0; i < length; i++ {
		pa.span[i] = 255 // Full coverage
	}
}

// initSpanWithCovers initializes the span buffer with given coverage values
func (pa *PixFmtAMaskAdaptor) initSpanWithCovers(length int, covers []basics.Int8u) {
	pa.reallocSpan(length)
	copy(pa.span[:length], covers[:length])
}

// Width returns the width of the pixel format
func (pa *PixFmtAMaskAdaptor) Width() int {
	return pa.pixfmt.Width()
}

// Height returns the height of the pixel format
func (pa *PixFmtAMaskAdaptor) Height() int {
	return pa.pixfmt.Height()
}

// GetPixel returns the color at the given coordinates (unaffected by mask)
func (pa *PixFmtAMaskAdaptor) GetPixel(x, y int) color.RGBA8[color.Linear] {
	return pa.pixfmt.GetPixel(x, y)
}

// CopyPixel copies a pixel with mask applied as coverage
func (pa *PixFmtAMaskAdaptor) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	cover := pa.amask.Pixel(x, y)
	pa.pixfmt.BlendPixel(x, y, c, cover)
}

// BlendPixel blends a pixel with mask combined with coverage
func (pa *PixFmtAMaskAdaptor) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	combinedCover := pa.amask.CombinePixel(x, y, cover)
	pa.pixfmt.BlendPixel(x, y, c, combinedCover)
}

// CopyHline copies a horizontal line with mask applied
func (pa *PixFmtAMaskAdaptor) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	pa.reallocSpan(length)
	pa.amask.FillHspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidHspan(x, y, length, c, pa.span[:length])
}

// BlendHline blends a horizontal line with mask combined with coverage
func (pa *PixFmtAMaskAdaptor) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	pa.initSpan(length)
	pa.amask.CombineHspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidHspan(x, y, length, c, pa.span[:length])
}

// CopyVline copies a vertical line with mask applied
func (pa *PixFmtAMaskAdaptor) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	pa.reallocSpan(length)
	pa.amask.FillVspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidVspan(x, y, length, c, pa.span[:length])
}

// BlendVline blends a vertical line with mask combined with coverage
func (pa *PixFmtAMaskAdaptor) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	pa.initSpan(length)
	pa.amask.CombineVspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidVspan(x, y, length, c, pa.span[:length])
}

// BlendSolidHspan blends a horizontal span with mask applied
func (pa *PixFmtAMaskAdaptor) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
		pa.amask.CombineHspan(x, y, pa.span[:length], length)
		pa.pixfmt.BlendSolidHspan(x, y, length, c, pa.span[:length])
	} else {
		pa.CopyHline(x, y, length, c)
	}
}

// BlendSolidVspan blends a vertical span with mask applied
func (pa *PixFmtAMaskAdaptor) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
		pa.amask.CombineVspan(x, y, pa.span[:length], length)
		pa.pixfmt.BlendSolidVspan(x, y, length, c, pa.span[:length])
	} else {
		pa.CopyVline(x, y, length, c)
	}
}

// CopyFrom copies from another rendering buffer (not affected by mask for source)
func (pa *PixFmtAMaskAdaptor) CopyFrom(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int) {
	if copier, ok := pa.pixfmt.(interface {
		CopyFrom(src interface {
			RowData(y int) []basics.Int8u
			Width() int
			Height() int
		}, xdst, ydst, xsrc, ysrc, length int)
	}); ok {
		copier.CopyFrom(src, xdst, ydst, xsrc, ysrc, length)
	}
}

// BlendFrom blends from another pixel format with mask applied
func (pa *PixFmtAMaskAdaptor) BlendFrom(src interface {
	GetPixel(x, y int) color.RGBA8[color.Linear]
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u) {
	// Blend pixel by pixel with mask applied
	for i := 0; i < length; i++ {
		if xsrc+i >= 0 && xsrc+i < src.Width() && ysrc >= 0 && ysrc < src.Height() {
			srcPixel := src.GetPixel(xsrc+i, ysrc)
			combinedCover := pa.amask.CombinePixel(xdst+i, ydst, cover)
			pa.pixfmt.BlendPixel(xdst+i, ydst, srcPixel, combinedCover)
		}
	}
}

// BlendColorHspan blends a horizontal span with varying colors and mask applied
func (pa *PixFmtAMaskAdaptor) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
	} else {
		pa.initSpan(length)
		for i := 0; i < length; i++ {
			pa.span[i] = cover
		}
	}

	// Apply mask to the span
	pa.amask.CombineHspan(x, y, pa.span[:length], length)

	// Blend each pixel individually since we have varying colors
	for i := 0; i < length; i++ {
		if pa.span[i] > 0 {
			pa.pixfmt.BlendPixel(x+i, y, colors[i], pa.span[i])
		}
	}
}

// BlendColorVspan blends a vertical span with varying colors and mask applied
func (pa *PixFmtAMaskAdaptor) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
	} else {
		pa.initSpan(length)
		for i := 0; i < length; i++ {
			pa.span[i] = cover
		}
	}

	// Apply mask to the span
	pa.amask.CombineVspan(x, y, pa.span[:length], length)

	// Blend each pixel individually since we have varying colors
	for i := 0; i < length; i++ {
		if pa.span[i] > 0 {
			pa.pixfmt.BlendPixel(x, y+i, colors[i], pa.span[i])
		}
	}
}

// SimpleAlphaMask provides a simple implementation of AlphaMaskInterface
// This can be used for testing or as a base for more complex masks
type SimpleAlphaMask struct {
	width  int
	height int
	data   []basics.Int8u
}

// NewSimpleAlphaMask creates a new simple alpha mask
func NewSimpleAlphaMask(width, height int) *SimpleAlphaMask {
	return &SimpleAlphaMask{
		width:  width,
		height: height,
		data:   make([]basics.Int8u, width*height),
	}
}

// Width returns the width of the mask
func (mask *SimpleAlphaMask) Width() int {
	return mask.width
}

// Height returns the height of the mask
func (mask *SimpleAlphaMask) Height() int {
	return mask.height
}

// Pixel returns the alpha value at the given coordinates
func (mask *SimpleAlphaMask) Pixel(x, y int) basics.Int8u {
	if x >= 0 && y >= 0 && x < mask.width && y < mask.height {
		return mask.data[y*mask.width+x]
	}
	return 0
}

// SetPixel sets the alpha value at the given coordinates
func (mask *SimpleAlphaMask) SetPixel(x, y int, alpha basics.Int8u) {
	if x >= 0 && y >= 0 && x < mask.width && y < mask.height {
		mask.data[y*mask.width+x] = alpha
	}
}

// CombinePixel combines the given coverage with the mask's alpha
func (mask *SimpleAlphaMask) CombinePixel(x, y int, cover basics.Int8u) basics.Int8u {
	maskAlpha := mask.Pixel(x, y)
	return basics.Int8u((uint32(cover) * uint32(maskAlpha)) / 255)
}

// FillHspan fills a horizontal span with mask alpha values
func (mask *SimpleAlphaMask) FillHspan(x, y int, dst []basics.Int8u, length int) {
	if y >= 0 && y < mask.height {
		for i := 0; i < length; i++ {
			if x+i >= 0 && x+i < mask.width {
				dst[i] = mask.data[y*mask.width+(x+i)]
			} else {
				dst[i] = 0
			}
		}
	} else {
		for i := 0; i < length; i++ {
			dst[i] = 0
		}
	}
}

// CombineHspan combines coverage values with mask alpha for a horizontal span
func (mask *SimpleAlphaMask) CombineHspan(x, y int, dst []basics.Int8u, length int) {
	if y >= 0 && y < mask.height {
		for i := 0; i < length; i++ {
			if x+i >= 0 && x+i < mask.width {
				maskAlpha := mask.data[y*mask.width+(x+i)]
				dst[i] = basics.Int8u((uint32(dst[i]) * uint32(maskAlpha)) / 255)
			} else {
				dst[i] = 0
			}
		}
	} else {
		for i := 0; i < length; i++ {
			dst[i] = 0
		}
	}
}

// FillVspan fills a vertical span with mask alpha values
func (mask *SimpleAlphaMask) FillVspan(x, y int, dst []basics.Int8u, length int) {
	if x >= 0 && x < mask.width {
		for i := 0; i < length; i++ {
			if y+i >= 0 && y+i < mask.height {
				dst[i] = mask.data[(y+i)*mask.width+x]
			} else {
				dst[i] = 0
			}
		}
	} else {
		for i := 0; i < length; i++ {
			dst[i] = 0
		}
	}
}

// CombineVspan combines coverage values with mask alpha for a vertical span
func (mask *SimpleAlphaMask) CombineVspan(x, y int, dst []basics.Int8u, length int) {
	if x >= 0 && x < mask.width {
		for i := 0; i < length; i++ {
			if y+i >= 0 && y+i < mask.height {
				maskAlpha := mask.data[(y+i)*mask.width+x]
				dst[i] = basics.Int8u((uint32(dst[i]) * uint32(maskAlpha)) / 255)
			} else {
				dst[i] = 0
			}
		}
	} else {
		for i := 0; i < length; i++ {
			dst[i] = 0
		}
	}
}

// Fill fills the entire mask with the given alpha value
func (mask *SimpleAlphaMask) Fill(alpha basics.Int8u) {
	for i := range mask.data {
		mask.data[i] = alpha
	}
}

// Clear clears the mask (sets all alpha to 0)
func (mask *SimpleAlphaMask) Clear() {
	mask.Fill(0)
}

// SetOpaque sets the mask to be fully opaque (alpha = 255)
func (mask *SimpleAlphaMask) SetOpaque() {
	mask.Fill(255)
}
