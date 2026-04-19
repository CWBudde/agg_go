//go:build freetype

package freetype

/*
#cgo pkg-config: freetype2 harfbuzz
#include <ft2build.h>
#include FT_FREETYPE_H
#include <hb.h>
#include <hb-ft.h>
#include <stdlib.h>

typedef struct {
    unsigned int codepoint;
    unsigned int cluster;
    int x_advance;
    int y_advance;
    int x_offset;
    int y_offset;
} agg_hb_glyph_position;

static int agg_hb_shape_text(FT_Face face, const char* text, size_t length, int load_flags,
                             agg_hb_glyph_position** out_glyphs, unsigned int* out_count) {
    hb_font_t* font = hb_ft_font_create_referenced(face);
    hb_buffer_t* buffer = hb_buffer_create();
    agg_hb_glyph_position* glyphs = NULL;

    if (!font || !buffer) {
        if (buffer) {
            hb_buffer_destroy(buffer);
        }
        if (font) {
            hb_font_destroy(font);
        }
        return 0;
    }

    hb_ft_font_set_load_flags(font, load_flags);
    hb_buffer_add_utf8(buffer, text, (int)length, 0, (int)length);
    hb_buffer_guess_segment_properties(buffer);
    hb_shape(font, buffer, NULL, 0);

    unsigned int count = 0;
    hb_glyph_info_t* infos = hb_buffer_get_glyph_infos(buffer, &count);
    hb_glyph_position_t* positions = hb_buffer_get_glyph_positions(buffer, &count);

    if (count > 0) {
        glyphs = (agg_hb_glyph_position*)calloc(count, sizeof(agg_hb_glyph_position));
        if (!glyphs) {
            hb_buffer_destroy(buffer);
            hb_font_destroy(font);
            return 0;
        }
        for (unsigned int i = 0; i < count; i++) {
            glyphs[i].codepoint = infos[i].codepoint;
            glyphs[i].cluster = infos[i].cluster;
            glyphs[i].x_advance = positions[i].x_advance;
            glyphs[i].y_advance = positions[i].y_advance;
            glyphs[i].x_offset = positions[i].x_offset;
            glyphs[i].y_offset = positions[i].y_offset;
        }
    }

    hb_buffer_destroy(buffer);
    hb_font_destroy(font);
    *out_glyphs = glyphs;
    *out_count = count;
    return 1;
}

static void agg_hb_free_shaped_glyphs(agg_hb_glyph_position* glyphs) {
    free(glyphs);
}
*/
import "C"

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/font"
)

// LayoutText shapes text into positioned glyphs using HarfBuzz so raster text
// placement can follow font-provided advances and ligature substitutions rather
// than a simple rune loop.
func (fe *FontEngineFreetype) LayoutText(text string) ([]font.PositionedGlyph, bool) {
	if fe.currentFace == nil || text == "" {
		return nil, false
	}

	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))

	var raw *C.agg_hb_glyph_position
	var count C.uint
	if C.agg_hb_shape_text(
		fe.currentFace,
		cstr,
		C.size_t(len(text)),
		C.int(fe.currentLoadFlags()),
		&raw,
		&count,
	) == 0 {
		return nil, false
	}
	if raw != nil {
		defer C.agg_hb_free_shaped_glyphs(raw)
	}
	if count == 0 {
		return nil, false
	}

	rawSlice := unsafe.Slice(raw, int(count))
	placed := make([]font.PositionedGlyph, 0, int(count))
	for _, glyph := range rawSlice {
		if glyph.codepoint == 0 {
			continue
		}
		placed = append(placed, font.PositionedGlyph{
			GlyphIndex: uint(glyph.codepoint),
			XAdvance:   float64(glyph.x_advance) / 64.0,
			YAdvance:   float64(glyph.y_advance) / 64.0,
			XOffset:    float64(glyph.x_offset) / 64.0,
			YOffset:    float64(glyph.y_offset) / 64.0,
		})
	}
	if len(placed) == 0 {
		return nil, false
	}
	return placed, true
}
