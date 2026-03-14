// Package pixfmt provides pixel format implementations for AGG.
// This file implements the alpha mask functionality from agg_alpha_mask_u8.h
package pixfmt

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/simd"
)

func zeroCovers(dst []basics.Int8u, count int) {
	for i := 0; i < count && i < len(dst); i++ {
		dst[i] = 0
	}
}

func fillMaskSpan(dst, src []basics.Int8u, step, offset, count int, maskFunc MaskFunction) {
	if count <= 0 || len(dst) < count {
		return
	}

	switch fn := maskFunc.(type) {
	case OneComponentMaskU8:
		fillOneComponentMaskSpan(dst, src, step, offset, count)
	case RGBToGrayMaskU8:
		fillRGBToGrayMaskSpan(dst, src, step, offset, count, fn)
	default:
		fillGenericMaskSpan(dst, src, step, offset, count, maskFunc)
	}
}

func combineMaskSpan(dst, src []basics.Int8u, step, offset, count int, maskFunc MaskFunction) {
	if count <= 0 || len(dst) < count {
		return
	}

	switch fn := maskFunc.(type) {
	case OneComponentMaskU8:
		combineOneComponentMaskSpan(dst, src, step, offset, count)
	case RGBToGrayMaskU8:
		combineRGBToGrayMaskSpan(dst, src, step, offset, count, fn)
	default:
		combineGenericMaskSpan(dst, src, step, offset, count, maskFunc)
	}
}

func fillOneComponentMaskSpan(dst, src []basics.Int8u, step, offset, count int) {
	if count <= 0 || len(dst) < count || offset < 0 {
		return
	}
	if step == 1 {
		if offset >= len(src) {
			zeroCovers(dst, count)
			return
		}
		end := offset + count
		if end > len(src) {
			copyCount := len(src) - offset
			if copyCount > 0 {
				simd.CopyMask1U8(dst[:copyCount], src[offset:], copyCount)
			}
			zeroCovers(dst[copyCount:], count-copyCount)
			return
		}
		simd.CopyMask1U8(dst[:count], src[offset:end], count)
		return
	}
	for i := 0; i < count; i++ {
		idx := offset + i*step
		if idx >= 0 && idx < len(src) {
			dst[i] = src[idx]
			continue
		}
		dst[i] = 0
	}
}

func combineOneComponentMaskSpan(dst, src []basics.Int8u, step, offset, count int) {
	if count <= 0 || len(dst) < count || offset < 0 {
		return
	}
	for i := 0; i < count; i++ {
		idx := offset + i*step
		if idx >= 0 && idx < len(src) {
			dst[i] = basics.Int8u((CoverFull + int(dst[i])*int(src[idx])) >> CoverShift)
			continue
		}
		dst[i] = 0
	}
}

func fillRGBToGrayMaskSpan(dst, src []basics.Int8u, step, offset, count int, fn RGBToGrayMaskU8) {
	if count <= 0 || len(dst) < count {
		return
	}
	if step == 3 && offset >= 0 && fn.ROffset == 0 && fn.GOffset == 1 && fn.BOffset == 2 {
		if offset >= len(src) {
			zeroCovers(dst, count)
			return
		}
		available := len(src) - offset
		maxCount := available / 3
		if maxCount > count {
			maxCount = count
		}
		if maxCount > 0 {
			simd.RGB24ToGrayU8(dst[:maxCount], src[offset:], maxCount)
		}
		if maxCount < count {
			zeroCovers(dst[maxCount:], count-maxCount)
		}
		return
	}
	for i := 0; i < count; i++ {
		base := offset + i*step
		if base < 0 {
			dst[i] = 0
			continue
		}
		maxOffset := base + basics.IMax(basics.IMax(fn.ROffset, fn.GOffset), fn.BOffset)
		if maxOffset >= len(src) {
			dst[i] = 0
			continue
		}
		dst[i] = basics.Int8u((int(src[base+fn.ROffset])*77 + int(src[base+fn.GOffset])*150 + int(src[base+fn.BOffset])*29) >> 8)
	}
}

