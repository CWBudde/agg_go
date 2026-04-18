//go:build agogo && cgo

package cppbridge

import (
	"math"
	"testing"
)

func TestImageCreateClearAndReadback(t *testing.T) {
	img, err := NewImage(8, 6)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	defer img.Close()

	if img.Width() != 8 || img.Height() != 6 || img.Stride() != 32 {
		t.Fatalf("unexpected image geometry: %dx%d stride=%d", img.Width(), img.Height(), img.Stride())
	}

	if err := img.Clear(10, 20, 30, 255); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	pixels, err := img.PixelsRGBA()
	if err != nil {
		t.Fatalf("PixelsRGBA() error = %v", err)
	}
	if len(pixels) != img.Stride()*img.Height() {
		t.Fatalf("unexpected pixel buffer length: got=%d want=%d", len(pixels), img.Stride()*img.Height())
	}
	if pixels[0] != 10 || pixels[1] != 20 || pixels[2] != 30 || pixels[3] != 255 {
		t.Fatalf("unexpected first pixel: %v", pixels[:4])
	}
}

func TestFillPathFillsRectangle(t *testing.T) {
	img, err := NewImage(12, 12)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	defer img.Close()
	if err := img.Clear(255, 255, 255, 255); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	path, err := NewPath()
	if err != nil {
		t.Fatalf("NewPath() error = %v", err)
	}
	defer path.Close()
	if err := path.MoveTo(2, 2); err != nil {
		t.Fatalf("MoveTo() error = %v", err)
	}
	if err := path.LineTo(10, 2); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}
	if err := path.LineTo(10, 10); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}
	if err := path.LineTo(2, 10); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}
	if err := path.ClosePath(); err != nil {
		t.Fatalf("ClosePath() error = %v", err)
	}

	if err := FillPath(img, path, FillRuleNonZero, 255, 0, 0, 255); err != nil {
		t.Fatalf("FillPath() error = %v", err)
	}

	goImg, err := img.ToGoImage()
	if err != nil {
		t.Fatalf("ToGoImage() error = %v", err)
	}
	center := goImg.RGBAAt(5, 5)
	if center.R < 200 || center.G != 0 || center.B != 0 || center.A != 255 {
		t.Fatalf("unexpected filled pixel: %+v", center)
	}
	outside := goImg.RGBAAt(0, 0)
	if outside.R != 255 || outside.G != 255 || outside.B != 255 || outside.A != 255 {
		t.Fatalf("unexpected outside pixel: %+v", outside)
	}
}

func TestFillPathRejectsDegeneratePath(t *testing.T) {
	img, err := NewImage(4, 4)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	defer img.Close()

	path, err := NewPath()
	if err != nil {
		t.Fatalf("NewPath() error = %v", err)
	}
	defer path.Close()
	if err := path.MoveTo(1, 1); err != nil {
		t.Fatalf("MoveTo() error = %v", err)
	}
	if err := path.LineTo(2, 2); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}

	if err := FillPath(img, path, FillRuleNonZero, 0, 0, 0, 255); err == nil {
		t.Fatal("expected FillPath() to reject degenerate path")
	}
}

func TestStrokePathDrawsHorizontalLine(t *testing.T) {
	img, err := NewImage(16, 16)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	defer img.Close()
	if err := img.Clear(255, 255, 255, 255); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	path, err := NewPath()
	if err != nil {
		t.Fatalf("NewPath() error = %v", err)
	}
	defer path.Close()
	if err := path.MoveTo(3, 8); err != nil {
		t.Fatalf("MoveTo() error = %v", err)
	}
	if err := path.LineTo(13, 8); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}

	opts := DefaultStrokeOptions()
	opts.Width = 3
	opts.LineCap = LineCapRound
	opts.LineJoin = LineJoinRound
	if err := StrokePath(img, path, opts, 0, 0, 255, 255); err != nil {
		t.Fatalf("StrokePath() error = %v", err)
	}

	goImg, err := img.ToGoImage()
	if err != nil {
		t.Fatalf("ToGoImage() error = %v", err)
	}
	center := goImg.RGBAAt(8, 8)
	if center.B < 200 || center.R != 0 || center.G != 0 || center.A != 255 {
		t.Fatalf("unexpected stroked center pixel: %+v", center)
	}
	outside := goImg.RGBAAt(8, 2)
	if outside.R != 255 || outside.G != 255 || outside.B != 255 || outside.A != 255 {
		t.Fatalf("unexpected outside pixel: %+v", outside)
	}
}

func TestStrokePathRejectsInvalidOptions(t *testing.T) {
	img, err := NewImage(8, 8)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	defer img.Close()

	path, err := NewPath()
	if err != nil {
		t.Fatalf("NewPath() error = %v", err)
	}
	defer path.Close()
	if err := path.MoveTo(1, 1); err != nil {
		t.Fatalf("MoveTo() error = %v", err)
	}
	if err := path.LineTo(6, 6); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}

	opts := DefaultStrokeOptions()
	opts.Width = 0
	if err := StrokePath(img, path, opts, 0, 0, 0, 255); err == nil {
		t.Fatal("expected StrokePath() to reject zero width")
	}
}

