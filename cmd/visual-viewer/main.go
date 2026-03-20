// Package main provides an HTTP server for visual comparison of AGG rendering outputs.
package main

import (
	"bytes"
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
	"path/filepath"
	"sort"
	"strings"

	"github.com/MeKo-Christian/agg_go/tests/visual/framework"
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
.metrics { color: #999; font-size: 11px; display: flex; gap: 12px; flex-wrap: wrap; }
.badge { padding: 2px 7px; border-radius: 3px; font-size: 11px; font-weight: bold; }
.badge-ok  { background: #1a3a1a; color: #5f5; border: 1px solid #3a6a3a; }
.badge-warn { background: #3a2a00; color: #fa0; border: 1px solid #6a5000; }
.badge-bad { background: #3a0000; color: #f55; border: 1px solid #6a0000; }
.img-grid {
  display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px;
}
.img-col { display: flex; flex-direction: column; gap: 4px; }
.img-col label { font-size: 11px; color: #888; text-align: center; }
.img-col { overflow: auto; }
/* Stretch mode (default): images fill column width with bilinear resampling */
.img-col img { display: block; image-rendering: auto; width: 100%; height: auto; }
/* Original mode: images shown at native pixel size, column scrolls if needed */
.original-size .img-col img { width: auto; height: auto; max-width: none; }
.col-raw { display: none; }
/* Slider */
.slider-wrap {
  position: relative; overflow: hidden; width: 100%; cursor: col-resize;
}
.slider-wrap img.base { display: block; image-rendering: auto; width: 100%; height: auto; }
.original-size .slider-wrap img.base { width: auto; height: auto; max-width: none; }
.slider-overlay {
  position: absolute; top: 0; left: 0; height: 100%; overflow: hidden; width: 50%;
}
.slider-overlay img {
  display: block; position: absolute; top: 0; left: 0;
  image-rendering: auto;
  width: 200%; /* will be updated by JS */
}
.original-size .slider-overlay img { width: auto !important; }
.slider-divider {
  position: absolute; top: 0; left: 50%; height: 100%;
  width: 3px; background: #fff; cursor: col-resize; transform: translateX(-50%);
}
.slider-divider::before {
  content: ''; position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 20px; height: 20px; background: #fff; border-radius: 50%;
  border: 2px solid #333;
}
</style>
</head>
<body>
<div class="sticky-header">
  <h1>AGG Visual Comparison</h1>
  <input type="text" id="search" placeholder="Search demos…" oninput="filterCards()" style="width:180px">
  <select id="sort-select" onchange="sortCards()">
    <option value="rmse-desc">Sort: RMSE ↓</option>
    <option value="rmse-asc">Sort: RMSE ↑</option>
    <option value="name-asc">Sort: Name ↑</option>
  </select>
  <select id="diff-mode" onchange="setDiffMode(this.value)">
    <option value="amp">Diff: Amplified</option>
    <option value="raw">Diff: Raw</option>
    <option value="both">Diff: Both</option>
  </select>
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
  // Card toggle
  document.querySelectorAll('.card-header').forEach(function(h) {
    h.addEventListener('click', function() {
      h.closest('.card').classList.toggle('open');
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
    cards.sort(function(a, b) {
      if (mode === 'rmse-desc') return parseFloat(b.dataset.rmse||0) - parseFloat(a.dataset.rmse||0);
      if (mode === 'rmse-asc')  return parseFloat(a.dataset.rmse||0) - parseFloat(b.dataset.rmse||0);
      if (mode === 'name-asc')  return (a.dataset.name||'').localeCompare(b.dataset.name||'');
      return 0;
    });
    cards.forEach(function(c) { container.appendChild(c); });
  }

  function setDiffMode(mode) {
    document.querySelectorAll('.col-amp').forEach(function(el) {
      el.style.display = (mode === 'amp' || mode === 'both') ? '' : 'none';
    });
    document.querySelectorAll('.col-raw').forEach(function(el) {
      el.style.display = (mode === 'raw' || mode === 'both') ? '' : 'none';
    });
  }

  function updateSummary() {
    var all = document.querySelectorAll('.card');
    var visible = Array.from(all).filter(function(c) { return c.style.display !== 'none'; });
    document.getElementById('summary').textContent = visible.length + ' / ' + all.length + ' demos';
  }

  // Slider logic
  document.querySelectorAll('.slider-wrap').forEach(function(wrap) {
    var divider = wrap.querySelector('.slider-divider');
    var overlay = wrap.querySelector('.slider-overlay');
    var dragging = false;

    function updateSlider(x) {
      var rect = wrap.getBoundingClientRect();
      var pct = Math.max(0, Math.min(1, (x - rect.left) / rect.width));
      overlay.style.width = (pct * 100) + '%';
      divider.style.left = (pct * 100) + '%';
      if (pct > 0) {
        overlay.querySelector('img').style.width = (100 / pct) + '%';
      } else {
        overlay.querySelector('img').style.width = '100%';
      }
    }

    divider.addEventListener('mousedown', function(e) {
      dragging = true;
      e.preventDefault();
    });
    document.addEventListener('mousemove', function(e) {
      if (dragging) updateSlider(e.clientX);
    });
    document.addEventListener('mouseup', function() { dragging = false; });
    wrap.addEventListener('click', function(e) { updateSlider(e.clientX); });
  });

  function setOriginalSize(on) {
    var container = document.getElementById('cards-container');
    if (on) {
      container.classList.add('original-size');
    } else {
      container.classList.remove('original-size');
    }
  }

  // Initial summary
  updateSummary();

  // Expose for onchange handlers
  window.filterCards = filterCards;
  window.sortCards = sortCards;
  window.setDiffMode = setDiffMode;
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

func renderCard(w io.Writer, d *demoEntry) {
	badge := badgeClass(d.RMSE)
	pctDiff := d.DiffRatio * 100.0

	fmt.Fprintf(w, `<div class="card" data-name="%s" data-rmse="%.4f">`, d.Name, d.RMSE)
	fmt.Fprintf(w, `<div class="card-header">`)
	fmt.Fprintf(w, `<span class="badge %s">RMSE %.2f</span>`, badge, d.RMSE)
	fmt.Fprintf(w, `<span class="card-title">%s</span>`, d.Name)
	fmt.Fprintf(w, `<span class="metrics">`)
	fmt.Fprintf(w, `<span>avg&nbsp;diff:&nbsp;%.2f</span>`, d.AvgDiff)
	fmt.Fprintf(w, `<span>max&nbsp;diff:&nbsp;%d</span>`, d.MaxDiff)
	fmt.Fprintf(w, `<span>diff&nbsp;pixels:&nbsp;%.2f%%</span>`, pctDiff)
	fmt.Fprintf(w, `</span>`)
	fmt.Fprintf(w, `</div>`) // card-header

	fmt.Fprintf(w, `<div class="card-body">`)
	fmt.Fprintf(w, `<div class="img-grid">`)

	// Column 1: C++ Reference
	fmt.Fprintf(w, `<div class="img-col">`)
	fmt.Fprintf(w, `<label>C++ Reference</label>`)
	fmt.Fprintf(w, `<img src="data:image/png;base64,%s" alt="cpp">`, d.CppB64)
	fmt.Fprintf(w, `</div>`)

	// Column 2: Go Output
	fmt.Fprintf(w, `<div class="img-col">`)
	fmt.Fprintf(w, `<label>Go Output</label>`)
	fmt.Fprintf(w, `<img src="data:image/png;base64,%s" alt="go">`, d.GoB64)
	fmt.Fprintf(w, `</div>`)

	// Column 3: Slider comparison (C++ base, Go overlay)
	fmt.Fprintf(w, `<div class="img-col">`)
	fmt.Fprintf(w, `<label>C++ vs Go (drag to compare)</label>`)
	fmt.Fprintf(w, `<div class="slider-wrap">`)
	fmt.Fprintf(w, `<img class="base" src="data:image/png;base64,%s" alt="base">`, d.CppB64)
	fmt.Fprintf(w, `<div class="slider-overlay"><img src="data:image/png;base64,%s" alt="overlay"></div>`, d.GoB64)
	fmt.Fprintf(w, `<div class="slider-divider"></div>`)
	fmt.Fprintf(w, `</div>`)
	fmt.Fprintf(w, `</div>`)

	// Column 4: Diff (amplified, shown by default)
	fmt.Fprintf(w, `<div class="img-col col-amp">`)
	fmt.Fprintf(w, `<label>Diff (amplified)</label>`)
	if d.AmpDiffB64 != "" {
		fmt.Fprintf(w, `<img src="data:image/png;base64,%s" alt="amp-diff">`, d.AmpDiffB64)
	}
	fmt.Fprintf(w, `</div>`)

	// Column 4b: Diff (raw subtract, hidden by default)
	fmt.Fprintf(w, `<div class="img-col col-raw" style="display:none">`)
	fmt.Fprintf(w, `<label>Diff (raw subtract)</label>`)
	if d.RawDiffB64 != "" {
		fmt.Fprintf(w, `<img src="data:image/png;base64,%s" alt="raw-diff">`, d.RawDiffB64)
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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		demos, err := loadDemos(cwd)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading demos: %v", err), http.StatusInternalServerError)
			return
		}
		renderPage(w, demos)
	})

	addr := ":" + port
	log.Printf("Visual viewer running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