func combineRGBToGrayMaskSpan(dst, src []basics.Int8u, step, offset, count int, fn RGBToGrayMaskU8) {
	if count <= 0 || len(dst) < count {
		return
	}
	for i := 0; i < count; i++ {
		base := offset + i*step
		if base < 0 {
			dst[i] = 0
			continue
		}
		maxOffset := base + basics.IMax(basics.IMax(fn.ROffset, fn.GOffset), fn.BOffset)
		if maxOffset >= len(src) {
			dst[i] = 0
			continue
		}
		gray := (int(src[base+fn.ROffset])*77 + int(src[base+fn.GOffset])*150 + int(src[base+fn.BOffset])*29) >> 8
		dst[i] = basics.Int8u((CoverFull + int(dst[i])*gray) >> CoverShift)
	}
}

func fillGenericMaskSpan(dst, src []basics.Int8u, step, offset, count int, maskFunc MaskFunction) {
	for i := 0; i < count; i++ {
		base := offset + i*step
		if base < 0 || base >= len(src) {
			dst[i] = 0
			continue
		}
		dst[i] = maskFunc.Calculate(src[base:])
	}
}

func combineGenericMaskSpan(dst, src []basics.Int8u, step, offset, count int, maskFunc MaskFunction) {
	for i := 0; i < count; i++ {
		base := offset + i*step
		if base < 0 || base >= len(src) {
			dst[i] = 0
			continue
		}
		maskVal := maskFunc.Calculate(src[base:])
		dst[i] = basics.Int8u((CoverFull + int(dst[i])*int(maskVal)) >> CoverShift)
	}
}

// MaskFunction defines the interface for mask calculation functions
type MaskFunction interface {
	Calculate(p []basics.Int8u) basics.Int8u
}

// OneComponentMaskU8 extracts a single component as mask value
type OneComponentMaskU8 struct{}

// Calculate returns the first byte as the mask value
func (m OneComponentMaskU8) Calculate(p []basics.Int8u) basics.Int8u {
	if len(p) > 0 {
		return p[0]
	}
	return 0
}

// RGBToGrayMaskU8 converts RGB to grayscale using weighted conversion
type RGBToGrayMaskU8 struct {
	ROffset int
	GOffset int
	BOffset int
}

// Calculate converts RGB to grayscale using standard luminance weights
func (m RGBToGrayMaskU8) Calculate(p []basics.Int8u) basics.Int8u {
	maxOffset := basics.IMax(basics.IMax(m.ROffset, m.GOffset), m.BOffset)
	if len(p) <= maxOffset {
		return 0
	}
	// Standard RGB to grayscale conversion: 0.299*R + 0.587*G + 0.114*B
	// Using integer approximation: (77*R + 150*G + 29*B) >> 8
	return basics.Int8u((int(p[m.ROffset])*77 + int(p[m.GOffset])*150 + int(p[m.BOffset])*29) >> 8)
}

// NewRGBToGrayMaskU8 creates a new RGB to grayscale mask with the specified offsets
func NewRGBToGrayMaskU8(rOffset, gOffset, bOffset int) RGBToGrayMaskU8 {
	return RGBToGrayMaskU8{
		ROffset: rOffset,
		GOffset: gOffset,
		BOffset: bOffset,
	}
}

// Cover scale constants
const (
	CoverShift = 8
)

// AlphaMaskU8 provides alpha masking with bounds checking
type AlphaMaskU8 struct {
	rbuf     *buffer.RenderingBufferU8
	maskFunc MaskFunction
	step     int
	offset   int
}

// NewAlphaMaskU8 creates a new alpha mask
func NewAlphaMaskU8(step, offset int, maskFunc MaskFunction) *AlphaMaskU8 {
	return &AlphaMaskU8{
		step:     step,
		offset:   offset,
		maskFunc: maskFunc,
	}
}

