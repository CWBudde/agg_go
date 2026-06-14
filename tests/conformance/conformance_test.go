// Package conformance contains cross-backend conformance tests: the same scene
// corpus rendered through the Port and CPP engines and compared with documented
// tolerance envelopes.
//
// The test file is untagged. It compiles and runs in the default build, where
// engine.Available() returns only Port and the cross-backend test skips. Built
// with -tags "agogo aggreal" (and "freetype" for the text scene), the engine
// package links the real C++ AGG backend, engine.Available() includes CPP, and
// the same test bodies exercise Port-vs-CPP parity.
package conformance

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/cwbudde/agg_go/engine"
	"github.com/cwbudde/agg_go/engine/scene"
	"github.com/cwbudde/agg_go/tests/visual/framework"
)

// knownDivergence lists scenes whose Port and CPP outputs are known to differ
// beyond an AA-noise envelope because the C++ backend's implementation of that
// feature is still partial (see docs/BACKENDS.md and PLAN.md §5.5). For these
// the suite renders both engines and LOGS the measured difference as a tracked
// baseline, but does not fail: failing here would gate CI on a documented
// backend gap rather than a regression. Promote a scene to strict once the
// corresponding CPP parity work lands.
//
// Currently empty: the compositing scenes (src/srcover/clear) were promoted to
// strict once the C++ backend switched to faithful per-span comp-op rendering
// (PLAN.md §5.5). The mechanism is retained for the next partial feature.
var knownDivergence = map[string]string{}

// toleranceFor returns the documented comparison envelope for a strict scene.
// Both engines are 8-bit RGBA, so solid/path/clip scenes match everywhere
// except anti-aliased edges, where two independent rasterizers disagree on a
// small, bounded fraction of pixels. Gradients and image sampling diverge
// slightly more. Text uses the widest envelope. These are documented in
// docs/BACKENDS.md; tighten them as the C++ backend matures.
func toleranceFor(s scene.Scene) framework.ComparisonOptions {
	base := framework.ComparisonOptions{GenerateDiffImage: true}
	switch s.Name {
	case "gradient_linear", "gradient_radial":
		base.Tolerance = 3
		base.MaxDifferentRatio = 0.02
	case "image_scaled":
		// Independent image samplers disagree along the upscaled hard edges of
		// the source tile (4 quadrant seams + diagonal); ~6% of pixels, bulk
		// avg < 1 LSB. Same float-sampler-noise class as image_filters2.
		base.Tolerance = 4
		base.MaxDifferentRatio = 0.08
	case "text_basic":
		// Documented divergence: native AGG FreeType AA/hinting differs from the
		// Go port's. Loose envelope; demote to render-only if it proves larger.
		base.Tolerance = 8
		base.MaxDifferentRatio = 0.10
	case "compositing_srcover", "compositing_src", "compositing_clear",
		"compositing_multiply", "compositing_xor":
		// The C++ backend renders solid fills directly through a comp-op pixfmt
		// with a straight-alpha (premultiply/op/demultiply) adaptor that mirrors
		// the port's CompositeBlenderPlain, so these are byte-exact apart from
		// occasional 1-LSB rounding on anti-aliased edges. This holds for the full
		// AGG operator set (Porter-Duff plus separable blend modes), not just the
		// original five — both engines evaluate the same comp_op in premultiplied
		// space. compositing_xor adds a circle, so its disagreement is the usual
		// edge-AA band; bound it like the other strict compositing scenes.
		base.Tolerance = 2
		base.MaxDifferentRatio = 0.02
	case "compositing_gradient":
		// Gradient fill under a non-src-over blend: the gradient layer is
		// composited through the comp-op operator using the shape's AA coverage as
		// the rasterizer cover. Divergence is the gradient-interpolation rounding
		// (same as the gradient scenes) plus a thin rim at the shape edge where a
		// sub-pixel coverage disagreement under src flips a pixel between the
		// translucent gradient and the opaque background (large per-pixel delta,
		// but confined to ~0.8% of pixels on the circle perimeter).
		base.Tolerance = 3
		base.MaxDifferentRatio = 0.02
	default: // solid_fill_stroke, dashed_stroke, path_nonzero, path_evenodd, clip_box
		// Bound the fraction of disagreeing edge pixels; the bulk matches within
		// 2 LSB. AA edges span ~1.5% of these scenes (dashed_stroke ~1.3%: both
		// backends run AGG conv_dash + conv_stroke, so segments line up).
		base.Tolerance = 2
		base.MaxDifferentRatio = 0.025
	}
	return base
}

