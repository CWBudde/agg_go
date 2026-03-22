package pixfmt

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
)

// AlphaMaskInterface is the Go equivalent of AGG alpha-mask classes used by
// pixfmt_amask_adaptor. It supplies per-pixel or per-span coverage that is
// combined with renderer-generated coverage before pixels reach the underlying
// pixfmt.
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

// PixFmtBlendInterface is the subset of pixfmt operations required by the
// alpha-mask adaptor. It intentionally matches AGG's span-oriented pixfmt
// surface rather than a minimal draw-pixel API, while remaining generic over
// the wrapped pixfmt color type.
type PixFmtBlendInterface[C any] interface {
	Width() int
	Height() int
	PixWidth() int
	Pixel(x, y int) C
	CopyPixel(x, y int, c C)
	BlendPixel(x, y int, c C, cover basics.Int8u)
	CopyHline(x, y, length int, c C)
	CopyVline(x, y, length int, c C)
	BlendHline(x, y, length int, c C, cover basics.Int8u)
	BlendVline(x, y, length int, c C, cover basics.Int8u)
	BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u)
	BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u)
	CopyBar(x1, y1, x2, y2 int, c C)
	BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u)
	CopyColorHspan(x, y, length int, colors []C)
	CopyColorVspan(x, y, length int, colors []C)
	BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u)
	BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u)
	Clear(c C)
	Fill(c C)
}

// PixFmtAMaskAdaptor is the Go equivalent of AGG's pixfmt_amask_adaptor.
//
// It leaves pixel reads unchanged but routes write coverage through an external
// alpha mask, which is how AGG composes clip masks and similar coverage maps
// without changing renderer code.
type PixFmtAMaskAdaptor[C any] struct {
	pixfmt PixFmtBlendInterface[C]
	amask  AlphaMaskInterface
	span   []basics.Int8u // Reusable span buffer
}

// SpanExtraTail matches AGG's span_extra_tail slack used to reduce span-buffer
// reallocations during many short masked draws.
const SpanExtraTail = 256

// NewPixFmtAMaskAdaptor creates a masked view over pixfmt using amask.
func NewPixFmtAMaskAdaptor[C any](pixfmt PixFmtBlendInterface[C], amask AlphaMaskInterface) *PixFmtAMaskAdaptor[C] {
	return &PixFmtAMaskAdaptor[C]{
		pixfmt: pixfmt,
		amask:  amask,
		span:   make([]basics.Int8u, 0),
	}
}

// AttachPixfmt replaces the wrapped pixel format.
func (pa *PixFmtAMaskAdaptor[C]) AttachPixfmt(pixfmt PixFmtBlendInterface[C]) {
	pa.pixfmt = pixfmt
}

// AttachAlphaMask replaces the alpha mask used for subsequent writes.
func (pa *PixFmtAMaskAdaptor[C]) AttachAlphaMask(amask AlphaMaskInterface) {
	pa.amask = amask
}

// reallocSpan ensures the reusable span buffer can hold length covers.
func (pa *PixFmtAMaskAdaptor[C]) reallocSpan(length int) {
	if length > len(pa.span) {
		pa.span = make([]basics.Int8u, length+SpanExtraTail)
	}
}

// initSpan fills the reusable span buffer with cover-full values.
func (pa *PixFmtAMaskAdaptor[C]) initSpan(length int) {
	pa.reallocSpan(length)
	for i := 0; i < length; i++ {
		pa.span[i] = 255 // Full coverage
	}
}

// initSpanWithCovers copies the caller-provided covers into the reusable span.
func (pa *PixFmtAMaskAdaptor[C]) initSpanWithCovers(length int, covers []basics.Int8u) {
	pa.reallocSpan(length)
	copy(pa.span[:length], covers[:length])
}

// Width returns the wrapped pixfmt width.
func (pa *PixFmtAMaskAdaptor[C]) Width() int {
	return pa.pixfmt.Width()
}

// Height returns the wrapped pixfmt height.
func (pa *PixFmtAMaskAdaptor[C]) Height() int {
	return pa.pixfmt.Height()
}

