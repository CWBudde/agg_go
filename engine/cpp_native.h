#ifndef AGG_GO_ENGINE_CPP_NATIVE_H
#define AGG_GO_ENGINE_CPP_NATIVE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct AggGoCPPImage AggGoCPPImage;
typedef struct AggGoCPPPath AggGoCPPPath;
typedef struct AggGoCPPMatrix AggGoCPPMatrix;
typedef struct AggGoCPPFont AggGoCPPFont;

enum AggGoCPPFillRule {
  AggGoCPPFillRuleNonZero = 0,
  AggGoCPPFillRuleEvenOdd = 1,
};

enum AggGoCPPLineCap {
  AggGoCPPLineCapButt = 0,
  AggGoCPPLineCapRound = 1,
  AggGoCPPLineCapSquare = 2,
};

enum AggGoCPPLineJoin {
  AggGoCPPLineJoinMiter = 0,
  AggGoCPPLineJoinRound = 1,
  AggGoCPPLineJoinBevel = 2,
};

int agg_go_cpp_bridge_probe(void);
int agg_go_cpp_bridge_is_stub(void);
const char* agg_go_cpp_bridge_build_id(void);
const char* agg_go_cpp_bridge_last_error(void);

AggGoCPPImage* agg_go_cpp_image_create(uint32_t width, uint32_t height);
void agg_go_cpp_image_free(AggGoCPPImage* image);
uint32_t agg_go_cpp_image_width(const AggGoCPPImage* image);
uint32_t agg_go_cpp_image_height(const AggGoCPPImage* image);
uint32_t agg_go_cpp_image_stride(const AggGoCPPImage* image);
const uint8_t* agg_go_cpp_image_pixels(const AggGoCPPImage* image);
uint8_t* agg_go_cpp_image_pixels_mut(AggGoCPPImage* image);
int agg_go_cpp_image_clear(AggGoCPPImage* image, uint8_t r, uint8_t g, uint8_t b, uint8_t a);
int agg_go_cpp_image_blit(AggGoCPPImage* dst, const AggGoCPPImage* src, int dst_x, int dst_y,
                          int src_x, int src_y, uint32_t width, uint32_t height);
int agg_go_cpp_image_composite(AggGoCPPImage* dst, const AggGoCPPImage* src, int dst_x, int dst_y,
                               int clip_x1, int clip_y1, int clip_x2, int clip_y2, int blend_mode);
int agg_go_cpp_image_composite_scaled(AggGoCPPImage* dst, const AggGoCPPImage* src, int src_x,
                                      int src_y, uint32_t src_width, uint32_t src_height,
                                      int dst_x, int dst_y, uint32_t dst_width, uint32_t dst_height,
                                      int clip_x1, int clip_y1, int clip_x2, int clip_y2,
                                      int blend_mode);
int agg_go_cpp_image_composite_quad(AggGoCPPImage* dst, const AggGoCPPImage* src, int src_x,
                                    int src_y, uint32_t src_width, uint32_t src_height,
                                    const double* quad_xy, int clip_x1, int clip_y1, int clip_x2,
                                    int clip_y2, int blend_mode);

AggGoCPPPath* agg_go_cpp_path_create(void);
void agg_go_cpp_path_free(AggGoCPPPath* path);
int agg_go_cpp_path_reset(AggGoCPPPath* path);
int agg_go_cpp_path_move_to(AggGoCPPPath* path, float x, float y);
int agg_go_cpp_path_line_to(AggGoCPPPath* path, float x, float y);
int agg_go_cpp_path_close(AggGoCPPPath* path);

AggGoCPPMatrix* agg_go_cpp_matrix_create(void);
void agg_go_cpp_matrix_free(AggGoCPPMatrix* matrix);
int agg_go_cpp_matrix_identity(AggGoCPPMatrix* matrix);
int agg_go_cpp_matrix_translate(AggGoCPPMatrix* matrix, float tx, float ty);
int agg_go_cpp_matrix_scale(AggGoCPPMatrix* matrix, float sx, float sy);
int agg_go_cpp_matrix_rotate(AggGoCPPMatrix* matrix, float angle);
int agg_go_cpp_matrix_transform_point(const AggGoCPPMatrix* matrix, double* x, double* y);
AggGoCPPPath* agg_go_cpp_path_transform(const AggGoCPPPath* path, const AggGoCPPMatrix* matrix);

int agg_go_cpp_render_fill_path(AggGoCPPImage* image, const AggGoCPPPath* path, int fill_rule,
                                uint8_t r, uint8_t g, uint8_t b, uint8_t a);
int agg_go_cpp_render_stroke_path(AggGoCPPImage* image, const AggGoCPPPath* path, float width,
                                  int line_cap, int line_join, float miter_limit, uint8_t r,
                                  uint8_t g, uint8_t b, uint8_t a);
// agg_go_cpp_render_stroke_path_dashed strokes a dashed outline. dashes points
// to dash_pair_count (dash_len, gap_len) pairs (2*dash_pair_count floats);
// dash_start is the phase offset along the path. dash_pair_count == 0 strokes
// solid. Only the real AGG-backed build applies the dash pattern; the stub
// build (never advertised as an available backend) strokes solid.
int agg_go_cpp_render_stroke_path_dashed(AggGoCPPImage* image, const AggGoCPPPath* path, float width,
                                         int line_cap, int line_join, float miter_limit,
                                         const float* dashes, int dash_pair_count, float dash_start,
                                         uint8_t r, uint8_t g, uint8_t b, uint8_t a);
// agg_go_cpp_render_fill_path_comp fills a path directly onto the destination
// through a comp-op-aware pixfmt, applying blend_mode per span with coverage and
// clipping to the half-open [clip_x1,clip_x2) x [clip_y1,clip_y2) rectangle.
int agg_go_cpp_render_fill_path_comp(AggGoCPPImage* image, const AggGoCPPPath* path, int fill_rule,
                                     int blend_mode, int clip_x1, int clip_y1, int clip_x2,
                                     int clip_y2, uint8_t r, uint8_t g, uint8_t b, uint8_t a);
// agg_go_cpp_render_stroke_path_comp strokes (optionally dashed) a path directly
// onto the destination through a comp-op-aware pixfmt with the same blend/clip
// semantics as agg_go_cpp_render_fill_path_comp. dash_pair_count == 0 strokes
// solid.
int agg_go_cpp_render_stroke_path_comp(AggGoCPPImage* image, const AggGoCPPPath* path, float width,
                                       int line_cap, int line_join, float miter_limit,
                                       const float* dashes, int dash_pair_count, float dash_start,
                                       int blend_mode, int clip_x1, int clip_y1, int clip_x2,
                                       int clip_y2, uint8_t r, uint8_t g, uint8_t b, uint8_t a);

AggGoCPPFont* agg_go_cpp_font_create(const char* font_path);
void agg_go_cpp_font_free(AggGoCPPFont* font);
int agg_go_cpp_font_set_size(AggGoCPPFont* font, float size);
int agg_go_cpp_font_set_hinting(AggGoCPPFont* font, int enabled);
int agg_go_cpp_font_set_flip_y(AggGoCPPFont* font, int flip_y);
int agg_go_cpp_render_text(AggGoCPPImage* image, AggGoCPPFont* font, const char* text, float x, float y,
                           uint8_t r, uint8_t g, uint8_t b, uint8_t a);
float agg_go_cpp_text_width(AggGoCPPFont* font, const char* text);
float agg_go_cpp_text_height(AggGoCPPFont* font);

#ifdef __cplusplus
}
#endif

#endif