// NewAlphaMaskU8WithBuffer creates a new alpha mask with a rendering buffer
func NewAlphaMaskU8WithBuffer(rbuf *buffer.RenderingBufferU8, step, offset int, maskFunc MaskFunction) *AlphaMaskU8 {
	return &AlphaMaskU8{
		rbuf:     rbuf,
		step:     step,
		offset:   offset,
		maskFunc: maskFunc,
	}
}

// Attach attaches a rendering buffer to the mask
func (m *AlphaMaskU8) Attach(rbuf *buffer.RenderingBufferU8) {
	m.rbuf = rbuf
}

// MaskFunction returns the mask function
func (m *AlphaMaskU8) MaskFunction() MaskFunction {
	return m.maskFunc
}

// Width returns the width of the alpha mask
func (m *AlphaMaskU8) Width() int {
	if m.rbuf == nil {
		return 0
	}
	return m.rbuf.Width()
}

// Height returns the height of the alpha mask
func (m *AlphaMaskU8) Height() int {
	if m.rbuf == nil {
		return 0
	}
	return m.rbuf.Height()
}

// Pixel returns the mask value at the given coordinates
func (m *AlphaMaskU8) Pixel(x, y int) basics.Int8u {
	if m.rbuf == nil {
		return 0
	}

	if x >= 0 && y >= 0 && x < m.rbuf.Width() && y < m.rbuf.Height() {
		// Request enough bytes for the mask function to work
		// For RGB, we need at least 3 bytes starting from the offset
		length := m.step
		rowPtr := m.rbuf.RowPtr(x*m.step+m.offset, y, length)
		if rowPtr != nil && len(rowPtr) > 0 {
			return m.maskFunc.Calculate(rowPtr)
		}
	}
	return 0
}

// CombinePixel combines the given coverage with the mask's alpha at the coordinates
func (m *AlphaMaskU8) CombinePixel(x, y int, val basics.Int8u) basics.Int8u {
	if m.rbuf == nil {
		return 0
	}

	if x >= 0 && y >= 0 && x < m.rbuf.Width() && y < m.rbuf.Height() {
		length := m.step
		rowPtr := m.rbuf.RowPtr(x*m.step+m.offset, y, length)
		if rowPtr != nil && len(rowPtr) > 0 {
			maskVal := m.maskFunc.Calculate(rowPtr)
			return basics.Int8u((CoverFull + int(val)*int(maskVal)) >> CoverShift)
		}
	}
	return 0
}

// FillHspan fills a horizontal span with mask alpha values
func (m *AlphaMaskU8) FillHspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}

	xmax := m.rbuf.Width() - 1
	ymax := m.rbuf.Height() - 1

	count := numPix
	covers := dst

	// Check if y is out of bounds
	if y < 0 || y > ymax {
		zeroCovers(dst, numPix)
		return
	}

	// Handle negative x
	if x < 0 {
		count += x
		if count <= 0 {
			zeroCovers(dst, numPix)
			return
		}
		zeroCovers(covers, -x)
		covers = covers[-x:]
		x = 0
	}

	// Handle x + count exceeding width
	if x+count > xmax+1 {
		rest := x + count - xmax - 1
		count -= rest
		if count <= 0 {
			zeroCovers(dst, numPix)
			return
		}
		zeroCovers(covers[count:], rest)
	}

	maskRow := m.rbuf.Row(y)
	if maskRow == nil {
		zeroCovers(covers, count)
		return
	}
	fillMaskSpan(covers, maskRow, m.step, x*m.step+m.offset, count, m.maskFunc)
}

