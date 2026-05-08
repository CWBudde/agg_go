// Package main provides an HTTP server for visual comparison of AGG rendering outputs.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/cwbudde/agg_go/tests/visual/framework"
)

const (
	cppDir = "tests/visual/reference/cpp/examples"
	goDir  = "tests/visual/reference/go/examples"
)

type demoEntry struct {
	Name        string
	RMSE        float64
	AvgDiff     float64
	MaxDiff     uint8
	DiffPixels  int
	TotalPixels int
	DiffRatio   float64
	CppB64      string
	GoB64       string
	RawDiffB64  string
	AmpDiffB64  string
}

type demoConfig struct {
	name string
	dir  string
}

var demoConfigs = []demoConfig{
	{name: "aa_demo", dir: "examples/core/intermediate/aa_demo"},
	{name: "aa_test", dir: "examples/core/intermediate/aa_test"},
	{name: "alpha_gradient", dir: "examples/core/intermediate/alpha_gradient"},
	{name: "alpha_mask", dir: "examples/core/intermediate/alpha_mask"},
	{name: "alpha_mask2", dir: "examples/core/intermediate/alpha_mask2"},
	{name: "alpha_mask3", dir: "examples/core/intermediate/alpha_mask3"},
	{name: "bezier_div", dir: "examples/core/intermediate/bezier_div"},
	{name: "blend_color", dir: "examples/core/intermediate/blend_color"},
	{name: "blur", dir: "examples/core/intermediate/blur"},
	{name: "bspline", dir: "examples/core/intermediate/bspline"},
	{name: "circles", dir: "examples/core/basic/circles"},
	{name: "component_rendering", dir: "examples/core/basic/component_rendering"},
	{name: "compositing", dir: "examples/core/intermediate/compositing"},
	{name: "compositing2", dir: "examples/core/intermediate/compositing2"},
	{name: "conv_contour", dir: "examples/core/intermediate/conv_contour"},
	{name: "conv_dash_marker", dir: "examples/core/intermediate/conv_dash_marker"},
	{name: "conv_stroke", dir: "examples/core/intermediate/conv_stroke"},
	{name: "distortions", dir: "examples/core/advanced/distortions"},
	{name: "flash_rasterizer", dir: "examples/core/advanced/flash_rasterizer"},
	{name: "flash_rasterizer2", dir: "examples/core/intermediate/flash_rasterizer2"},
	{name: "gamma_correction", dir: "examples/core/advanced/gamma_correction"},
	{name: "gamma_ctrl", dir: "examples/core/advanced/gamma_ctrl"},
	{name: "gamma_tuner", dir: "examples/core/advanced/gamma_tuner"},
	{name: "gouraud", dir: "examples/core/intermediate/gouraud"},
	{name: "gouraud_mesh", dir: "examples/core/intermediate/gouraud_mesh"},
	{name: "gradient_focal", dir: "examples/core/intermediate/gradient_focal"},
	{name: "gradients", dir: "examples/core/intermediate/gradients"},
	{name: "gradients_contour", dir: "examples/core/intermediate/gradients_contour"},
	{name: "graph_test", dir: "examples/core/intermediate/graph_test"},
	{name: "idea", dir: "examples/core/intermediate/idea"},
	{name: "image1", dir: "examples/core/intermediate/image1"},
	{name: "image_alpha", dir: "examples/core/intermediate/image_alpha"},
	{name: "image_filters", dir: "examples/core/intermediate/image_filters"},
	{name: "image_filters2", dir: "examples/core/intermediate/image_filters2"},
	{name: "image_fltr_graph", dir: "examples/core/intermediate/image_fltr_graph"},
	{name: "image_perspective", dir: "examples/core/intermediate/image_perspective"},
	{name: "image_resample", dir: "examples/core/intermediate/image_resample"},
	{name: "image_transforms", dir: "examples/core/intermediate/image_transforms"},
	{name: "line_patterns", dir: "examples/core/intermediate/line_patterns"},
	{name: "line_patterns_clip", dir: "examples/core/intermediate/line_patterns_clip"},
	{name: "line_thickness", dir: "examples/core/intermediate/line_thickness"},
	{name: "lion", dir: "examples/core/intermediate/lion"},
	{name: "lion_lens", dir: "examples/core/intermediate/lion_lens"},
	{name: "lion_outline", dir: "examples/core/intermediate/lion_outline"},
	{name: "mol_view", dir: "examples/core/intermediate/mol_view"},
	{name: "multi_clip", dir: "examples/core/basic/multi_clip"},
	{name: "pattern_fill", dir: "examples/core/intermediate/pattern_fill"},
	{name: "pattern_perspective", dir: "examples/core/intermediate/pattern_perspective"},
	{name: "pattern_resample", dir: "examples/core/intermediate/pattern_resample"},
	{name: "perspective", dir: "examples/core/intermediate/perspective"},
	{name: "polymorphic_renderer", dir: "examples/core/intermediate/polymorphic_renderer"},
	{name: "raster_text", dir: "examples/core/intermediate/raster_text"},
	{name: "rasterizer_compound", dir: "examples/core/intermediate/rasterizer_compound"},
	{name: "rasterizers", dir: "examples/core/intermediate/rasterizers"},
	{name: "rasterizers2", dir: "examples/core/intermediate/rasterizers2"},
	{name: "rounded_rect", dir: "examples/core/basic/rounded_rect"},
	{name: "scanline_boolean", dir: "examples/core/intermediate/scanline_boolean"},
	{name: "scanline_boolean2", dir: "examples/core/intermediate/scanline_boolean2"},
	{name: "simple_blur", dir: "examples/core/basic/simple_blur"},
	{name: "trans_polar", dir: "examples/core/intermediate/trans_polar"},
}

