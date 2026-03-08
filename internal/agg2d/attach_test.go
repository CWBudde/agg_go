// Package agg2d — Attach / AttachImage parity tests.
//
// C++ source reference: agg2d.cpp lines 125-157
//
//	void Agg2D::attach(unsigned char* buf, unsigned width, unsigned height, int stride)
//	void Agg2D::attach(Image& img)
package agg2d

import (
	"testing"
)

// pixelAtStride reads one RGBA pixel from a flat byte buffer.
func pixelAtStride(buf []byte, stride, x, y int) (r, g, b, a uint8) {
	i := y*stride + x*4
	return buf[i], buf[i+1], buf[i+2], buf[i+3]
}

// TestAttachResetsState verifies that Attach resets all mutable rendering state
// to the AGG defaults documented in agg2d.cpp:125-154:
//
//	m_renBase.reset_clipping(true)   → clip == full buffer
//	resetTransformations()           → identity transform
//	lineWidth(1.0)
//	lineColor(0,0,0)                 → black stroke
//	fillColor(255,255,255)           → white fill
//	textAlignment(AlignLeft,AlignBottom)
//	clipBox(0,0,width,height)        → full buffer clip
//	lineCap(CapRound)
//	lineJoin(JoinRound)
//	imageFilter(Bilinear)
//	imageResample(NoResample)
//	m_masterAlpha = 1.0
//	m_antiAliasGamma = 1.0
//	m_blendMode = BlendAlpha
func TestAttachResetsState(t *testing.T) {
	agg2d := NewAgg2D()

	buf := make([]uint8, 16*16*4)
	agg2d.Attach(buf, 16, 16, 16*4)

	// Mutate all state to non-default values.
	agg2d.LineWidth(5.0)
	agg2d.LineCap(CapSquare)
	agg2d.LineJoin(JoinBevel)
	agg2d.ImageFilter(Hanning)
	agg2d.ImageResample(ResampleAlways)
	agg2d.SetMasterAlpha(0.5)
	agg2d.ClipBox(4, 4, 8, 8)

	// Re-attach to the same buffer — must reset all state.
	agg2d.Attach(buf, 16, 16, 16*4)

	// Clip box must be reset to the full buffer (0,0,16,16).
	x1, y1, x2, y2 := agg2d.GetClipBox()
	if x1 != 0 || y1 != 0 || x2 != 16 || y2 != 16 {
		t.Fatalf("clip box after Attach = (%v,%v,%v,%v), want (0,0,16,16)", x1, y1, x2, y2)
	}

	// Line width resets to 1.0.
	if got := agg2d.GetLineWidth(); got != 1.0 {
		t.Fatalf("line width after Attach = %v, want 1.0", got)
	}

	// Line cap resets to CapRound.
	if got := agg2d.GetLineCap(); got != CapRound {
		t.Fatalf("line cap after Attach = %v, want CapRound", got)
	}

	// Line join resets to JoinRound.
	if got := agg2d.GetLineJoin(); got != JoinRound {
		t.Fatalf("line join after Attach = %v, want JoinRound", got)
	}

	// Image filter resets to Bilinear.
	if got := agg2d.GetImageFilter(); got != Bilinear {
		t.Fatalf("image filter after Attach = %v, want Bilinear", got)
	}

	// Image resample resets to NoResample.
	if got := agg2d.GetImageResample(); got != NoResample {
		t.Fatalf("image resample after Attach = %v, want NoResample", got)
	}

	// Master alpha resets to 1.0.
	if got := agg2d.GetMasterAlpha(); got != 1.0 {
		t.Fatalf("master alpha after Attach = %v, want 1.0", got)
	}
}

// TestAttachRendersToNewBuffer verifies that after re-attach to a different buffer,
// rendering writes pixels into the new buffer and not the old one.
func TestAttachRendersToNewBuffer(t *testing.T) {
	agg2d := NewAgg2D()

	old := make([]uint8, 8*8*4)
	agg2d.Attach(old, 8, 8, 8*4)

	// Switch to a fresh buffer.
	fresh := make([]uint8, 8*8*4)
	agg2d.Attach(fresh, 8, 8, 8*4)

	// Fill with red using ClearAll.
	agg2d.ClearAll([4]uint8{255, 0, 0, 255})

	// The new buffer should be red; the old buffer must be untouched.
	r, _, _, a := pixelAtStride(fresh, 8*4, 0, 0)
	if r != 255 || a != 255 {
		t.Fatalf("fresh buffer pixel after ClearAll = (%d,…,%d), want red (255,*,*,255)", r, a)
	}
	r, _, _, a = pixelAtStride(old, 8*4, 0, 0)
	if r != 0 || a != 0 {
		t.Fatalf("old buffer was written after re-attach: pixel = (%d,…,%d), want zero", r, a)
	}
}