// CombineHspan combines coverage values with mask alpha for a horizontal span
func (m *AlphaMaskU8) CombineHspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}

	xmax := m.rbuf.Width() - 1
	ymax := m.rbuf.Height() - 1

	count := numPix
	covers := dst

	// Check if y is out of bounds
	if y < 0 || y > ymax {
		zeroCovers(dst, numPix)
		return
	}

	// Handle negative x
	if x < 0 {
		count += x
		if count <= 0 {
			zeroCovers(dst, numPix)
			return
		}
		zeroCovers(covers, -x)
		covers = covers[-x:]
		x = 0
	}

	// Handle x + count exceeding width
	if x+count > xmax+1 {
		rest := x + count - xmax - 1
		count -= rest
		if count <= 0 {
			zeroCovers(dst, numPix)
			return
		}
		zeroCovers(covers[count:], rest)
	}

	maskRow := m.rbuf.Row(y)
	if maskRow == nil {
		zeroCovers(covers, count)
		return
	}
	combineMaskSpan(covers, maskRow, m.step, x*m.step+m.offset, count, m.maskFunc)
}

// FillVspan fills a vertical span with mask alpha values
func (m *AlphaMaskU8) FillVspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}

	xmax := m.rbuf.Width() - 1
	ymax := m.rbuf.Height() - 1

	count := numPix
	covers := dst

	// Check if x is out of bounds
	if x < 0 || x > xmax {
		for i := 0; i < numPix; i++ {
			dst[i] = 0
		}
		return
	}

	// Handle negative y
	if y < 0 {
		count += y
		if count <= 0 {
			for i := 0; i < numPix; i++ {
				dst[i] = 0
			}
			return
		}
		// Fill negative portion with zeros
		for i := 0; i < -y; i++ {
			covers[i] = 0
		}
		covers = covers[-y:]
		y = 0
	}

	// Handle y + count exceeding height
	if y+count > ymax+1 {
		rest := y + count - ymax - 1
		count -= rest
		if count <= 0 {
			for i := 0; i < numPix; i++ {
				dst[i] = 0
			}
			return
		}
		// Fill overflow portion with zeros
		for i := count; i < count+rest && i < len(covers); i++ {
			covers[i] = 0
		}
	}

	// Fill the valid portion
	for i := 0; i < count; i++ {
		maskPtr := m.rbuf.RowPtr(x*m.step+m.offset, y+i, m.step)
		if maskPtr != nil && len(maskPtr) > 0 {
			covers[i] = m.maskFunc.Calculate(maskPtr)
		} else {
			covers[i] = 0
		}
	}
}

// CombineVspan combines coverage values with mask alpha for a vertical span
func (m *AlphaMaskU8) CombineVspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}

	xmax := m.rbuf.Width() - 1
	ymax := m.rbuf.Height() - 1

	count := numPix
	covers := dst

	// Check if x is out of bounds
	if x < 0 || x > xmax {
		for i := 0; i < numPix; i++ {
			dst[i] = 0
		}
		return
	}

	// Handle negative y
	if y < 0 {
		count += y
		if count <= 0 {
			for i := 0; i < numPix; i++ {
				dst[i] = 0
			}
			return
		}
		// Set negative portion to zero
		for i := 0; i < -y; i++ {
			covers[i] = 0
		}
		covers = covers[-y:]
		y = 0
	}

	// Handle y + count exceeding height
	if y+count > ymax+1 {
		rest := y + count - ymax - 1
		count -= rest
		if count <= 0 {
			for i := 0; i < numPix; i++ {
				dst[i] = 0
			}
			return
		}
		// Set overflow portion to zero
		for i := count; i < count+rest && i < len(covers); i++ {
			covers[i] = 0
		}
	}

	// Combine the valid portion
	for i := 0; i < count; i++ {
		maskPtr := m.rbuf.RowPtr(x*m.step+m.offset, y+i, m.step)
		if maskPtr != nil && len(maskPtr) > 0 {
			maskVal := m.maskFunc.Calculate(maskPtr)
			covers[i] = basics.Int8u((CoverFull + int(covers[i])*int(maskVal)) >> CoverShift)
		} else {
			covers[i] = 0
		}
	}
}