// PixWidth returns the wrapped pixfmt storage width in bytes.
func (pa *PixFmtAMaskAdaptor[C]) PixWidth() int {
	return pa.pixfmt.PixWidth()
}

// Pixel reads through to the wrapped pixfmt; the alpha mask only affects writes.
func (pa *PixFmtAMaskAdaptor[C]) Pixel(x, y int) C {
	return pa.pixfmt.Pixel(x, y)
}

// CopyPixel writes c using the mask value as coverage.
func (pa *PixFmtAMaskAdaptor[C]) CopyPixel(x, y int, c C) {
	cover := pa.amask.Pixel(x, y)
	pa.pixfmt.BlendPixel(x, y, c, cover)
}

// BlendPixel combines the caller's cover with the mask before blending.
func (pa *PixFmtAMaskAdaptor[C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
	combinedCover := pa.amask.CombinePixel(x, y, cover)
	pa.pixfmt.BlendPixel(x, y, c, combinedCover)
}

// CopyHline writes a solid horizontal run using mask coverage.
func (pa *PixFmtAMaskAdaptor[C]) CopyHline(x, y, length int, c C) {
	pa.reallocSpan(length)
	pa.amask.FillHspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidHspan(x, y, length, c, pa.span[:length])
}

// BlendHline combines uniform input coverage with the mask across one row.
func (pa *PixFmtAMaskAdaptor[C]) BlendHline(x, y, length int, c C, cover basics.Int8u) {
	pa.initSpan(length)
	for i := 0; i < length; i++ {
		pa.span[i] = cover
	}
	pa.amask.CombineHspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidHspan(x, y, length, c, pa.span[:length])
}

// CopyVline writes a solid vertical run using mask coverage.
func (pa *PixFmtAMaskAdaptor[C]) CopyVline(x, y, length int, c C) {
	pa.reallocSpan(length)
	pa.amask.FillVspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidVspan(x, y, length, c, pa.span[:length])
}

// BlendVline combines uniform input coverage with the mask down one column.
func (pa *PixFmtAMaskAdaptor[C]) BlendVline(x, y, length int, c C, cover basics.Int8u) {
	pa.initSpan(length)
	for i := 0; i < length; i++ {
		pa.span[i] = cover
	}
	pa.amask.CombineVspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendSolidVspan(x, y, length, c, pa.span[:length])
}

// BlendSolidHspan combines AA covers with the mask across a horizontal span.
func (pa *PixFmtAMaskAdaptor[C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
		pa.amask.CombineHspan(x, y, pa.span[:length], length)
		pa.pixfmt.BlendSolidHspan(x, y, length, c, pa.span[:length])
	} else {
		pa.CopyHline(x, y, length, c)
	}
}

// BlendSolidVspan combines AA covers with the mask across a vertical span.
func (pa *PixFmtAMaskAdaptor[C]) BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
		pa.amask.CombineVspan(x, y, pa.span[:length], length)
		pa.pixfmt.BlendSolidVspan(x, y, length, c, pa.span[:length])
	} else {
		pa.CopyVline(x, y, length, c)
	}
}

