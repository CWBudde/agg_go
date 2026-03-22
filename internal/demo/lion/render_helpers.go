package lion

import "github.com/MeKo-Christian/agg_go/internal/color"

// GetColor exposes per-path fill colors for renderer/scanline.RenderAllPaths.
func (ld *LionData) GetColor(index int) color.RGBA8[color.Linear] {
	return ld.Colors[index]
}

// GetPathID exposes per-path path IDs for renderer/scanline.RenderAllPaths.
func (ld *LionData) GetPathID(index int) uint32 {
	return uint32(ld.PathIdx[index])
}

// OpaqueRGBColorView exposes lion colors as opaque RGB8 values for RGB-only
// destination pixfmts such as BGR24.
type OpaqueRGBColorView struct {
	LionData *LionData
}

// GetColor returns the path fill color without alpha.
func (v OpaqueRGBColorView) GetColor(index int) color.RGB8[color.Linear] {
	c := v.LionData.Colors[index]
	return color.RGB8[color.Linear]{R: c.R, G: c.G, B: c.B}
}

// GetPathID forwards the path ID lookup.
func (v OpaqueRGBColorView) GetPathID(index int) uint32 {
	return uint32(v.LionData.PathIdx[index])
}