// AMaskNoClipU8 provides alpha masking without bounds checking (faster)
type AMaskNoClipU8 struct {
	rbuf     *buffer.RenderingBufferU8
	maskFunc MaskFunction
	step     int
	offset   int
}

// NewAMaskNoClipU8 creates a new alpha mask without clipping
func NewAMaskNoClipU8(step, offset int, maskFunc MaskFunction) *AMaskNoClipU8 {
	return &AMaskNoClipU8{
		step:     step,
		offset:   offset,
		maskFunc: maskFunc,
	}
}

// NewAMaskNoClipU8WithBuffer creates a new alpha mask without clipping with a rendering buffer
func NewAMaskNoClipU8WithBuffer(rbuf *buffer.RenderingBufferU8, step, offset int, maskFunc MaskFunction) *AMaskNoClipU8 {
	return &AMaskNoClipU8{
		rbuf:     rbuf,
		step:     step,
		offset:   offset,
		maskFunc: maskFunc,
	}
}

// Attach attaches a rendering buffer to the mask
func (m *AMaskNoClipU8) Attach(rbuf *buffer.RenderingBufferU8) {
	m.rbuf = rbuf
}

// MaskFunction returns the mask function
func (m *AMaskNoClipU8) MaskFunction() MaskFunction {
	return m.maskFunc
}

// Width returns the width of the alpha mask
func (m *AMaskNoClipU8) Width() int {
	if m.rbuf == nil {
		return 0
	}
	return m.rbuf.Width()
}

// Height returns the height of the alpha mask
func (m *AMaskNoClipU8) Height() int {
	if m.rbuf == nil {
		return 0
	}
	return m.rbuf.Height()
}

// Pixel returns the mask value at the given coordinates (no bounds checking)
func (m *AMaskNoClipU8) Pixel(x, y int) basics.Int8u {
	if m.rbuf == nil {
		return 0
	}

	rowPtr := m.rbuf.RowPtr(x*m.step+m.offset, y, m.step)
	if rowPtr != nil && len(rowPtr) > 0 {
		return m.maskFunc.Calculate(rowPtr)
	}
	return 0
}

// CombinePixel combines the given coverage with the mask's alpha (no bounds checking)
func (m *AMaskNoClipU8) CombinePixel(x, y int, val basics.Int8u) basics.Int8u {
	if m.rbuf == nil {
		return 0
	}

	rowPtr := m.rbuf.RowPtr(x*m.step+m.offset, y, m.step)
	if rowPtr != nil && len(rowPtr) > 0 {
		maskVal := m.maskFunc.Calculate(rowPtr)
		return basics.Int8u((CoverFull + int(val)*int(maskVal)) >> CoverShift)
	}
	return 0
}

// FillHspan fills a horizontal span with mask alpha values (no bounds checking)
func (m *AMaskNoClipU8) FillHspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}
	maskRow := m.rbuf.Row(y)
	if maskRow == nil {
		zeroCovers(dst, numPix)
		return
	}
	fillMaskSpan(dst, maskRow, m.step, x*m.step+m.offset, numPix, m.maskFunc)
}

// CombineHspan combines coverage values with mask alpha for a horizontal span (no bounds checking)
func (m *AMaskNoClipU8) CombineHspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}
	maskRow := m.rbuf.Row(y)
	if maskRow == nil {
		zeroCovers(dst, numPix)
		return
	}
	combineMaskSpan(dst, maskRow, m.step, x*m.step+m.offset, numPix, m.maskFunc)
}

// FillVspan fills a vertical span with mask alpha values (no bounds checking)
func (m *AMaskNoClipU8) FillVspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}

	for i := 0; i < numPix; i++ {
		maskPtr := m.rbuf.RowPtr(x*m.step+m.offset, y+i, m.step)
		if maskPtr != nil && len(maskPtr) > 0 {
			dst[i] = m.maskFunc.Calculate(maskPtr)
		} else {
			dst[i] = 0
		}
	}
}