// CopyBar writes a solid rectangle using mask coverage.
func (pa *PixFmtAMaskAdaptor[C]) CopyBar(x1, y1, x2, y2 int, c C) {
	for y := y1; y <= y2; y++ {
		pa.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar combines uniform rectangle coverage with the mask.
func (pa *PixFmtAMaskAdaptor[C]) BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u) {
	for y := y1; y <= y2; y++ {
		pa.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// CopyColorHspan matches AGG: derive per-pixel covers from the mask and route
// through BlendColorHspan on the wrapped pixfmt.
func (pa *PixFmtAMaskAdaptor[C]) CopyColorHspan(x, y, length int, colors []C) {
	pa.reallocSpan(length)
	pa.amask.FillHspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendColorHspan(x, y, length, colors, pa.span[:length], basics.CoverFull)
}

// CopyColorVspan matches AGG: derive per-pixel covers from the mask and route
// through BlendColorVspan on the wrapped pixfmt.
func (pa *PixFmtAMaskAdaptor[C]) CopyColorVspan(x, y, length int, colors []C) {
	pa.reallocSpan(length)
	pa.amask.FillVspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendColorVspan(x, y, length, colors, pa.span[:length], basics.CoverFull)
}

// Clear fills the wrapped pixfmt directly; the alpha mask does not participate.
func (pa *PixFmtAMaskAdaptor[C]) Clear(c C) {
	pa.pixfmt.Clear(c)
}

// Fill is an alias for Clear.
func (pa *PixFmtAMaskAdaptor[C]) Fill(c C) {
	pa.pixfmt.Fill(c)
}

// BlendColorHspan combines either per-pixel covers or a uniform cover with the
// mask, then forwards the blended span to the wrapped pixfmt.
func (pa *PixFmtAMaskAdaptor[C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
	} else {
		pa.reallocSpan(length)
		for i := 0; i < length; i++ {
			pa.span[i] = cover
		}
	}
	pa.amask.CombineHspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendColorHspan(x, y, length, colors, pa.span[:length], cover)
}

// BlendColorVspan combines either per-pixel covers or a uniform cover with the
// mask, then forwards the blended span to the wrapped pixfmt.
func (pa *PixFmtAMaskAdaptor[C]) BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	if covers != nil {
		pa.initSpanWithCovers(length, covers)
	} else {
		pa.reallocSpan(length)
		for i := 0; i < length; i++ {
			pa.span[i] = cover
		}
	}
	pa.amask.CombineVspan(x, y, pa.span[:length], length)
	pa.pixfmt.BlendColorVspan(x, y, length, colors, pa.span[:length], cover)
}

// CopyFrom copies from another rendering buffer (not affected by mask for source)
func (pa *PixFmtAMaskAdaptor[C]) CopyFrom(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int,
) {
	if copier, ok := pa.pixfmt.(interface {
		CopyFrom(src interface {
			RowData(y int) []basics.Int8u
			Width() int
			Height() int
		}, xdst, ydst, xsrc, ysrc, length int)
	}); ok {
		copier.CopyFrom(src, xdst, ydst, xsrc, ysrc, length)
		return
	}

	if ysrc < 0 || ysrc >= src.Height() {
		return
	}
	bytesPerPixel := detectBytesPerPixel(src, ysrc)
	srcRowData := src.RowData(ysrc)
	if srcRowData == nil {
		return
	}

	for i := 0; i < length; i++ {
		srcPixelX := xsrc + i
		dstPixelX := xdst + i
		if srcPixelX < 0 || srcPixelX >= src.Width() {
			continue
		}

		srcColor, ok := decodeRGBA8FromRowData(srcRowData, bytesPerPixel, srcPixelX)
		if !ok {
			continue
		}

		if !copyPixelCompat(pa.pixfmt, dstPixelX, ydst, srcColor) {
			return
		}
	}
}

// BlendFrom blends from another pixel format with mask applied
func (pa *PixFmtAMaskAdaptor[C]) BlendFrom(src interface {
	Pixel(x, y int) C
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	for i := 0; i < length; i++ {
		if xsrc+i >= 0 && xsrc+i < src.Width() && ysrc >= 0 && ysrc < src.Height() {
			srcPixel := src.Pixel(xsrc+i, ysrc)
			combinedCover := pa.amask.CombinePixel(xdst+i, ydst, cover)
			pa.pixfmt.BlendPixel(xdst+i, ydst, srcPixel, combinedCover)
		}
	}
}

func copyPixelCompat[C any](dst PixFmtBlendInterface[C], x, y int, src color.RGBA8[color.Linear]) bool {
	switch typed := any(dst).(type) {
	case interface {
		CopyPixel(int, int, color.RGBA8[color.Linear])
	}:
		typed.CopyPixel(x, y, src)
		return true
	case interface {
		CopyPixel(int, int, color.RGB8[color.Linear])
	}:
		typed.CopyPixel(x, y, color.RGB8[color.Linear]{R: src.R, G: src.G, B: src.B})
		return true
	default:
		return false
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
