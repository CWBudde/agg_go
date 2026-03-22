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