// CombineVspan combines coverage values with mask alpha for a vertical span (no bounds checking)
func (m *AMaskNoClipU8) CombineVspan(x, y int, dst []basics.Int8u, numPix int) {
	if m.rbuf == nil || numPix <= 0 || len(dst) < numPix {
		return
	}

	for i := 0; i < numPix; i++ {
		maskPtr := m.rbuf.RowPtr(x*m.step+m.offset, y+i, m.step)
		if maskPtr != nil && len(maskPtr) > 0 {
			maskVal := m.maskFunc.Calculate(maskPtr)
			dst[i] = basics.Int8u((CoverFull + int(dst[i])*int(maskVal)) >> CoverShift)
		} else {
			dst[i] = 0
		}
	}
}

// Predefined alpha mask constructor functions for common pixel formats

// Grayscale alpha masks
type (
	AlphaMaskGray8   = *AlphaMaskU8
	AMaskNoClipGray8 = *AMaskNoClipU8
)

// RGB24 alpha masks (R, G, B components)
type (
	AlphaMaskRGB24R = *AlphaMaskU8
	AlphaMaskRGB24G = *AlphaMaskU8
	AlphaMaskRGB24B = *AlphaMaskU8
)

type (
	AMaskNoClipRGB24R = *AMaskNoClipU8
	AMaskNoClipRGB24G = *AMaskNoClipU8
	AMaskNoClipRGB24B = *AMaskNoClipU8
)

// BGR24 alpha masks (B, G, R components)
type (
	AlphaMaskBGR24R = *AlphaMaskU8
	AlphaMaskBGR24G = *AlphaMaskU8
	AlphaMaskBGR24B = *AlphaMaskU8
)

type (
	AMaskNoClipBGR24R = *AMaskNoClipU8
	AMaskNoClipBGR24G = *AMaskNoClipU8
	AMaskNoClipBGR24B = *AMaskNoClipU8
)

// RGBA32 alpha masks (R, G, B, A components)
type (
	AlphaMaskRGBA32R = *AlphaMaskU8
	AlphaMaskRGBA32G = *AlphaMaskU8
	AlphaMaskRGBA32B = *AlphaMaskU8
	AlphaMaskRGBA32A = *AlphaMaskU8
)

type (
	AMaskNoClipRGBA32R = *AMaskNoClipU8
	AMaskNoClipRGBA32G = *AMaskNoClipU8
	AMaskNoClipRGBA32B = *AMaskNoClipU8
	AMaskNoClipRGBA32A = *AMaskNoClipU8
)

// ARGB32 alpha masks (A, R, G, B components)
type (
	AlphaMaskARGB32R = *AlphaMaskU8
	AlphaMaskARGB32G = *AlphaMaskU8
	AlphaMaskARGB32B = *AlphaMaskU8
	AlphaMaskARGB32A = *AlphaMaskU8
)

type (
	AMaskNoClipARGB32R = *AMaskNoClipU8
	AMaskNoClipARGB32G = *AMaskNoClipU8
	AMaskNoClipARGB32B = *AMaskNoClipU8
	AMaskNoClipARGB32A = *AMaskNoClipU8
)

// BGRA32 alpha masks (B, G, R, A components)
type (
	AlphaMaskBGRA32R = *AlphaMaskU8
	AlphaMaskBGRA32G = *AlphaMaskU8
	AlphaMaskBGRA32B = *AlphaMaskU8
	AlphaMaskBGRA32A = *AlphaMaskU8
)

type (
	AMaskNoClipBGRA32R = *AMaskNoClipU8
	AMaskNoClipBGRA32G = *AMaskNoClipU8
	AMaskNoClipBGRA32B = *AMaskNoClipU8
	AMaskNoClipBGRA32A = *AMaskNoClipU8
)