func TestImageBlitCopiesRegion(t *testing.T) {
	src, err := NewImage(6, 6)
	if err != nil {
		t.Fatalf("NewImage(src) error = %v", err)
	}
	defer src.Close()
	if err := src.Clear(255, 0, 0, 255); err != nil {
		t.Fatalf("src.Clear() error = %v", err)
	}

	dst, err := NewImage(10, 10)
	if err != nil {
		t.Fatalf("NewImage(dst) error = %v", err)
	}
	defer dst.Close()
	if err := dst.Clear(255, 255, 255, 255); err != nil {
		t.Fatalf("dst.Clear() error = %v", err)
	}

	if err := dst.BlitFrom(src, 2, 3, 1, 1, 3, 3); err != nil {
		t.Fatalf("BlitFrom() error = %v", err)
	}

	goImg, err := dst.ToGoImage()
	if err != nil {
		t.Fatalf("ToGoImage() error = %v", err)
	}
	inside := goImg.RGBAAt(3, 4)
	if inside.R < 200 || inside.G != 0 || inside.B != 0 || inside.A != 255 {
		t.Fatalf("unexpected blitted pixel: %+v", inside)
	}
	outside := goImg.RGBAAt(0, 0)
	if outside.R != 255 || outside.G != 255 || outside.B != 255 || outside.A != 255 {
		t.Fatalf("unexpected untouched pixel: %+v", outside)
	}
}

func TestImageBlitRejectsOutOfBoundsRegion(t *testing.T) {
	src, err := NewImage(4, 4)
	if err != nil {
		t.Fatalf("NewImage(src) error = %v", err)
	}
	defer src.Close()
	dst, err := NewImage(4, 4)
	if err != nil {
		t.Fatalf("NewImage(dst) error = %v", err)
	}
	defer dst.Close()

	if err := dst.BlitFrom(src, 2, 2, 0, 0, 3, 3); err == nil {
		t.Fatal("expected BlitFrom() to reject out-of-bounds destination region")
	}
}

func TestMatrixTransformPointTranslateRotateScale(t *testing.T) {
	matrix, err := NewMatrix()
	if err != nil {
		t.Fatalf("NewMatrix() error = %v", err)
	}
	defer matrix.Close()

	if err := matrix.Translate(10, 20); err != nil {
		t.Fatalf("Translate() error = %v", err)
	}
	if err := matrix.RotateDegrees(90); err != nil {
		t.Fatalf("RotateDegrees() error = %v", err)
	}
	if err := matrix.Scale(2, 1); err != nil {
		t.Fatalf("Scale() error = %v", err)
	}

	x, y, err := matrix.TransformPoint(1, 0)
	if err != nil {
		t.Fatalf("TransformPoint() error = %v", err)
	}
	if math.Abs(x-10) > 0.01 || math.Abs(y-22) > 0.01 {
		t.Fatalf("unexpected transformed point: got=(%.3f, %.3f) want≈(10, 22)", x, y)
	}
}

func TestPathTransformMovesFilledGeometry(t *testing.T) {
	path, err := NewPath()
	if err != nil {
		t.Fatalf("NewPath() error = %v", err)
	}
	defer path.Close()
	if err := path.MoveTo(1, 1); err != nil {
		t.Fatalf("MoveTo() error = %v", err)
	}
	if err := path.LineTo(5, 1); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}
	if err := path.LineTo(5, 5); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}
	if err := path.LineTo(1, 5); err != nil {
		t.Fatalf("LineTo() error = %v", err)
	}
	if err := path.ClosePath(); err != nil {
		t.Fatalf("ClosePath() error = %v", err)
	}

	matrix, err := NewMatrix()
	if err != nil {
		t.Fatalf("NewMatrix() error = %v", err)
	}
	defer matrix.Close()
	if err := matrix.Translate(6, 3); err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	transformed, err := path.Transform(matrix)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	defer transformed.Close()

	img, err := NewImage(16, 16)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	defer img.Close()
	if err := img.Clear(255, 255, 255, 255); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if err := FillPath(img, transformed, FillRuleNonZero, 0, 255, 0, 255); err != nil {
		t.Fatalf("FillPath() error = %v", err)
	}

	goImg, err := img.ToGoImage()
	if err != nil {
		t.Fatalf("ToGoImage() error = %v", err)
	}
	inside := goImg.RGBAAt(8, 5)
	if inside.G < 200 || inside.R != 0 || inside.B != 0 || inside.A != 255 {
		t.Fatalf("unexpected transformed fill pixel: %+v", inside)
	}
	outside := goImg.RGBAAt(2, 2)
	if outside.R != 255 || outside.G != 255 || outside.B != 255 || outside.A != 255 {
		t.Fatalf("unexpected untouched pixel: %+v", outside)
	}
}
