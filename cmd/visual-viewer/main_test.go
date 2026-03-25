package main

import (
	"bytes"
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
