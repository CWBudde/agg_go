//go:build agogo && cgo

package engine

import (
	"errors"
	"math"
	"strings"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

var (
	_ Context = (*cppContext)(nil)
	_ Image   = (*cppImage)(nil)
)

func TestCurrentCPPNativeMetadataReportsStub(t *testing.T) {
	meta := currentCPPNativeMetadata()
	if !meta.Stub {
		t.Fatal("expected agogo native backend metadata to report stub mode")
	}
	if meta.BuildID == "" {
		t.Fatal("expected build id to be set")
	}
}

func TestProbeCPPNativeReturnsStubFailure(t *testing.T) {
	err := probeCPPNative()
	if err == nil {
		t.Fatal("expected stub probe failure")
	}
	if !strings.Contains(err.Error(), "stub") {
		t.Fatalf("expected stub failure message, got %q", err.Error())
	}
}

func TestCPPNativeImageCreateClearAndReadback(t *testing.T) {
	img, err := newCPPNativeImage(8, 6)
	if err != nil {
		t.Fatalf("newCPPNativeImage() error = %v", err)
	}
	defer img.close()

	if img.width() != 8 || img.height() != 6 || img.stride() != 32 {
		t.Fatalf("unexpected image geometry: %dx%d stride=%d", img.width(), img.height(), img.stride())
	}
	if err := img.clear(10, 20, 30, 255); err != nil {
		t.Fatalf("clear() error = %v", err)
	}
	pixels, err := img.pixelsRGBA()
	if err != nil {
		t.Fatalf("pixelsRGBA() error = %v", err)
	}
	if len(pixels) != img.stride()*img.height() {
		t.Fatalf("unexpected pixel buffer length: got=%d want=%d", len(pixels), img.stride()*img.height())
	}
	if pixels[0] != 10 || pixels[1] != 20 || pixels[2] != 30 || pixels[3] != 255 {
		t.Fatalf("unexpected first pixel: %v", pixels[:4])
	}
}

func TestFillCPPNativePathFillsRectangle(t *testing.T) {
	img, err := newCPPNativeImage(12, 12)
	if err != nil {
		t.Fatalf("newCPPNativeImage() error = %v", err)
	}
	defer img.close()
	if err := img.clear(255, 255, 255, 255); err != nil {
		t.Fatalf("clear() error = %v", err)
	}

	path, err := newCPPNativePath()
	if err != nil {
		t.Fatalf("newCPPNativePath() error = %v", err)
	}
	defer path.close()
	if err := path.moveTo(2, 2); err != nil {
		t.Fatalf("moveTo() error = %v", err)
	}
	if err := path.lineTo(10, 2); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}
	if err := path.lineTo(10, 10); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}
	if err := path.lineTo(2, 10); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}
	if err := path.closePath(); err != nil {
		t.Fatalf("closePath() error = %v", err)
	}

	if err := fillCPPNativePath(img, path, cppNativeFillRuleNonZero, 255, 0, 0, 255); err != nil {
		t.Fatalf("fillCPPNativePath() error = %v", err)
	}

	goImg, err := img.toGoImage()
	if err != nil {
		t.Fatalf("toGoImage() error = %v", err)
	}
	center := goImg.RGBAAt(5, 5)
	if center.R < 200 || center.G != 0 || center.B != 0 || center.A != 255 {
		t.Fatalf("unexpected filled pixel: %+v", center)
	}
}

func TestFillCPPNativePathRejectsDegeneratePath(t *testing.T) {
	img, err := newCPPNativeImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPNativeImage() error = %v", err)
	}
	defer img.close()

	path, err := newCPPNativePath()
	if err != nil {
		t.Fatalf("newCPPNativePath() error = %v", err)
	}
	defer path.close()
	if err := path.moveTo(1, 1); err != nil {
		t.Fatalf("moveTo() error = %v", err)
	}
	if err := path.lineTo(2, 2); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}

	if err := fillCPPNativePath(img, path, cppNativeFillRuleNonZero, 0, 0, 0, 255); err == nil {
		t.Fatal("expected fillCPPNativePath() to reject degenerate path")
	}
}