var regenerateMu sync.Mutex

func findDemoConfig(name string) (demoConfig, bool) {
	for _, demo := range demoConfigs {
		if demo.name == name {
			return demo, true
		}
	}
	return demoConfig{}, false
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func pngToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func absDiff8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// rawSubtractImage computes per-channel absolute difference.
// Identical pixels are shown as green (#00aa00) so they are clearly distinguishable
// from black difference pixels.
func rawSubtractImage(ref, gen image.Image) *image.RGBA {
	bounds := ref.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rc := ref.At(bounds.Min.X+x, bounds.Min.Y+y)
			gc := gen.At(bounds.Min.X+x, bounds.Min.Y+y)
			rr, rg, rb, _ := rc.RGBA()
			gr, gg, gb, _ := gc.RGBA()
			dr := absDiff8(uint8(rr>>8), uint8(gr>>8))
			dg := absDiff8(uint8(rg>>8), uint8(gg>>8))
			db := absDiff8(uint8(rb>>8), uint8(gb>>8))
			if dr == 0 && dg == 0 && db == 0 {
				out.Set(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
			} else {
				out.Set(x, y, color.RGBA{R: dr, G: dg, B: db, A: 255})
			}
		}
	}
	return out
}

func buildEntry(name, cppPath, goPath string) (demoEntry, error) {
	cppImg, err := loadPNG(cppPath)
	if err != nil {
		return demoEntry{}, fmt.Errorf("loading cpp image: %w", err)
	}

	goImg, err := loadPNG(goPath)
	if err != nil {
		return demoEntry{}, fmt.Errorf("loading go image: %w", err)
	}

	opts := framework.ComparisonOptions{
		GenerateDiffImage: true,
		IgnoreAlpha:       true,
	}
	result := framework.CompareImages(cppImg, goImg, opts)

	rawDiff := rawSubtractImage(cppImg, goImg)

	// Recolor identical pixels in the amplified diff to green so they are
	// visually distinct from black-difference pixels.
	if result.DiffImage != nil {
		ab := result.DiffImage.Bounds()
		for y := ab.Min.Y; y < ab.Max.Y; y++ {
			for x := ab.Min.X; x < ab.Max.X; x++ {
				p := result.DiffImage.RGBAAt(x, y)
				if p.R == 0 && p.G == p.B { // identical pixel: gray tint set by framework
					result.DiffImage.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				}
			}
		}
	}

	// RMSE in [0,255] scale
	bounds := cppImg.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	total := w * h
	var sumSq float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rc := cppImg.At(bounds.Min.X+x, bounds.Min.Y+y)
			gc := goImg.At(bounds.Min.X+x, bounds.Min.Y+y)
			rr, rg, rb, _ := rc.RGBA()
			gr, gg, gb, _ := gc.RGBA()
			dr := float64(int(rr>>8) - int(gr>>8))
			dg := float64(int(rg>>8) - int(gg>>8))
			db := float64(int(rb>>8) - int(gb>>8))
			sumSq += dr*dr + dg*dg + db*db
		}
	}
	rmse := 0.0
	if total > 0 {
		meanSq := sumSq / float64(total*3)
		rmse = math.Sqrt(meanSq)
	}

	cppB64, err := pngToBase64(cppImg)
	if err != nil {
		return demoEntry{}, fmt.Errorf("encoding cpp image: %w", err)
	}
	goB64, err := pngToBase64(goImg)
	if err != nil {
		return demoEntry{}, fmt.Errorf("encoding go image: %w", err)
	}
	rawB64, err := pngToBase64(rawDiff)
	if err != nil {
		return demoEntry{}, fmt.Errorf("encoding raw diff image: %w", err)
	}

	var ampB64 string
	if result.DiffImage != nil {
		ampB64, err = pngToBase64(result.DiffImage)
		if err != nil {
			return demoEntry{}, fmt.Errorf("encoding amp diff image: %w", err)
		}
	}

	return demoEntry{
		Name:        name,
		RMSE:        rmse,
		AvgDiff:     result.AverageDifference,
		MaxDiff:     result.MaxDifference,
		DiffPixels:  result.DifferentPixels,
		TotalPixels: result.TotalPixels,
		DiffRatio:   result.DifferentRatio,
		CppB64:      cppB64,
		GoB64:       goB64,
		RawDiffB64:  rawB64,
		AmpDiffB64:  ampB64,
	}, nil
}

func loadDemos(baseDir string) ([]demoEntry, error) {
	cppFull := filepath.Join(baseDir, cppDir)
	goFull := filepath.Join(baseDir, goDir)

	entries, err := os.ReadDir(cppFull)
	if err != nil {
		return nil, fmt.Errorf("reading cpp dir %s: %w", cppFull, err)
	}

	var demos []demoEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".png") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".png")
		cppPath := filepath.Join(cppFull, e.Name())
		goPath := filepath.Join(goFull, e.Name())

		if _, err := os.Stat(goPath); err != nil {
			log.Printf("warning: no Go reference for %s, skipping: %v", name, err)
			continue
		}

		entry, err := buildEntry(name, cppPath, goPath)
		if err != nil {
			log.Printf("warning: failed to build entry for %s: %v", name, err)
			continue
		}
		demos = append(demos, entry)
	}

	sort.Slice(demos, func(i, j int) bool {
		return demos[i].RMSE > demos[j].RMSE
	})

	return demos, nil
}

