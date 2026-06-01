package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderCardEmitsSortableMetrics(t *testing.T) {
	demo := demoEntry{
		Name:       "example",
		RMSE:       12.3456,
		AvgDiff:    3.21,
		MaxDiff:    17,
		DiffPixels: 42,
	}

	var buf bytes.Buffer
	renderCard(&buf, &demo)

	html := buf.String()
	for _, want := range []string{
		`sort-metric-badge`,
		`rmse-badge`,
		`data-rmse="12.3456"`,
		`data-avg-diff="3.2100"`,
		`data-max-diff="17"`,
		`data-diff-pixels="42"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("renderCard() missing %q in output:\n%s", want, html)
		}
	}
}

func TestRenderCardEmitsRegenerateForm(t *testing.T) {
	demo := demoEntry{Name: "line_patterns_clip"}

	var buf bytes.Buffer
	renderCard(&buf, &demo)

	html := buf.String()
	for _, want := range []string{
		`class="regen-form"`,
		`method="post"`,
		`action="/regenerate?name=line_patterns_clip"`,
		`Regenerate`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("renderCard() missing %q in output:\n%s", want, html)
		}
	}
}

func TestRenderCardEmitsZoomSurfaces(t *testing.T) {
	demo := demoEntry{
		Name:       "line_patterns_clip",
		CppB64:     "cpp",
		GoB64:      "go",
		AmpDiffB64: "amp",
		RawDiffB64: "raw",
	}

	var buf bytes.Buffer
	renderCard(&buf, &demo)

	html := buf.String()
	for _, want := range []string{
		`class="zoom-surface"`,
		`class="zoom-transform"`,
		`class="zoom-selection"`,
		`class="comparison-image"`,
		`class="slider-wrap zoom-surface"`,
		`class="zoom-transform zoom-overlay-layer"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("renderCard() missing %q in output:\n%s", want, html)
		}
	}
}

func TestPageHeaderEmitsRegenerateAllForm(t *testing.T) {
	for _, want := range []string{
		`class="regen-form"`,
		`method="post"`,
		`action="/regenerate-all"`,
		`Regenerate All`,
	} {
		if !strings.Contains(pageHeader, want) {
			t.Fatalf("pageHeader missing %q", want)
		}
	}
}

func TestPageInterceptsRegenerateFormsWithFetch(t *testing.T) {
	requiredSnippets := []string{
		`function regenerateFromForm(form) {`,
		`return fetch(form.action, {`,
		`method: (form.method || 'POST').toUpperCase()`,
		`headers: { 'X-Visual-Viewer-Regenerate': '1' }`,
		`form.addEventListener('submit', function(e) {`,
		`e.preventDefault();`,
		`window.alert(err.message);`,
		`setRegenerateButtonsDisabled(false);`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing regenerate fetch snippet %q", snippet)
		}
	}
}

func TestPageUsesCacheBustingNavigationAfterRegenerate(t *testing.T) {
	requiredSnippets := []string{
		`function navigateToFreshPage() {`,
		`saveViewerState();`,
		`url.searchParams.set('_vv', String(Date.now()));`,
		`window.location.assign(url.toString());`,
		`regenerateFromForm(form).then(function() {`,
		`navigateToFreshPage();`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing regenerate freshness snippet %q", snippet)
		}
	}
	if strings.Contains(pageFooter, `window.location.reload();`) {
		t.Fatal("pageFooter still uses plain reload after regenerate")
	}
}

func TestPagePersistsViewerStateAroundRegenerate(t *testing.T) {
	requiredSnippets := []string{
		`var viewerStateStorageKey = 'agg-visual-viewer-state-v1';`,
		`var viewerStateControlIDs = ['search', 'sort-select', 'diff-mode', 'resample-mode', 'original-size'];`,
		`function saveViewerState() {`,
		`state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);`,
		`state.scrollY = window.scrollY || window.pageYOffset || 0;`,
		`window.sessionStorage.setItem(viewerStateStorageKey, JSON.stringify(state));`,
		`function restoreViewerState() {`,
		`card.classList.toggle('open', openCardSet.has(cardStateKey(card)));`,
		`restoreViewerState();`,
		`bindViewerStatePersistence();`,
		`restoreScrollPosition();`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing viewer state snippet %q", snippet)
		}
	}
}

func TestPageRegenerateButtonsSeparateDisabledAndBusyCursors(t *testing.T) {
	requiredSnippets := []string{
		`.regen-button:disabled { opacity: 0.6; cursor: not-allowed; }`,
		`.regen-button.is-regenerating { cursor: wait; }`,
		`button.classList.toggle('is-regenerating', disabled);`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageHeader, snippet) && !strings.Contains(pageFooter, snippet) {
			t.Fatalf("page markup missing regenerate cursor snippet %q", snippet)
		}
	}
	if strings.Contains(pageHeader, `.regen-button:disabled { opacity: 0.6; cursor: wait; }`) {
		t.Fatal("disabled regenerate buttons still use wait cursor")
	}
}

func TestPageHeaderEmitsResampleModeControl(t *testing.T) {
	for _, want := range []string{
		`id="resample-mode"`,
		`setResampleMode(this.value)`,
		`value="smooth"`,
		`value="pixelated"`,
		`Scaling: Antialiased`,
		`Scaling: Pixelated`,
	} {
		if !strings.Contains(pageHeader, want) {
			t.Fatalf("pageHeader missing %q", want)
		}
	}
}

func TestIsRegenerateFetchRequiresHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/regenerate", http.NoBody)
	if isRegenerateFetch(req) {
		t.Fatal("isRegenerateFetch accepted request without fetch header")
	}

	req.Header.Set("X-Visual-Viewer-Regenerate", "1")
	if !isRegenerateFetch(req) {
		t.Fatal("isRegenerateFetch rejected request with fetch header")
	}
}

func TestFindDemoConfigRejectsUnknownName(t *testing.T) {
	if _, ok := findDemoConfig("does_not_exist"); ok {
		t.Fatal("findDemoConfig accepted an unknown demo")
	}
}

func TestFindDemoConfigAcceptsKnownName(t *testing.T) {
	demo, ok := findDemoConfig("line_patterns_clip")
	if !ok {
		t.Fatal("findDemoConfig rejected a known demo")
	}
	if demo.dir != "examples/core/intermediate/line_patterns_clip" {
		t.Fatalf("unexpected demo dir: %s", demo.dir)
	}
}

func TestSortSelectIncludesAdditionalMetrics(t *testing.T) {
	for _, want := range []string{
		`value="diff-pixels-desc"`,
		`value="diff-pixels-asc"`,
		`value="avg-diff-desc"`,
		`value="avg-diff-asc"`,
		`value="max-diff-desc"`,
		`value="max-diff-asc"`,
	} {
		if !strings.Contains(pageHeader, want) {
			t.Fatalf("pageHeader missing %q", want)
		}
	}
}