func TestStrokeCPPNativePathDrawsHorizontalLine(t *testing.T) {
	img, err := newCPPNativeImage(16, 16)
	if err != nil {
		t.Fatalf("newCPPNativeImage() error = %v", err)
	}
	defer img.close()
	if err := img.clear(255, 255, 255, 255); err != nil {
		t.Fatalf("clear() error = %v", err)
	}

	path, err := newCPPNativePath()
	if err != nil {
		t.Fatalf("newCPPNativePath() error = %v", err)
	}
	defer path.close()
	if err := path.moveTo(3, 8); err != nil {
		t.Fatalf("moveTo() error = %v", err)
	}
	if err := path.lineTo(13, 8); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}

	opts := defaultCPPNativeStrokeOptions()
	opts.Width = 3
	opts.LineCap = cppNativeLineCapRound
	opts.LineJoin = cppNativeLineJoinRound
	if err := strokeCPPNativePath(img, path, opts, 0, 0, 255, 255); err != nil {
		t.Fatalf("strokeCPPNativePath() error = %v", err)
	}

	goImg, err := img.toGoImage()
	if err != nil {
		t.Fatalf("toGoImage() error = %v", err)
	}
	center := goImg.RGBAAt(8, 8)
	if center.B < 200 || center.R != 0 || center.G != 0 || center.A != 255 {
		t.Fatalf("unexpected stroked center pixel: %+v", center)
	}
}

func TestStrokeCPPNativePathRejectsInvalidOptions(t *testing.T) {
	img, err := newCPPNativeImage(8, 8)
	if err != nil {
		t.Fatalf("newCPPNativeImage() error = %v", err)
	}
	defer img.close()

	path, err := newCPPNativePath()
	if err != nil {
		t.Fatalf("newCPPNativePath() error = %v", err)
	}
	defer path.close()
	if err := path.moveTo(1, 1); err != nil {
		t.Fatalf("moveTo() error = %v", err)
	}
	if err := path.lineTo(6, 6); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}

	opts := defaultCPPNativeStrokeOptions()
	opts.Width = 0
	if err := strokeCPPNativePath(img, path, opts, 0, 0, 0, 255); err == nil {
		t.Fatal("expected strokeCPPNativePath() to reject zero width")
	}
}

func TestCPPNativeImageBlitCopiesRegion(t *testing.T) {
	src, err := newCPPNativeImage(6, 6)
	if err != nil {
		t.Fatalf("newCPPNativeImage(src) error = %v", err)
	}
	defer src.close()
	if err := src.clear(255, 0, 0, 255); err != nil {
		t.Fatalf("src.clear() error = %v", err)
	}

	dst, err := newCPPNativeImage(10, 10)
	if err != nil {
		t.Fatalf("newCPPNativeImage(dst) error = %v", err)
	}
	defer dst.close()
	if err := dst.clear(255, 255, 255, 255); err != nil {
		t.Fatalf("dst.clear() error = %v", err)
	}

	if err := dst.blitFrom(src, 2, 3, 1, 1, 3, 3); err != nil {
		t.Fatalf("blitFrom() error = %v", err)
	}

	goImg, err := dst.toGoImage()
	if err != nil {
		t.Fatalf("toGoImage() error = %v", err)
	}
	inside := goImg.RGBAAt(3, 4)
	if inside.R < 200 || inside.G != 0 || inside.B != 0 || inside.A != 255 {
		t.Fatalf("unexpected blitted pixel: %+v", inside)
	}
}

func TestCPPNativeImageBlitRejectsOutOfBoundsRegion(t *testing.T) {
	src, err := newCPPNativeImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPNativeImage(src) error = %v", err)
	}
	defer src.close()
	dst, err := newCPPNativeImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPNativeImage(dst) error = %v", err)
	}
	defer dst.close()

	if err := dst.blitFrom(src, 2, 2, 0, 0, 3, 3); err == nil {
		t.Fatal("expected blitFrom() to reject out-of-bounds destination region")
	}
}

func TestCPPNativeMatrixTransformPointTranslateRotateScale(t *testing.T) {
	matrix, err := newCPPNativeMatrix()
	if err != nil {
		t.Fatalf("newCPPNativeMatrix() error = %v", err)
	}
	defer matrix.close()

	if err := matrix.translate(10, 20); err != nil {
		t.Fatalf("translate() error = %v", err)
	}
	if err := matrix.rotateDegrees(90); err != nil {
		t.Fatalf("rotateDegrees() error = %v", err)
	}
	if err := matrix.scale(2, 1); err != nil {
		t.Fatalf("scale() error = %v", err)
	}

	x, y, err := matrix.transformPoint(1, 0)
	if err != nil {
		t.Fatalf("transformPoint() error = %v", err)
	}
	if math.Abs(x-10) > 0.01 || math.Abs(y-22) > 0.01 {
		t.Fatalf("unexpected transformed point: got=(%.3f, %.3f) want≈(10, 22)", x, y)
	}
}