// ABGR32 alpha masks (A, B, G, R components)
type (
	AlphaMaskABGR32R = *AlphaMaskU8
	AlphaMaskABGR32G = *AlphaMaskU8
	AlphaMaskABGR32B = *AlphaMaskU8
	AlphaMaskABGR32A = *AlphaMaskU8
)

type (
	AMaskNoClipABGR32R = *AMaskNoClipU8
	AMaskNoClipABGR32G = *AMaskNoClipU8
	AMaskNoClipABGR32B = *AMaskNoClipU8
	AMaskNoClipABGR32A = *AMaskNoClipU8
)

// RGB to grayscale conversion alpha masks
type (
	AlphaMaskRGB24Gray  = *AlphaMaskU8
	AlphaMaskBGR24Gray  = *AlphaMaskU8
	AlphaMaskRGBA32Gray = *AlphaMaskU8
	AlphaMaskARGB32Gray = *AlphaMaskU8
	AlphaMaskBGRA32Gray = *AlphaMaskU8
	AlphaMaskABGR32Gray = *AlphaMaskU8
)

type (
	AMaskNoClipRGB24Gray  = *AMaskNoClipU8
	AMaskNoClipBGR24Gray  = *AMaskNoClipU8
	AMaskNoClipRGBA32Gray = *AMaskNoClipU8
	AMaskNoClipARGB32Gray = *AMaskNoClipU8
	AMaskNoClipBGRA32Gray = *AMaskNoClipU8
	AMaskNoClipABGR32Gray = *AMaskNoClipU8
)

// Constructor helpers for predefined types

// NewAlphaMaskGray8 creates a new grayscale alpha mask
func NewAlphaMaskGray8() AlphaMaskGray8 {
	return NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
}

// NewAMaskNoClipGray8 creates a new grayscale alpha mask without clipping
func NewAMaskNoClipGray8() AMaskNoClipGray8 {
	return NewAMaskNoClipU8(1, 0, OneComponentMaskU8{})
}

// NewAlphaMaskRGB24Gray creates a new RGB24 to grayscale alpha mask
func NewAlphaMaskRGB24Gray() AlphaMaskRGB24Gray {
	return NewAlphaMaskU8(3, 0, NewRGBToGrayMaskU8(0, 1, 2))
}

// NewAMaskNoClipRGB24Gray creates a new RGB24 to grayscale alpha mask without clipping
func NewAMaskNoClipRGB24Gray() AMaskNoClipRGB24Gray {
	return NewAMaskNoClipU8(3, 0, NewRGBToGrayMaskU8(0, 1, 2))
}

// NewAlphaMaskBGR24Gray creates a new BGR24 to grayscale alpha mask
func NewAlphaMaskBGR24Gray() AlphaMaskBGR24Gray {
	return NewAlphaMaskU8(3, 0, NewRGBToGrayMaskU8(2, 1, 0))
}

// NewAMaskNoClipBGR24Gray creates a new BGR24 to grayscale alpha mask without clipping
func NewAMaskNoClipBGR24Gray() AMaskNoClipBGR24Gray {
	return NewAMaskNoClipU8(3, 0, NewRGBToGrayMaskU8(2, 1, 0))
}

// NewAlphaMaskRGB24R creates a new RGB24 R component alpha mask
func NewAlphaMaskRGB24R() AlphaMaskRGB24R {
	return NewAlphaMaskU8(3, 0, OneComponentMaskU8{})
}

// NewAlphaMaskRGB24G creates a new RGB24 G component alpha mask
func NewAlphaMaskRGB24G() AlphaMaskRGB24G {
	return NewAlphaMaskU8(3, 1, OneComponentMaskU8{})
}

// NewAlphaMaskRGB24B creates a new RGB24 B component alpha mask
func NewAlphaMaskRGB24B() AlphaMaskRGB24B {
	return NewAlphaMaskU8(3, 2, OneComponentMaskU8{})
}