// render draws a scene through the given engine kind and returns the resulting
// image. It returns scene.ErrAssetUnavailable unchanged so callers can skip.
func render(s scene.Scene, kind engine.Kind, assets *scene.Assets) (image.Image, error) {
	ctx, err := engine.NewContext(s.Width, s.Height, engine.Config{Kind: kind})
	if err != nil {
		return nil, fmt.Errorf("new %s context: %w", kind, err)
	}
	if err := s.Draw(ctx, assets); err != nil {
		return nil, err
	}
	// Compare images as the engines naturally produce them; do not premultiply
	// or demultiply (the CPP backend lacks image_interop).
	return ctx.GetImage().ToGoImage(), nil
}

// skipReason classifies a render error as a skip-worthy condition (a missing
// asset, or an operation a backend explicitly reports as unsupported) versus a
// real failure. Treating typed capability errors as skips is the correct
// response to the documented CPP partial-coverage gaps (e.g. scaled image draw
// under an active transform).
func skipReason(err error) (bool, string) {
	switch {
	case err == nil:
		return false, ""
	case errors.Is(err, scene.ErrAssetUnavailable):
		return true, "asset unavailable: " + err.Error()
	case errors.Is(err, engine.ErrUnsupportedCapability):
		return true, "capability unsupported: " + err.Error()
	default:
		return false, ""
	}
}

func TestCrossBackendConformance(t *testing.T) {
	if !slices.Contains(engine.Available(), engine.CPP) {
		t.Skip("cpp backend unavailable in this build; run with -tags \"agogo aggreal\"")
	}

	portAssets, err := scene.BuildAssets(engine.Port)
	if err != nil {
		t.Fatalf("BuildAssets(Port): %v", err)
	}
	cppAssets, err := scene.BuildAssets(engine.CPP)
	if err != nil {
		t.Fatalf("BuildAssets(CPP): %v", err)
	}

	artifactDir := os.Getenv("CONFORMANCE_ARTIFACT_DIR")

	for _, s := range scene.All() {
		t.Run(s.Name, func(t *testing.T) {
			if !s.SupportedBy(engine.Port) || !s.SupportedBy(engine.CPP) {
				t.Skipf("scene caps %v not supported on both backends", s.Caps)
			}

			portImg, err := render(s, engine.Port, portAssets)
			if skip, msg := skipReason(err); skip {
				t.Skipf("port: %s", msg)
			}
			if err != nil {
				t.Fatalf("render port: %v", err)
			}

			cppImg, err := render(s, engine.CPP, cppAssets)
			if skip, msg := skipReason(err); skip {
				t.Skipf("cpp: %s", msg)
			}
			if err != nil {
				t.Fatalf("render cpp: %v", err)
			}

			// Always compute and log the difference with the diff-image enabled so
			// known-divergence baselines and failures both yield triage artifacts.
			res := framework.CompareImages(portImg, cppImg, framework.ComparisonOptions{
				Tolerance: 2, GenerateDiffImage: true,
			})
			t.Logf("scene=%s diff_pixels=%d/%d ratio=%.6f max=%d avg=%.4f",
				s.Name, res.DifferentPixels, res.TotalPixels, res.DifferentRatio,
				res.MaxDifference, res.AverageDifference)
			if artifactDir != "" {
				saveArtifacts(t, artifactDir, s.Name, portImg, cppImg, res.DiffImage)
			}

			if reason, known := knownDivergence[s.Name]; known {
				t.Logf("known divergence (not failed): %s", reason)
				return
			}

			// Strict scene: enforce the documented tolerance envelope.
			opts := toleranceFor(s)
			verdict := framework.CompareImages(portImg, cppImg, opts)
			if !verdict.Passed {
				t.Errorf("scene %s exceeds tolerance (tol=%d maxRatio=%.4f): diff_pixels=%d/%d ratio=%.6f max=%d avg=%.4f",
					s.Name, opts.Tolerance, opts.MaxDifferentRatio,
					verdict.DifferentPixels, verdict.TotalPixels, verdict.DifferentRatio,
					verdict.MaxDifference, verdict.AverageDifference)
			}
		})
	}
}

func saveArtifacts(t *testing.T, dir, name string, portImg, cppImg image.Image, diff *image.RGBA) {
	t.Helper()
	if err := framework.SaveImage(portImg, filepath.Join(dir, name+"_port.png")); err != nil {
		t.Logf("save port artifact: %v", err)
	}
	if err := framework.SaveImage(cppImg, filepath.Join(dir, name+"_cpp.png")); err != nil {
		t.Logf("save cpp artifact: %v", err)
	}
	if diff != nil {
		if err := framework.SaveDiffImage(diff, filepath.Join(dir, name+"_diff.png")); err != nil {
			t.Logf("save diff artifact: %v", err)
		}
	}
}