func TestCPPNativePathTransformMovesFilledGeometry(t *testing.T) {
	path, err := newCPPNativePath()
	if err != nil {
		t.Fatalf("newCPPNativePath() error = %v", err)
	}
	defer path.close()
	if err := path.moveTo(1, 1); err != nil {
		t.Fatalf("moveTo() error = %v", err)
	}
	if err := path.lineTo(5, 1); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}
	if err := path.lineTo(5, 5); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}
	if err := path.lineTo(1, 5); err != nil {
		t.Fatalf("lineTo() error = %v", err)
	}
	if err := path.closePath(); err != nil {
		t.Fatalf("closePath() error = %v", err)
	}

	matrix, err := newCPPNativeMatrix()
	if err != nil {
		t.Fatalf("newCPPNativeMatrix() error = %v", err)
	}
	defer matrix.close()
	if err := matrix.translate(6, 3); err != nil {
		t.Fatalf("translate() error = %v", err)
	}

	transformed, err := path.transform(matrix)
	if err != nil {
		t.Fatalf("transform() error = %v", err)
	}
	defer transformed.close()

	img, err := newCPPNativeImage(16, 16)
	if err != nil {
		t.Fatalf("newCPPNativeImage() error = %v", err)
	}
	defer img.close()
	if err := img.clear(255, 255, 255, 255); err != nil {
		t.Fatalf("clear() error = %v", err)
	}
	if err := fillCPPNativePath(img, transformed, cppNativeFillRuleNonZero, 0, 255, 0, 255); err != nil {
		t.Fatalf("fillCPPNativePath() error = %v", err)
	}

	goImg, err := img.toGoImage()
	if err != nil {
		t.Fatalf("toGoImage() error = %v", err)
	}
	inside := goImg.RGBAAt(8, 5)
	if inside.G < 200 || inside.R != 0 || inside.B != 0 || inside.A != 255 {
		t.Fatalf("unexpected transformed fill pixel: %+v", inside)
	}
}

func TestCPPBackendContextFillRectangle(t *testing.T) {
	ctx, err := newCPPBackendContext(16, 16)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}

	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(2, 2, 10, 10)

	got := ctx.GetImage().ToGoImage().RGBAAt(6, 6)
	if got.R < 200 || got.G != 0 || got.B != 0 || got.A != 255 {
		t.Fatalf("unexpected rendered pixel: %+v", got)
	}
}

func TestCPPBackendContextDrawImage(t *testing.T) {
	src, err := newCPPBackendImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPBackendImage(src) error = %v", err)
	}
	ctxSrc, err := newCPPBackendContextForImage(src)
	if err != nil {
		t.Fatalf("newCPPBackendContextForImage(src) error = %v", err)
	}
	ctxSrc.Clear(agg.Blue)

	dst, err := newCPPBackendContext(12, 12)
	if err != nil {
		t.Fatalf("newCPPBackendContext(dst) error = %v", err)
	}
	dst.Clear(agg.White)
	if err := dst.DrawImage(src, 3, 4); err != nil {
		t.Fatalf("DrawImage() error = %v", err)
	}

	got := dst.GetImage().ToGoImage().RGBAAt(4, 5)
	if got.B < 200 || got.R != 0 || got.G != 0 || got.A != 255 {
		t.Fatalf("unexpected drawn image pixel: %+v", got)
	}
}

func TestCPPBackendContextTransformAffectsFill(t *testing.T) {
	ctx, err := newCPPBackendContext(20, 20)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Green)
	ctx.Translate(5, 4)
	ctx.FillRectangle(1, 1, 4, 4)

	got := ctx.GetImage().ToGoImage().RGBAAt(7, 6)
	if got.G < 200 || got.R != 0 || got.B != 0 || got.A != 255 {
		t.Fatalf("unexpected transformed fill pixel: %+v", got)
	}
}