// TestAttachImageDelegatesToAttach verifies that AttachImage(img) is equivalent to
// Attach(img.renBuf.buf(), img.renBuf.width(), img.renBuf.height(), img.renBuf.stride()).
//
// C++ reference: agg2d.cpp:153-157
//
//	void Agg2D::attach(Image& img) {
//	    attach(img.renBuf.buf(), img.renBuf.width(), img.renBuf.height(), img.renBuf.stride());
//	}
func TestAttachImageDelegatesToAttach(t *testing.T) {
	// Setup: draw red into an image via a separate context.
	imgBuf := make([]uint8, 4*4*4)
	img := NewImage(imgBuf, 4, 4, 4*4)

	// Fill imgBuf with magenta.
	for i := 0; i < len(imgBuf); i += 4 {
		imgBuf[i+0] = 255
		imgBuf[i+1] = 0
		imgBuf[i+2] = 255
		imgBuf[i+3] = 255
	}

	// Attach via image — state should be fully reset, rendering routed to imgBuf.
	agg2d := NewAgg2D()
	agg2d.AttachImage(img)

	// Clip box must match image dimensions (4×4).
	x1, y1, x2, y2 := agg2d.GetClipBox()
	if x1 != 0 || y1 != 0 || x2 != 4 || y2 != 4 {
		t.Fatalf("clip box after AttachImage = (%v,%v,%v,%v), want (0,0,4,4)", x1, y1, x2, y2)
	}

	// ClearAll should overwrite imgBuf with blue.
	agg2d.ClearAll([4]uint8{0, 0, 255, 255})

	r, g, b, a := pixelAtStride(imgBuf, 4*4, 0, 0)
	if r != 0 || g != 0 || b != 255 || a != 255 {
		t.Fatalf("imgBuf after ClearAll via AttachImage = (%d,%d,%d,%d), want blue", r, g, b, a)
	}
}

// TestAttachImageNilIsSafe verifies that AttachImage(nil) does not panic.
func TestAttachImageNilIsSafe(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 4*4*4)
	agg2d.Attach(buf, 4, 4, 4*4)

	// Should be a no-op, not a panic.
	agg2d.AttachImage(nil)

	x1, y1, x2, y2 := agg2d.GetClipBox()
	if x1 != 0 || y1 != 0 || x2 != 4 || y2 != 4 {
		t.Fatalf("clip box changed after AttachImage(nil): (%v,%v,%v,%v)", x1, y1, x2, y2)
	}
}

// TestAttachImageResetsStateLikeAttach verifies that AttachImage resets rendering state
// identically to Attach, as required by the C++ delegate implementation.
func TestAttachImageResetsStateLikeAttach(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 8*8*4)
	agg2d.Attach(buf, 8, 8, 8*4)

	// Mutate state.
	agg2d.LineWidth(7.0)
	agg2d.LineCap(CapSquare)
	agg2d.SetMasterAlpha(0.25)
	agg2d.ClipBox(2, 2, 6, 6)

	// AttachImage to a fresh image — all state must reset.
	imgBuf := make([]uint8, 8*8*4)
	img := NewImage(imgBuf, 8, 8, 8*4)
	agg2d.AttachImage(img)

	if got := agg2d.GetLineWidth(); got != 1.0 {
		t.Fatalf("line width after AttachImage = %v, want 1.0", got)
	}
	if got := agg2d.GetLineCap(); got != CapRound {
		t.Fatalf("line cap after AttachImage = %v, want CapRound", got)
	}
	if got := agg2d.GetMasterAlpha(); got != 1.0 {
		t.Fatalf("master alpha after AttachImage = %v, want 1.0", got)
	}

	x1, y1, x2, y2 := agg2d.GetClipBox()
	if x1 != 0 || y1 != 0 || x2 != 8 || y2 != 8 {
		t.Fatalf("clip box after AttachImage = (%v,%v,%v,%v), want (0,0,8,8)", x1, y1, x2, y2)
	}
}
