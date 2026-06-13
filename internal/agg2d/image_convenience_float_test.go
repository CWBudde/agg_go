package agg2d

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

// TestAgg2DFloatCopyImageSimple verifies the whole-image float-dst convenience
// copy lands the source at the rounded destination, mirroring the 8-bit
// CopyImageSimple semantics (WorldToScreen + integer truncation of dst).
func TestAgg2DFloatCopyImageSimple(t *testing.T) {
	a, dst := setupFloatTarget(12, 12)
	src := filledFloatImage(4, 4, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 1.0))

	if err := a.CopyImageSimple(src, 3, 3); err != nil {
		t.Fatalf("CopyImageSimple: %v", err)
	}

	in := dst.GetPixel(4, 4)
	if !approxF(in.R, 0.2) || !approxF(in.G, 0.4) || !approxF(in.B, 0.6) || !approxF(in.A, 1.0) {
		t.Fatalf("copied pixel(4,4) = %+v, want {0.2,0.4,0.6,1}", in)
	}
	if out := dst.GetPixel(0, 0); out.A != 0 {
		t.Fatalf("outside-copy pixel(0,0) alpha = %v, want 0", out.A)
	}
}

// TestAgg2DFloatBlendImageSimple verifies the whole-image float-dst convenience
// blend honors the requested alpha.
func TestAgg2DFloatBlendImageSimple(t *testing.T) {
	a, dst := setupFloatTarget(12, 12)
	src := filledFloatImage(4, 4, color.NewRGBA32[color.Linear](1.0, 0.0, 0.0, 1.0))

	if err := a.BlendImageSimple(src, 3, 3, 255); err != nil {
		t.Fatalf("BlendImageSimple: %v", err)
	}

	in := dst.GetPixel(5, 5)
	if !approxF(in.R, 1.0) || in.A <= 0 {
		t.Fatalf("blended pixel(5,5) = %+v, want opaque red", in)
	}
}

// TestAgg2DFloatBlendImageDefaultAlpha verifies the default-alpha (255) whole-
// image blend behaves identically to BlendImage with cover 255.
func TestAgg2DFloatBlendImageDefaultAlpha(t *testing.T) {
	a, dst := setupFloatTarget(12, 12)
	src := filledFloatImage(4, 4, color.NewRGBA32[color.Linear](0.0, 1.0, 0.0, 1.0))

	a.BlendImageDefaultAlpha(src, 3, 3)

	in := dst.GetPixel(5, 5)
	if !approxF(in.G, 1.0) || in.A <= 0 {
		t.Fatalf("blended pixel(5,5) = %+v, want opaque green", in)
	}
}

// TestAgg2DFloatBlendImageSimpleDefaultAlpha verifies the float-dst default-alpha
// convenience blend.
func TestAgg2DFloatBlendImageSimpleDefaultAlpha(t *testing.T) {
	a, dst := setupFloatTarget(12, 12)
	src := filledFloatImage(4, 4, color.NewRGBA32[color.Linear](0.0, 0.0, 1.0, 1.0))

	if err := a.BlendImageSimpleDefaultAlpha(src, 3, 3); err != nil {
		t.Fatalf("BlendImageSimpleDefaultAlpha: %v", err)
	}

	in := dst.GetPixel(5, 5)
	if !approxF(in.B, 1.0) || in.A <= 0 {
		t.Fatalf("blended pixel(5,5) = %+v, want opaque blue", in)
	}
}

// TestAgg2DFloatCopyImageSimpleNil verifies a nil source returns an error like
// the 8-bit twin.
func TestAgg2DFloatCopyImageSimpleNil(t *testing.T) {
	a, _ := setupFloatTarget(8, 8)
	if err := a.CopyImageSimple(nil, 0, 0); err == nil {
		t.Fatal("CopyImageSimple(nil) should return an error")
	}
	if err := a.BlendImageSimple(nil, 0, 0, 255); err == nil {
		t.Fatal("BlendImageSimple(nil) should return an error")
	}
}

// TestAgg2DFloatSaveImagePPM verifies the attached float buffer is written as a
// binary PPM with the straight RGB channels rounded to 8-bit, matching the
// 8-bit twin's SaveImagePPM format (header + RGB triples, no alpha).
func TestAgg2DFloatSaveImagePPM(t *testing.T) {
	const w, h = 3, 2
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	// Paint a known straight color across the whole buffer.
	a.ClearAll(NewColor(200, 100, 50, 255))

	path := filepath.Join(t.TempDir(), "out.ppm")
	if err := a.SaveImagePPM(path); err != nil {
		t.Fatalf("SaveImagePPM: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ppm: %v", err)
	}

	r := bufio.NewReader(bytes.NewReader(raw))
	var magic string
	var gotW, gotH, maxv int
	if _, err := readPPMHeader(r, &magic, &gotW, &gotH, &maxv); err != nil {
		t.Fatalf("parse header: %v", err)
	}
	if magic != "P6" || gotW != w || gotH != h || maxv != 255 {
		t.Fatalf("header = %q %d %d %d, want P6 %d %d 255", magic, gotW, gotH, maxv, w, h)
	}

	body := make([]byte, w*h*3)
	if _, err := io.ReadFull(r, body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	// roundToU8(200/255) etc. — straight channels written verbatim.
	wantR, wantG, wantB := roundToU8(200.0/255), roundToU8(100.0/255), roundToU8(50.0/255)
	for i := 0; i < w*h; i++ {
		gr, gg, gb := body[i*3], body[i*3+1], body[i*3+2]
		if gr != wantR || gg != wantG || gb != wantB {
			t.Fatalf("pixel %d = (%d,%d,%d), want (%d,%d,%d)", i, gr, gg, gb, wantR, wantG, wantB)
		}
	}
}

// TestAgg2DFloatSaveImagePPMNoBuffer verifies SaveImagePPM errors when nothing
// is attached.
func TestAgg2DFloatSaveImagePPMNoBuffer(t *testing.T) {
	a := NewAgg2DFloat()
	path := filepath.Join(t.TempDir(), "nope.ppm")
	if err := a.SaveImagePPM(path); err == nil {
		t.Fatal("SaveImagePPM with no attached buffer should error")
	}
}

// readPPMHeader parses a binary PPM header ("P6\n<w> <h>\n<max>\n").
func readPPMHeader(r *bufio.Reader, magic *string, w, h, maxv *int) (int, error) {
	if _, err := fscan(r, magic); err != nil {
		return 0, err
	}
	if _, err := fscanInt(r, w); err != nil {
		return 0, err
	}
	if _, err := fscanInt(r, h); err != nil {
		return 0, err
	}
	return fscanInt(r, maxv)
}

func fscan(r *bufio.Reader, s *string) (int, error) {
	var b []byte
	for {
		c, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			if len(b) == 0 {
				continue
			}
			break
		}
		b = append(b, c)
	}
	*s = string(b)
	return 1, nil
}

func fscanInt(r *bufio.Reader, v *int) (int, error) {
	var s string
	if _, err := fscan(r, &s); err != nil {
		return 0, err
	}
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	*v = n
	return 1, nil
}