func regenerateGoReference(ctx context.Context, baseDir string, demo demoConfig) error {
	outDir := filepath.Join(baseDir, goDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create go reference dir: %w", err)
	}

	if err := tryGenerateFromDir(ctx, outDir, demo, filepath.Join(baseDir, demo.dir), []string{"go", "run", "."}); err == nil {
		return nil
	}
	return tryGenerateFromDir(ctx, outDir, demo, baseDir, []string{"go", "run", "./" + demo.dir})
}

func regenerateAllGoReferences(ctx context.Context, baseDir string) error {
	for _, demo := range demoConfigs {
		if err := regenerateGoReference(ctx, baseDir, demo); err != nil {
			return fmt.Errorf("%s: %w", demo.name, err)
		}
	}
	return nil
}

func tryGenerateFromDir(ctx context.Context, outDir string, demo demoConfig, runDir string, args []string) error {
	stamp, err := os.CreateTemp(runDir, ".demo-stamp-*")
	if err != nil {
		return err
	}
	stampPath := stamp.Name()
	_ = stamp.Close()
	defer os.Remove(stampPath)

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = runDir
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s in %s failed: %w\n%s", strings.Join(args, " "), runDir, err, strings.TrimSpace(string(output)))
	}

	generated, err := findGeneratedPNG(runDir, stampPath)
	if err != nil {
		return fmt.Errorf("find generated png after %s in %s: %w\n%s", strings.Join(args, " "), runDir, err, strings.TrimSpace(string(output)))
	}
	defer os.Remove(generated)

	dstPath := filepath.Join(outDir, demo.name+".png")
	return copyFile(generated, dstPath)
}

