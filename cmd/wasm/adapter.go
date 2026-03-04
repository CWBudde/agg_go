package main

import (
	"agg_go/internal/path"
	"agg_go/internal/rasterizer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
)

// pathSourceAdapter bridges PathStorageStl (uint Rewind) to the rasterizer's
// VertexSource interface (uint32 Rewind + pointer-based Vertex).
type pathSourceAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathSourceAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

// rasScanlineAdapter adapts scanline.ScanlineU8 to rasterizer.ScanlineInterface
type rasScanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

// rasterizerAdapter adapts internal rasterizer to renderer/scanline.RasterizerInterface
type rasterizerAdapter struct {
	ras interface {
		RewindScanlines() bool
		SweepScanline(sl rasterizer.ScanlineInterface) bool
		MinX() int
		MaxX() int
	}
}

func (r *rasterizerAdapter) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdapter) MinX() int             { return r.ras.MinX() }
func (r *rasterizerAdapter) MaxX() int             { return r.ras.MaxX() }

func (r *rasterizerAdapter) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapperU8); ok {
		return r.ras.SweepScanline(&rasScanlineAdapter{sl: w.sl})
	}
	if w, ok := sl.(*scanlineWrapperP8); ok {
		return r.ras.SweepScanline(&rasScanlineAdapterP8{sl: w.sl})
	}
	return false
}

// rasScanlineAdapterP8 adapts scanline.ScanlineP8 to rasterizer.ScanlineInterface
type rasScanlineAdapterP8 struct {
	sl *scanline.ScanlineP8
}

func (a *rasScanlineAdapterP8) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapterP8) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapterP8) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapterP8) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapterP8) NumSpans() int  { return a.sl.NumSpans() }

type scanlineWrapperU8 struct{ sl *scanline.ScanlineU8 }

func (w *scanlineWrapperU8) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapperU8) Y() int               { return w.sl.Y() }
func (w *scanlineWrapperU8) NumSpans() int        { return w.sl.NumSpans() }

type spanIterU8 struct {
	spans []scanline.Span
	idx   int
}

func (it *spanIterU8) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIterU8) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapperU8) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIterU8{spans: nil, idx: 0}
	}
	return &spanIterU8{spans: spans, idx: 0}
}

type scanlineWrapperP8 struct{ sl *scanline.ScanlineP8 }

func (w *scanlineWrapperP8) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapperP8) Y() int               { return w.sl.Y() }
func (w *scanlineWrapperP8) NumSpans() int        { return w.sl.NumSpans() }

type spanIterP8 struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *spanIterP8) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIterP8) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapperP8) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIterP8{spans: nil, idx: 0}
	}
	return &spanIterP8{spans: spans, idx: 0}
}