func TestCPPBackendContextClipBoxClipsFill(t *testing.T) {
	ctx, err := newCPPBackendContext(16, 16)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}

	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.ClipBox(4, 4, 8, 8)
	ctx.FillRectangle(2, 2, 8, 8)

	inside := ctx.GetImage().ToGoImage().RGBAAt(5, 5)
	if inside.R < 200 || inside.G != 0 || inside.B != 0 || inside.A != 255 {
		t.Fatalf("unexpected clipped inside pixel: %+v", inside)
	}
	outside := ctx.GetImage().ToGoImage().RGBAAt(3, 3)
	if outside.R != agg.White.R || outside.G != agg.White.G || outside.B != agg.White.B || outside.A != agg.White.A {
		t.Fatalf("unexpected clipped outside pixel: %+v", outside)
	}
}

func TestCPPBackendContextDrawImageScaled(t *testing.T) {
	ctx, err := newCPPBackendContext(10, 10)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}
	img, err := newCPPBackendImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPBackendImage() error = %v", err)
	}
	srcCtx, err := newCPPBackendContextForImage(img)
	if err != nil {
		t.Fatalf("newCPPBackendContextForImage() error = %v", err)
	}
	srcCtx.Clear(agg.Blue)
	ctx.Clear(agg.White)

	err = ctx.DrawImageScaled(img, 1, 1, 8, 8)
	if err != nil {
		t.Fatalf("DrawImageScaled() error = %v", err)
	}

	got := ctx.GetImage().ToGoImage().RGBAAt(7, 7)
	if got.B < 200 || got.R != 0 || got.G != 0 || got.A != 255 {
		t.Fatalf("unexpected scaled image pixel: %+v", got)
	}
}

func TestCPPBackendTextOperationsAreTypedUnsupported(t *testing.T) {
	ctx, err := newCPPBackendContext(10, 10)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}

	err = ctx.DrawText("hello", 1, 1)
	if err == nil {
		t.Fatal("expected DrawText() to fail")
	}
	if !errors.Is(err, ErrUnsupportedCapability) {
		t.Fatalf("expected ErrUnsupportedCapability, got %v", err)
	}
}

func TestCPPBackendContextDrawImageRegionScaledHonorsClipAndBlend(t *testing.T) {
	src, err := newCPPBackendImage(2, 2)
	if err != nil {
		t.Fatalf("newCPPBackendImage(src) error = %v", err)
	}
	srcCtx, err := newCPPBackendContextForImage(src)
	if err != nil {
		t.Fatalf("newCPPBackendContextForImage(src) error = %v", err)
	}
	srcCtx.Clear(agg.Color{R: 255, G: 0, B: 0, A: 128})

	dst, err := newCPPBackendContext(8, 8)
	if err != nil {
		t.Fatalf("newCPPBackendContext(dst) error = %v", err)
	}
	dst.Clear(agg.White)
	dst.ClipBox(2, 2, 6, 6)

	if err := dst.DrawImageRegion(src, 0, 0, 2, 2, 1, 1, 6, 6); err != nil {
		t.Fatalf("DrawImageRegion() error = %v", err)
	}

	clippedOut := dst.GetImage().ToGoImage().RGBAAt(1, 1)
	if clippedOut.R != agg.White.R || clippedOut.G != agg.White.G || clippedOut.B != agg.White.B || clippedOut.A != agg.White.A {
		t.Fatalf("unexpected pixel outside clip: %+v", clippedOut)
	}
	inside := dst.GetImage().ToGoImage().RGBAAt(3, 3)
	if inside.R < 240 || inside.G > 140 || inside.B > 140 || inside.A != 255 {
		t.Fatalf("unexpected blended pixel inside clip: %+v", inside)
	}
}

func TestCPPBackendUnsupportedBlendModePanicsTypedOnFill(t *testing.T) {
	ctx, err := newCPPBackendContext(8, 8)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}
	ctx.SetBlendMode(agg.BlendMultiply)

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic for unsupported blend mode")
		}
		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("expected error panic, got %T", recovered)
		}
		if !errors.Is(err, ErrUnsupportedCapability) {
			t.Fatalf("expected ErrUnsupportedCapability panic, got %v", err)
		}
	}()

	ctx.FillRectangle(1, 1, 4, 4)
}