func findGeneratedPNG(runDir, stampPath string) (string, error) {
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return "", err
	}
	stampInfo, err := os.Stat(stampPath)
	if err != nil {
		return "", err
	}

	var found []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}
		path := filepath.Join(runDir, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if info.ModTime().After(stampInfo.ModTime()) {
			found = append(found, path)
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("no generated png found")
	}
	sort.Strings(found)
	return found[0], nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

const pageHeader = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>AGG Visual Comparison Viewer</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #111; color: #ddd; font-family: monospace; font-size: 13px; }
.sticky-header {
  position: sticky; top: 0; z-index: 100;
  background: #1a1a1a; border-bottom: 1px solid #333;
  padding: 8px 12px; display: flex; align-items: center; gap: 12px; flex-wrap: wrap;
}
.sticky-header h1 { font-size: 15px; color: #eee; }
.sticky-header input, .sticky-header select {
  background: #222; color: #ddd; border: 1px solid #444; padding: 4px 8px;
  font-family: monospace; font-size: 12px;
}
.regen-button {
  background: #262626; color: #ddd; border: 1px solid #555; padding: 4px 8px;
  font-family: monospace; font-size: 12px; cursor: pointer; border-radius: 3px;
}
.regen-button:hover { background: #303030; border-color: #777; }
.regen-button:disabled { opacity: 0.6; cursor: not-allowed; }
.regen-button.is-regenerating { cursor: wait; }
.regen-form { display: inline-flex; }
#summary { color: #888; font-size: 12px; margin-left: auto; }
.container { padding: 12px; }
.card {
  background: #1a1a1a; border: 1px solid #333; margin-bottom: 8px;
  border-radius: 4px; overflow: hidden;
}
.card-header {
  padding: 8px 12px; cursor: pointer; display: flex; align-items: center; gap: 8px;
  background: #222; user-select: none;
}
.card-header:hover { background: #2a2a2a; }
.card-body { display: none; padding: 10px; }
.card.open .card-body { display: block; }
.card-title { font-size: 13px; color: #eee; flex: 1; }
.right-badges { display: flex; gap: 4px; align-items: center; flex-wrap: wrap; }
.badge { padding: 2px 7px; border-radius: 3px; font-size: 11px; font-weight: bold; }
.badge-neutral { background: #2a2a2a; color: #ccc; border: 1px solid #555; }
.badge-ok  { background: #1a3a1a; color: #5f5; border: 1px solid #3a6a3a; }
.badge-warn { background: #3a2a00; color: #fa0; border: 1px solid #6a5000; }
.badge-bad { background: #3a0000; color: #f55; border: 1px solid #6a0000; }
.img-grid {
  display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 8px;
}
.img-col { display: flex; flex-direction: column; gap: 4px; min-width: 0; overflow: auto; }
.img-col label { font-size: 11px; color: #888; text-align: center; }
.zoom-surface {
  position: relative; overflow: hidden; width: 100%; max-width: 100%;
  background: #000; cursor: crosshair;
}
.zoom-transform { position: relative; transform-origin: 0 0; will-change: transform; }
.zoom-selection {
  display: none; position: absolute; border: 1px solid #8ab4ff;
  background: rgba(80, 150, 255, 0.18); pointer-events: none; z-index: 4;
}
.comparison-image {
  display: block; image-rendering: auto; width: 100%; height: auto; max-width: 100%;
}
.resample-pixelated .comparison-image { image-rendering: pixelated; }
.original-size .img-grid { grid-template-columns: repeat(5, max-content); }
.original-size .img-col { min-width: max-content; }
.original-size .zoom-surface, .original-size .slider-wrap { align-self: flex-start; width: auto; }
.original-size .comparison-image { width: auto; height: auto; max-width: none; }
.col-raw { display: none; }
/* Slider */
.slider-wrap {
  position: relative; overflow: hidden; width: 100%; cursor: crosshair;
}
.slider-wrap img.base, .slider-wrap .zoom-overlay-layer img {
  display: block; image-rendering: auto; width: 100%; height: auto; max-width: 100%;
}
.resample-pixelated .slider-wrap img.base,
.resample-pixelated .slider-wrap .zoom-overlay-layer img { image-rendering: pixelated; }
.original-size .slider-wrap img.base,
.original-size .slider-wrap .zoom-overlay-layer img { width: auto; height: auto; max-width: none; }
.slider-overlay {
  position: absolute; top: 0; left: 0; height: 100%; overflow: hidden; width: 50%; z-index: 2;
  pointer-events: none;
}
.slider-divider {
  position: absolute; top: 0; left: 50%; height: 100%;
  width: 3px; background: #fff; cursor: col-resize; transform: translateX(-50%); z-index: 3;
}
.slider-divider::before {
  content: ''; position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 20px; height: 20px; background: #fff; border-radius: 50%;
  border: 2px solid #333;
}
.zoom-dragging, .zoom-dragging * { user-select: none; }
</style>
</head>
<body>
<div class="sticky-header">
  <h1>AGG Visual Comparison</h1>
  <input type="text" id="search" placeholder="Search demos…" oninput="filterCards()" style="width:180px">
<select id="sort-select" onchange="sortCards()">
    <option value="rmse-desc">Sort: RMSE ↓</option>
    <option value="rmse-asc">Sort: RMSE ↑</option>
    <option value="diff-pixels-desc">Sort: Different pixels ↓</option>
    <option value="diff-pixels-asc">Sort: Different pixels ↑</option>
    <option value="avg-diff-desc">Sort: Avg diff ↓</option>
    <option value="avg-diff-asc">Sort: Avg diff ↑</option>
    <option value="max-diff-desc">Sort: Max diff ↓</option>
    <option value="max-diff-asc">Sort: Max diff ↑</option>
    <option value="name-asc">Sort: Name ↑</option>
  </select>
  <select id="diff-mode" onchange="setDiffMode(this.value)">
    <option value="amp">Diff: Amplified</option>
    <option value="raw">Diff: Raw</option>
    <option value="both">Diff: Both</option>
  </select>
  <select id="resample-mode" onchange="setResampleMode(this.value)">
    <option value="smooth">Scaling: Antialiased</option>
    <option value="pixelated">Scaling: Pixelated</option>
  </select>
  <form class="regen-form" method="post" action="/regenerate-all">
    <button class="regen-button" type="submit">Regenerate All</button>
  </form>
  <label style="font-size:12px;display:flex;align-items:center;gap:4px;cursor:pointer">
    <input type="checkbox" id="original-size" onchange="setOriginalSize(this.checked)"> Original size
  </label>
  <span id="summary"></span>
</div>
<div class="container" id="cards-container">
`

const pageFooter = `</div>
<script>
(function() {
  var viewerStateStorageKey = 'agg-visual-viewer-state-v1';
  var viewerStateControlIDs = ['search', 'sort-select', 'diff-mode', 'resample-mode', 'original-size'];
  var minZoomDragPixels = 4;
  var activeSlider = null;
  var activeZoomSelection = null;

  if ('scrollRestoration' in window.history) {
    window.history.scrollRestoration = 'manual';
  }

  function cardStateKey(card) {
    return card.dataset.name || '';
  }

  function loadViewerState() {
    try {
      var raw = window.sessionStorage.getItem(viewerStateStorageKey);
      if (!raw) return {};
      var parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== 'object') return {};
      return parsed;
    } catch (err) {
      return {};
    }
  }

  function saveViewerState() {
    var state = {};
    viewerStateControlIDs.forEach(function(id) {
      var el = document.getElementById(id);
      if (!el) return;
      state[id] = el.type === 'checkbox' ? el.checked : el.value;
    });
    state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);
    state.scrollY = window.scrollY || window.pageYOffset || 0;
    try {
      window.sessionStorage.setItem(viewerStateStorageKey, JSON.stringify(state));
    } catch (err) {
    }
  }

  function restoreViewerState() {
    var state = loadViewerState();
    viewerStateControlIDs.forEach(function(id) {
      if (!Object.prototype.hasOwnProperty.call(state, id)) return;
      var el = document.getElementById(id);
      if (!el) return;
      if (el.type === 'checkbox') {
        el.checked = !!state[id];
        return;
      }
      if (typeof state[id] === 'string') {
        el.value = state[id];
      }
    });
    var openCards = Array.isArray(state.openCards) ? state.openCards : [];
    var openCardSet = new Set(openCards);
    document.querySelectorAll('.card').forEach(function(card) {
      card.classList.toggle('open', openCardSet.has(cardStateKey(card)));
    });
  }

  function restoreScrollPosition() {
    var state = loadViewerState();
    if (typeof state.scrollY !== 'number') {
      return;
    }
    var targetY = Math.max(0, state.scrollY);
    function scrollToTarget() {
      window.scrollTo(0, targetY);
    }
    requestAnimationFrame(function() {
      scrollToTarget();
      setTimeout(scrollToTarget, 0);
    });
    window.addEventListener('load', scrollToTarget, { once: true });
  }

  function bindViewerStatePersistence() {
    viewerStateControlIDs.forEach(function(id) {
      var el = document.getElementById(id);
      if (!el) return;
      el.addEventListener(el.type === 'text' ? 'input' : 'change', saveViewerState);
    });
    window.addEventListener('beforeunload', saveViewerState);
  }

  function clamp(value, minValue, maxValue) {
    return Math.min(Math.max(value, minValue), maxValue);
  }

  // Card toggle
  document.querySelectorAll('.card-header').forEach(function(h) {
    h.addEventListener('click', function() {
      h.closest('.card').classList.toggle('open');
      saveViewerState();
    });
  });
  document.querySelectorAll('.regen-form').forEach(function(form) {
    form.addEventListener('click', function(e) {
      e.stopPropagation();
    });
  });

  function setRegenerateButtonsDisabled(disabled) {
    document.querySelectorAll('.regen-button').forEach(function(button) {
      button.disabled = disabled;
      button.classList.toggle('is-regenerating', disabled);
    });
  }

  function navigateToFreshPage() {
    saveViewerState();
    var url = new URL(window.location.href);
    url.searchParams.set('_vv', String(Date.now()));
    window.location.assign(url.toString());
  }

  function regenerateFromForm(form) {
    return fetch(form.action, {
      method: (form.method || 'POST').toUpperCase(),
      headers: { 'X-Visual-Viewer-Regenerate': '1' }
    }).then(function(response) {
      if (!response.ok) {
        return response.text().then(function(text) {
          throw new Error(text.trim() || 'regenerate failed');
        });
      }
    });
  }

  document.querySelectorAll('.regen-form').forEach(function(form) {
    form.addEventListener('submit', function(e) {
      e.preventDefault();
      e.stopPropagation();
      saveViewerState();
      setRegenerateButtonsDisabled(true);
      regenerateFromForm(form).then(function() {
        navigateToFreshPage();
      }).catch(function(err) {
        window.alert(err.message);
        setRegenerateButtonsDisabled(false);
      });
    });
  });

  function filterCards() {
    var q = document.getElementById('search').value.toLowerCase();
    var cards = document.querySelectorAll('.card');
    cards.forEach(function(c) {
      var name = (c.dataset.name || '').toLowerCase();
      c.style.display = name.includes(q) ? '' : 'none';
    });
    updateSummary();
  }

  function sortCards() {
    var mode = document.getElementById('sort-select').value;
    var container = document.getElementById('cards-container');
    var cards = Array.from(container.querySelectorAll('.card'));
    function metric(card, attr) {
      return parseFloat(card.dataset[attr] || 0);
    }
    cards.sort(function(a, b) {
      if (mode === 'rmse-desc') return metric(b, 'rmse') - metric(a, 'rmse');
      if (mode === 'rmse-asc') return metric(a, 'rmse') - metric(b, 'rmse');
      if (mode === 'diff-pixels-desc') return metric(b, 'diffPixels') - metric(a, 'diffPixels');
      if (mode === 'diff-pixels-asc') return metric(a, 'diffPixels') - metric(b, 'diffPixels');
      if (mode === 'avg-diff-desc') return metric(b, 'avgDiff') - metric(a, 'avgDiff');
      if (mode === 'avg-diff-asc') return metric(a, 'avgDiff') - metric(b, 'avgDiff');
      if (mode === 'max-diff-desc') return metric(b, 'maxDiff') - metric(a, 'maxDiff');
      if (mode === 'max-diff-asc') return metric(a, 'maxDiff') - metric(b, 'maxDiff');
      if (mode === 'name-asc') return (a.dataset.name||'').localeCompare(b.dataset.name||'');
      return 0;
    });
    cards.forEach(function(c) { container.appendChild(c); });
    updateSortMetricBadges(mode);
  }

  function badgeColorClass(attr, value, diffRatio) {
    if (attr === 'rmse') {
      if (value <= 5) return 'badge-ok';
      if (value <= 20) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'avgDiff') {
      if (value <= 2) return 'badge-ok';
      if (value <= 8) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'maxDiff') {
      if (value <= 10) return 'badge-ok';
      if (value <= 40) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'diffPixels') {
      var r = parseFloat(diffRatio || 0);
      if (r <= 0.01) return 'badge-ok';
      if (r <= 0.05) return 'badge-warn';
      return 'badge-bad';
    }
    return 'badge-neutral';
  }

  function updateSortMetricBadges(mode) {
    var label = '';
    var attr = '';
    var formatter = function(value) { return String(value); };

    if (mode === 'rmse-desc' || mode === 'rmse-asc') {
      label = 'RMSE';
      attr = 'rmse';
      formatter = function(value) { return Number(value).toFixed(2); };
    } else if (mode === 'diff-pixels-desc' || mode === 'diff-pixels-asc') {
      label = 'diff px';
      attr = 'diffPixels';
      formatter = function(value) { return String(Math.round(Number(value))); };
    } else if (mode === 'avg-diff-desc' || mode === 'avg-diff-asc') {
      label = 'avg';
      attr = 'avgDiff';
      formatter = function(value) { return Number(value).toFixed(2); };
    } else if (mode === 'max-diff-desc' || mode === 'max-diff-asc') {
      label = 'max';
      attr = 'maxDiff';
      formatter = function(value) { return String(Math.round(Number(value))); };
    }

    document.querySelectorAll('.card').forEach(function(card) {
      var leftBadge = card.querySelector('.sort-metric-badge');
      if (!leftBadge) return;

      if (!attr) {
        leftBadge.style.display = 'none';
        return;
      }

      var value = parseFloat(card.dataset[attr] || 0);
      var colorClass = badgeColorClass(attr, value, card.dataset.diffRatio);
      leftBadge.className = 'badge ' + colorClass + ' sort-metric-badge';
      leftBadge.textContent = label + ' ' + formatter(value);
      leftBadge.style.display = '';
    });
  }

  function setDiffMode(mode) {
    document.querySelectorAll('.col-amp').forEach(function(el) {
      el.style.display = (mode === 'amp' || mode === 'both') ? 'flex' : 'none';
    });
    document.querySelectorAll('.col-raw').forEach(function(el) {
      el.style.display = (mode === 'raw' || mode === 'both') ? 'flex' : 'none';
    });
    refreshCardZooms();
  }

  function updateSummary() {
    var all = document.querySelectorAll('.card');
    var visible = Array.from(all).filter(function(c) { return c.style.display !== 'none'; });
    document.getElementById('summary').textContent = visible.length + ' / ' + all.length + ' demos';
  }

  function setResampleMode(mode) {
    var container = document.getElementById('cards-container');
    container.classList.remove('resample-smooth', 'resample-pixelated');
    container.classList.add(mode === 'pixelated' ? 'resample-pixelated' : 'resample-smooth');
  }

  function setOriginalSize(on) {
    var container = document.getElementById('cards-container');
    if (on) {
      container.classList.add('original-size');
    } else {
      container.classList.remove('original-size');
    }
    refreshCardZooms();
  }

  function ensureCardZoomState(card) {
    if (!card.__zoomState) {
      card.__zoomState = { scale: 1, x: 0, y: 0 };
    }
    return card.__zoomState;
  }

  function applyCardZoom(card) {
    var state = ensureCardZoomState(card);
    card.querySelectorAll('.zoom-surface').forEach(function(surface) {
      var width = surface.clientWidth;
      var height = surface.clientHeight;
      surface.querySelectorAll('.zoom-transform').forEach(function(layer) {
        if (!width || !height || state.scale <= 1) {
          layer.style.transform = '';
          return;
        }
        var tx = -state.x * width * state.scale;
        var ty = -state.y * height * state.scale;
        layer.style.transform = 'matrix(' + state.scale + ',0,0,' + state.scale + ',' + tx + ',' + ty + ')';
      });
    });
  }

  function refreshCardZooms() {
    document.querySelectorAll('.card').forEach(applyCardZoom);
  }

  function setCardZoomFromSelection(card, rect) {
    var width = clamp(rect.width, 0, 1);
    var height = clamp(rect.height, 0, 1);
    if (width <= 0 || height <= 0) {
      return;
    }
    var scale = 1 / Math.max(width, height);
    var visibleWidth = 1 / scale;
    var visibleHeight = 1 / scale;
    var centerX = rect.x + width / 2;
    var centerY = rect.y + height / 2;
    var state = ensureCardZoomState(card);
    state.scale = scale;
    state.x = clamp(centerX - visibleWidth / 2, 0, Math.max(0, 1 - visibleWidth));
    state.y = clamp(centerY - visibleHeight / 2, 0, Math.max(0, 1 - visibleHeight));
    applyCardZoom(card);
  }

  function resetCardZoom(card) {
    var state = ensureCardZoomState(card);
    state.scale = 1;
    state.x = 0;
    state.y = 0;
    applyCardZoom(card);
  }

  function showSelectionBox(selection, x0, y0, x1, y1) {
    var left = Math.min(x0, x1);
    var top = Math.min(y0, y1);
    selection.style.display = 'block';
    selection.style.left = left + 'px';
    selection.style.top = top + 'px';
    selection.style.width = Math.abs(x1 - x0) + 'px';
    selection.style.height = Math.abs(y1 - y0) + 'px';
  }

  function hideSelectionBox(selection) {
    if (!selection) return;
    selection.style.display = 'none';
    selection.style.left = '0';
    selection.style.top = '0';
    selection.style.width = '0';
    selection.style.height = '0';
  }

  // Slider logic
  document.querySelectorAll('.slider-wrap').forEach(function(wrap) {
    var divider = wrap.querySelector('.slider-divider');
    var overlay = wrap.querySelector('.slider-overlay');

    function applyPos(pct) {
      pct = Math.max(0, Math.min(1, pct));
      overlay.style.width = (pct * 100) + '%';
      divider.style.left = (pct * 100) + '%';
      var overlayLayer = wrap.querySelector('.zoom-overlay-layer');
      if (overlayLayer) {
        overlayLayer.style.width = pct <= 0 ? '0%' : (100 / pct) + '%';
      }
    }

    function setPos(x) {
      var rect = wrap.getBoundingClientRect();
      if (rect.width <= 0) {
        applyPos(0.5);
      } else {
        applyPos((x - rect.left) / rect.width);
      }
    }

    wrap.__setSliderClientX = setPos;
    divider.addEventListener('mousedown', function(e) {
      activeSlider = { setPos: setPos };
      e.preventDefault();
      e.stopPropagation();
    });
    applyPos(0.5);
  });

  document.addEventListener('mousemove', function(e) {
    if (activeSlider) {
      activeSlider.setPos(e.clientX);
    }
    if (!activeZoomSelection) {
      return;
    }
    var rect = activeZoomSelection.rect;
    var x = clamp(e.clientX - rect.left, 0, rect.width);
    var y = clamp(e.clientY - rect.top, 0, rect.height);
    activeZoomSelection.currentX = x;
    activeZoomSelection.currentY = y;
    showSelectionBox(activeZoomSelection.selection, activeZoomSelection.startX, activeZoomSelection.startY, x, y);
  });

  document.addEventListener('mouseup', function(e) {
    if (activeSlider && e.button === 0) {
      activeSlider = null;
    }
    if (!activeZoomSelection || e.button !== 0) {
      return;
    }
    var drag = activeZoomSelection;
    var width = Math.abs(drag.currentX - drag.startX);
    var height = Math.abs(drag.currentY - drag.startY);
    hideSelectionBox(drag.selection);
    document.body.classList.remove('zoom-dragging');
    activeZoomSelection = null;

    if (width > minZoomDragPixels && height > minZoomDragPixels) {
      setCardZoomFromSelection(drag.card, {
        x: Math.min(drag.startX, drag.currentX) / drag.rect.width,
        y: Math.min(drag.startY, drag.currentY) / drag.rect.height,
        width: width / drag.rect.width,
        height: height / drag.rect.height,
      });
      return;
    }

    if (drag.surface.classList.contains('slider-wrap') && typeof drag.surface.__setSliderClientX === 'function') {
      drag.surface.__setSliderClientX(e.clientX);
    }
  });

  document.querySelectorAll('.zoom-surface').forEach(function(surface) {
    surface.addEventListener('mousedown', function(e) {
      if (e.button !== 0) {
        return;
      }
      if (e.target.closest('.slider-divider')) {
        return;
      }
      var rect = surface.getBoundingClientRect();
      if (rect.width <= 0 || rect.height <= 0) {
        return;
      }
      var selection = surface.querySelector('.zoom-selection');
      if (!selection) {
        return;
      }
      var startX = clamp(e.clientX - rect.left, 0, rect.width);
      var startY = clamp(e.clientY - rect.top, 0, rect.height);
      activeZoomSelection = {
        card: surface.closest('.card'),
        surface: surface,
        rect: rect,
        selection: selection,
        startX: startX,
        startY: startY,
        currentX: startX,
        currentY: startY,
      };
      showSelectionBox(selection, startX, startY, startX, startY);
      document.body.classList.add('zoom-dragging');
      e.preventDefault();
    });
    surface.addEventListener('contextmenu', function(e) {
      e.preventDefault();
      resetCardZoom(surface.closest('.card'));
    });
  });

  window.addEventListener('resize', refreshCardZooms);

  restoreViewerState();
  bindViewerStatePersistence();
  sortCards();
  setDiffMode(document.getElementById('diff-mode').value);
  setResampleMode(document.getElementById('resample-mode').value);
  setOriginalSize(document.getElementById('original-size').checked);
  refreshCardZooms();
  filterCards();
  restoreScrollPosition();

  // Expose for onchange handlers
  window.filterCards = filterCards;
  window.sortCards = sortCards;
  window.setDiffMode = setDiffMode;
  window.setOriginalSize = setOriginalSize;
  window.setResampleMode = setResampleMode;
})();
</script>
</body>
</html>
`

func badgeClass(rmse float64) string {
	if rmse <= 5 {
		return "badge-ok"
	}
	if rmse <= 20 {
		return "badge-warn"
	}
	return "badge-bad"
}

func badgeClassAvgDiff(v float64) string {
	if v <= 2 {
		return "badge-ok"
	}
	if v <= 8 {
		return "badge-warn"
	}
	return "badge-bad"
}

func badgeClassMaxDiff(v uint8) string {
	if v <= 10 {
		return "badge-ok"
	}
	if v <= 40 {
		return "badge-warn"
	}
	return "badge-bad"
}

func badgeClassDiffRatio(r float64) string {
	if r <= 0.01 {
		return "badge-ok"
	}
	if r <= 0.05 {
		return "badge-warn"
	}
	return "badge-bad"
}

func renderCard(w io.Writer, d *demoEntry) {
	pctDiff := d.DiffRatio * 100.0

	fmt.Fprintf(w, `<div class="card" data-name="%s" data-rmse="%.4f" data-avg-diff="%.4f" data-max-diff="%d" data-diff-pixels="%d" data-diff-ratio="%.6f">`, d.Name, d.RMSE, d.AvgDiff, d.MaxDiff, d.DiffPixels, d.DiffRatio)
	fmt.Fprintf(w, `<div class="card-header">`)
	fmt.Fprintf(w, `<span class="badge badge-neutral sort-metric-badge" style="display:none"></span>`)
	fmt.Fprintf(w, `<span class="card-title">%s</span>`, d.Name)
	fmt.Fprintf(w, `<div class="right-badges">`)
	fmt.Fprintf(w, `<span class="badge %s rmse-badge">RMSE %.2f</span>`, badgeClass(d.RMSE), d.RMSE)
	fmt.Fprintf(w, `<span class="badge %s">avg %.2f</span>`, badgeClassAvgDiff(d.AvgDiff), d.AvgDiff)
	fmt.Fprintf(w, `<span class="badge %s">max %d</span>`, badgeClassMaxDiff(d.MaxDiff), d.MaxDiff)
	fmt.Fprintf(w, `<span class="badge %s">diff %.2f%%</span>`, badgeClassDiffRatio(d.DiffRatio), pctDiff)
	fmt.Fprintf(w, `<form class="regen-form" method="post" action="/regenerate?name=%s">`, d.Name)
	fmt.Fprintf(w, `<button class="regen-button" type="submit">Regenerate</button>`)
	fmt.Fprintf(w, `</form>`)
	fmt.Fprintf(w, `</div>`)
	fmt.Fprintf(w, `</div>`) // card-header

	fmt.Fprintf(w, `<div class="card-body">`)
	fmt.Fprintf(w, `<div class="img-grid">`)

	// Column 1: C++ Reference
	fmt.Fprintf(w, `<div class="img-col">`)
	fmt.Fprintf(w, `<label>C++ Reference</label>`)
	fmt.Fprintf(w, `<div class="zoom-surface">`)
	fmt.Fprintf(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="comparison-image" src="data:image/png;base64,%s" alt="cpp">`, d.CppB64)
	fmt.Fprintf(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprintf(w, `</div>`)

	// Column 2: Go Output
	fmt.Fprintf(w, `<div class="img-col">`)
	fmt.Fprintf(w, `<label>Go Output</label>`)
	fmt.Fprintf(w, `<div class="zoom-surface">`)
	fmt.Fprintf(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="comparison-image" src="data:image/png;base64,%s" alt="go">`, d.GoB64)
	fmt.Fprintf(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprintf(w, `</div>`)

	// Column 3: Slider comparison (C++ base, Go overlay)
	fmt.Fprintf(w, `<div class="img-col">`)
	fmt.Fprintf(w, `<label>C++ vs Go (drag to compare)</label>`)
	fmt.Fprintf(w, `<div class="slider-wrap zoom-surface">`)
	fmt.Fprintf(w, `<div class="zoom-transform zoom-base-layer">`)
	fmt.Fprintf(w, `<img class="base" src="data:image/png;base64,%s" alt="base">`, d.CppB64)
	fmt.Fprintf(w, `</div>`)
	fmt.Fprintf(w, `<div class="slider-overlay">`)
	fmt.Fprintf(w, `<div class="zoom-transform zoom-overlay-layer">`)
	fmt.Fprintf(w, `<img src="data:image/png;base64,%s" alt="overlay">`, d.GoB64)
	fmt.Fprintf(w, `</div></div>`)
	fmt.Fprintf(w, `<div class="slider-divider"></div>`)
	fmt.Fprintf(w, `<div class="zoom-selection"></div>`)
	fmt.Fprintf(w, `</div>`)
	fmt.Fprintf(w, `</div>`)

	// Column 4: Diff (amplified, shown by default)
	fmt.Fprintf(w, `<div class="img-col col-amp">`)
	fmt.Fprintf(w, `<label>Diff (amplified)</label>`)
	if d.AmpDiffB64 != "" {
		fmt.Fprintf(w, `<div class="zoom-surface">`)
		fmt.Fprintf(w, `<div class="zoom-transform">`)
		fmt.Fprintf(w, `<img class="comparison-image" src="data:image/png;base64,%s" alt="amp-diff">`, d.AmpDiffB64)
		fmt.Fprintf(w, `</div><div class="zoom-selection"></div></div>`)
	}
	fmt.Fprintf(w, `</div>`)

	// Column 4b: Diff (raw subtract, hidden by default)
	fmt.Fprintf(w, `<div class="img-col col-raw" style="display:none">`)
	fmt.Fprintf(w, `<label>Diff (raw subtract)</label>`)
	if d.RawDiffB64 != "" {
		fmt.Fprintf(w, `<div class="zoom-surface">`)
		fmt.Fprintf(w, `<div class="zoom-transform">`)
		fmt.Fprintf(w, `<img class="comparison-image" src="data:image/png;base64,%s" alt="raw-diff">`, d.RawDiffB64)
		fmt.Fprintf(w, `</div><div class="zoom-selection"></div></div>`)
	}
	fmt.Fprintf(w, `</div>`)

	fmt.Fprintf(w, `</div>`) // img-grid
	fmt.Fprintf(w, `</div>`) // card-body
	fmt.Fprintf(w, `</div>`) // card
}

func renderPage(w io.Writer, demos []demoEntry) {
	fmt.Fprint(w, pageHeader)
	for i := range demos {
		renderCard(w, &demos[i])
	}
	fmt.Fprint(w, pageFooter)
}

func isRegenerateFetch(r *http.Request) bool {
	return r.Header.Get("X-Visual-Viewer-Regenerate") == "1"
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		demos, err := loadDemos(cwd)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading demos: %v", err), http.StatusInternalServerError)
			return
		}
		renderPage(w, demos)
	})
	http.HandleFunc("/regenerate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		demo, ok := findDemoConfig(name)
		if !ok {
			http.Error(w, fmt.Sprintf("unknown demo %q", name), http.StatusBadRequest)
			return
		}

		regenerateMu.Lock()
		defer regenerateMu.Unlock()

		if err := regenerateGoReference(r.Context(), cwd, demo); err != nil {
			http.Error(w, fmt.Sprintf("failed to regenerate %s: %v", name, err), http.StatusInternalServerError)
			return
		}
		if isRegenerateFetch(r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/regenerate-all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		regenerateMu.Lock()
		defer regenerateMu.Unlock()

		if err := regenerateAllGoReferences(r.Context(), cwd); err != nil {
			http.Error(w, fmt.Sprintf("failed to regenerate all demos: %v", err), http.StatusInternalServerError)
			return
		}
		if isRegenerateFetch(r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	addr := ":" + port
	log.Printf("Visual viewer running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
