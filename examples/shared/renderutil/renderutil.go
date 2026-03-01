package renderutil

import agg "agg_go"

// SavePNG persists an AGG image as a PNG file.
func SavePNG(img *agg.Image, filename string) error {
	return img.SaveToPNG(filename)
}
