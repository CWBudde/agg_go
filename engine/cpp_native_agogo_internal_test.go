//go:build agogo && cgo

package engine

import (
	"errors"
	"math"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

var (
	_ Context = (*cppContext)(nil)
	_ Image   = (*cppImage)(nil)
)

func TestCurrentCPPNativeMetadataHasBuildID(t *testing.T) {
	meta := currentCPPNativeMetadata()
	if meta.BuildID == "" {
		t.Fatal("expected build id to be set")
	}
}

func TestProbeCPPNativeMatchesMetadata(t *testing.T) {
	meta := currentCPPNativeMetadata()
	err := probeCPPNative()
	if meta.Stub {
		if err == nil {
			t.Fatal("expected stub probe failure")
		}
		return
	}
	if err != nil {
		t.Fatalf("expected real native probe to succeed, got %v", err)
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
	if err := strokeCPPNativePath(img, path, nil, opts, 0, 0, 255, 255); err != nil {
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
	if err := strokeCPPNativePath(img, path, nil, opts, 0, 0, 0, 255); err == nil {
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

	// AGG trans_affine semantics: each incremental op composes so the FIRST call
	// is applied to the point first (innermost) and the LAST call is applied last
	// (outermost). So (1,0) flows translate → rotate → scale:
	//   translate(10,20): (1,0)   -> (11,20)
	//   rotate(90°):      (11,20) -> (-20,11)
	//   scale(2,1):       (-20,11)-> (-40,11)
	// This must match the faithful Port (internal/transform.TransAffine); an
	// earlier reversed-order matrix produced (10,22) instead.
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
	if math.Abs(x-(-40)) > 0.01 || math.Abs(y-11) > 0.01 {
		t.Fatalf("unexpected transformed point: got=(%.3f, %.3f) want≈(-40, 11)", x, y)
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

func TestCPPBackendContextDrawImageQuad(t *testing.T) {
	src, err := newCPPBackendImage(3, 3)
	if err != nil {
		t.Fatalf("newCPPBackendImage(src) error = %v", err)
	}
	srcCtx, err := newCPPBackendContextForImage(src)
	if err != nil {
		t.Fatalf("newCPPBackendContextForImage(src) error = %v", err)
	}
	srcCtx.Clear(agg.Blue)

	dst, err := newCPPBackendContext(12, 12)
	if err != nil {
		t.Fatalf("newCPPBackendContext(dst) error = %v", err)
	}
	dst.Clear(agg.White)

	quad := [8]float64{2, 1, 9, 2, 8, 9, 1, 8}
	if err := dst.DrawImageQuad(src, quad); err != nil {
		t.Fatalf("DrawImageQuad() error = %v", err)
	}

	inside := dst.GetImage().ToGoImage().RGBAAt(5, 5)
	if inside.B < 200 || inside.R != 0 || inside.G != 0 || inside.A != 255 {
		t.Fatalf("unexpected quad pixel: %+v", inside)
	}
	outside := dst.GetImage().ToGoImage().RGBAAt(10, 10)
	if outside.R != agg.White.R || outside.G != agg.White.G || outside.B != agg.White.B || outside.A != agg.White.A {
		t.Fatalf("unexpected pixel outside quad: %+v", outside)
	}
}

func TestCPPBackendContextFillLinearGradient(t *testing.T) {
	ctx, err := newCPPBackendContext(16, 10)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}
	ctx.Clear(agg.White)
	ctx.SetLinearGradient(2, 0, 14, 0, agg.Red, agg.Blue)
	ctx.FillRectangle(2, 1, 12, 8)

	left := ctx.GetImage().ToGoImage().RGBAAt(3, 5)
	right := ctx.GetImage().ToGoImage().RGBAAt(12, 5)
	if left.R <= left.B {
		t.Fatalf("expected left gradient pixel to skew red, got %+v", left)
	}
	if right.B <= right.R {
		t.Fatalf("expected right gradient pixel to skew blue, got %+v", right)
	}
}

func TestCPPBackendContextStrokeRadialGradient(t *testing.T) {
	ctx, err := newCPPBackendContext(18, 18)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}
	ctx.Clear(agg.White)
	ctx.SetLineWidth(3)
	ctx.SetStrokeRadialGradient(9, 9, 6, agg.Green, agg.Blue)
	ctx.DrawCircle(9, 9, 5)

	top := ctx.GetImage().ToGoImage().RGBAAt(9, 4)
	right := ctx.GetImage().ToGoImage().RGBAAt(14, 9)
	if (top.R == agg.White.R && top.G == agg.White.G && top.B == agg.White.B && top.A == agg.White.A) || top.A != 255 || (int(top.G)+int(top.B)) < 120 {
		t.Fatalf("expected stroke gradient pixel to be colored, got %+v", top)
	}
	if (right.R == agg.White.R && right.G == agg.White.G && right.B == agg.White.B && right.A == agg.White.A) || right.A != 255 || (int(right.G)+int(right.B)) < 120 {
		t.Fatalf("expected stroke gradient pixel to be colored, got %+v", right)
	}
	if ctx.GetFillGradientType() != agg.SolidGradient {
		t.Fatalf("unexpected fill gradient type: %v", ctx.GetFillGradientType())
	}
	if ctx.GetStrokeGradientType() != agg.RadialGradient {
		t.Fatalf("unexpected stroke gradient type: %v", ctx.GetStrokeGradientType())
	}
}

func TestCPPBackendTextOperationsAreTypedUnsupported(t *testing.T) {
	ctx, err := newCPPBackendContext(10, 10)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}

	if !currentCPPNativeMetadata().Stub {
		t.Skip("text is implemented in non-stub builds")
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

func TestCPPBackendContextDrawImageRegionQuadHonorsClip(t *testing.T) {
	src, err := newCPPBackendImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPBackendImage(src) error = %v", err)
	}
	srcCtx, err := newCPPBackendContextForImage(src)
	if err != nil {
		t.Fatalf("newCPPBackendContextForImage(src) error = %v", err)
	}
	srcCtx.Clear(agg.Green)

	dst, err := newCPPBackendContext(10, 10)
	if err != nil {
		t.Fatalf("newCPPBackendContext(dst) error = %v", err)
	}
	dst.Clear(agg.White)
	dst.ClipBox(3, 3, 8, 8)

	quad := [8]float64{1, 1, 9, 2, 8, 9, 2, 8}
	if err := dst.DrawImageRegionQuad(src, 1, 1, 2, 2, quad); err != nil {
		t.Fatalf("DrawImageRegionQuad() error = %v", err)
	}

	outside := dst.GetImage().ToGoImage().RGBAAt(2, 2)
	if outside.R != agg.White.R || outside.G != agg.White.G || outside.B != agg.White.B || outside.A != agg.White.A {
		t.Fatalf("unexpected pixel outside clip: %+v", outside)
	}
	inside := dst.GetImage().ToGoImage().RGBAAt(5, 5)
	if inside.G < 200 || inside.R != 0 || inside.B != 0 || inside.A != 255 {
		t.Fatalf("unexpected quad region pixel: %+v", inside)
	}
}

func TestCPPBackendExtendedBlendModeOnFill(t *testing.T) {
	ctx, err := newCPPBackendContext(8, 8)
	if err != nil {
		t.Fatalf("newCPPBackendContext() error = %v", err)
	}
	ctx.SetBlendMode(agg.BlendMultiply)

	if currentCPPNativeMetadata().Stub {
		// The stub vector path cannot honour comp-ops beyond the five image-blit
		// modes, so its native layer rejects the extended mode and must() panics
		// with that error. The stub backend is never advertised as available, so
		// the exact error type is not part of the contract — only that it fails
		// loudly rather than silently rendering a plain solid fill.
		defer func() {
			recovered := recover()
			if recovered == nil {
				t.Fatal("expected panic for unsupported blend mode in stub build")
			}
			if _, ok := recovered.(error); !ok {
				t.Fatalf("expected error panic, got %T", recovered)
			}
		}()
		ctx.FillRectangle(1, 1, 4, 4)
		return
	}

	// The real backend renders the full AGG comp-op set on the vector fill/stroke
	// path, so an extended mode like multiply must render rather than panic.
	// Multiply of opaque black over opaque white yields black.
	ctx.Clear(agg.White)
	ctx.SetFillColor(agg.NewColorRGB(0, 0, 0))
	ctx.FillRectangle(1, 1, 4, 4)
	got := ctx.GetImage().ToGoImage().RGBAAt(2, 2)
	if got.R > 4 || got.G > 4 || got.B > 4 || got.A != 255 {
		t.Fatalf("multiply of black over white = %+v, want opaque black", got)
	}
}

func TestCPPBackendExtendedBlendModeOnDrawImageQuad(t *testing.T) {
	src, err := newCPPBackendImage(4, 4)
	if err != nil {
		t.Fatalf("newCPPBackendImage(src) error = %v", err)
	}
	srcCtx, err := newCPPBackendContextForImage(src)
	if err != nil {
		t.Fatalf("newCPPBackendContextForImage(src) error = %v", err)
	}
	srcCtx.Clear(agg.NewColorRGB(128, 128, 128)) // opaque mid-grey tile

	dst, err := newCPPBackendContext(12, 12)
	if err != nil {
		t.Fatalf("newCPPBackendContext(dst) error = %v", err)
	}
	dst.Clear(agg.NewColorRGB(100, 150, 200))
	dst.SetBlendMode(agg.BlendMultiply)

	quad := [8]float64{2, 2, 10, 2, 10, 10, 2, 10}

	if currentCPPNativeMetadata().Stub {
		// The stub image-composite path only honours the five composite_pixel
		// operators, so it rejects an extended mode loudly. The stub backend is
		// never advertised as available, so the exact error type is not part of the
		// contract — only that it fails rather than silently rendering wrong.
		if err := dst.DrawImageQuad(src, quad); err == nil {
			t.Fatal("expected DrawImageQuad() under multiply to fail in stub build")
		}
		return
	}

	// The real backend blits images through comp_op_adaptor_rgba_plain, so an
	// extended mode like multiply must render rather than error. Multiply of the
	// grey tile over the background gives the per-channel product (~50,75,100).
	if err := dst.DrawImageQuad(src, quad); err != nil {
		t.Fatalf("DrawImageQuad() under multiply error = %v", err)
	}
	got := dst.GetImage().ToGoImage().RGBAAt(6, 6)
	within := func(got, want uint8, tol int) bool {
		d := int(got) - int(want)
		return d >= -tol && d <= tol
	}
	if !within(got.R, 50, 3) || !within(got.G, 75, 3) || !within(got.B, 100, 3) || got.A != 255 {
		t.Fatalf("multiply quad pixel = %+v, want ~(50,75,100,255)", got)
	}
}
