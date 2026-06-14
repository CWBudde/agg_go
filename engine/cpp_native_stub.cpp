//go:build agogo && cgo

#include "cpp_native.h"

#include <algorithm>
#include <cmath>
#include <cstdint>
#include <cstring>
#include <memory>
#include <string>
#include <vector>

#ifdef AGG_GO_CPP_REAL
#include <agg_color_rgba.h>
#include <agg_font_cache_manager.h>
#include <agg_font_freetype.h>
#include <agg_conv_dash.h>
#include <agg_conv_stroke.h>
#include <agg_conv_curve.h>
#include <agg_conv_transform.h>
#include <agg_trans_affine.h>
#include <agg_path_storage.h>
#include <agg_pixfmt_rgba.h>
#include <agg_rasterizer_scanline_aa.h>
#include <agg_renderer_base.h>
#include <agg_renderer_scanline.h>
#include <agg_rendering_buffer.h>
#include <agg_scanline_u.h>
#endif

struct Point {
  float x;
  float y;
};

struct PointD {
  double x;
  double y;
};

struct AggGoCPPImage {
  uint32_t width;
  uint32_t height;
  uint32_t stride;
  std::vector<uint8_t> pixels;
#ifdef AGG_GO_CPP_REAL
  std::unique_ptr<agg::rendering_buffer> rendering_buf;
  std::unique_ptr<agg::pixfmt_rgba32> pixfmt;
  std::unique_ptr<agg::renderer_base<agg::pixfmt_rgba32>> ren_base;
#endif
};

struct AggGoCPPPath {
  std::vector<Point> points;
  bool closed = false;
};

struct AggGoCPPMatrix {
  double a;
  double b;
  double c;
  double d;
  double e;
  double f;
};

struct AggGoCPPFont {
#ifdef AGG_GO_CPP_REAL
  using font_engine_type = agg::font_engine_freetype_int32;
  using font_manager_type = agg::font_cache_manager<font_engine_type>;
  using path_adaptor_type = font_manager_type::path_adaptor_type;
  using conv_curves_type = agg::conv_curve<path_adaptor_type>;

  std::unique_ptr<font_engine_type> font_engine;
  std::unique_ptr<font_manager_type> font_manager;
#endif
  std::string font_path;
  float font_size = 12.0f;
  bool hinting_enabled = true;
  bool flip_y = true;
};

namespace {

// Blend-mode constants mirror agg.BlendMode (internal/agg2d/blend_modes.go).
// kBlendAlpha is the default standard alpha blend (aliases src-over); the rest
// are the Porter-Duff operators and separable blend modes, in enum order.
constexpr int kBlendAlpha = 0;
constexpr int kBlendClear = 1;
constexpr int kBlendSrc = 2;
constexpr int kBlendDst = 3;
constexpr int kBlendSrcOver = 4;
constexpr int kBlendDstOver = 5;
constexpr int kBlendSrcIn = 6;
constexpr int kBlendDstIn = 7;
constexpr int kBlendSrcOut = 8;
constexpr int kBlendDstOut = 9;
constexpr int kBlendSrcAtop = 10;
constexpr int kBlendDstAtop = 11;
constexpr int kBlendXor = 12;
constexpr int kBlendAdd = 13;
constexpr int kBlendMultiply = 14;
constexpr int kBlendScreen = 15;
constexpr int kBlendOverlay = 16;
constexpr int kBlendDarken = 17;
constexpr int kBlendLighten = 18;
constexpr int kBlendColorDodge = 19;
constexpr int kBlendColorBurn = 20;
constexpr int kBlendHardLight = 21;
constexpr int kBlendSoftLight = 22;
constexpr int kBlendDifference = 23;
constexpr int kBlendExclusion = 24;

std::string& last_error_storage() {
  static std::string value = "stub bridge: no AGG-backed implementation linked";
  return value;
}

struct RectI {
  int x1;
  int y1;
  int x2;
  int y2;
};

bool valid_image(const AggGoCPPImage* image) { return image != nullptr; }

bool valid_path(const AggGoCPPPath* path) {
  return path != nullptr && path->points.size() >= 3;
}

bool valid_stroke_path(const AggGoCPPPath* path) {
  return path != nullptr && path->points.size() >= 2;
}

bool valid_matrix(const AggGoCPPMatrix* matrix) { return matrix != nullptr; }
bool valid_font(const AggGoCPPFont* font) { return font != nullptr; }

void set_last_error(const char* value) { last_error_storage() = value; }

#ifdef AGG_GO_CPP_REAL
void init_rendering(AggGoCPPImage* image) {
  image->rendering_buf =
      std::make_unique<agg::rendering_buffer>(image->pixels.data(), image->width, image->height,
                                              static_cast<int>(image->stride));
  image->pixfmt = std::make_unique<agg::pixfmt_rgba32>(*image->rendering_buf);
  image->ren_base = std::make_unique<agg::renderer_base<agg::pixfmt_rgba32>>(*image->pixfmt);
}

void convert_path_to_agg(const AggGoCPPPath& path, agg::path_storage* agg_path) {
  if (path.points.empty()) {
    return;
  }
  agg_path->move_to(path.points[0].x, path.points[0].y);
  for (size_t i = 1; i < path.points.size(); ++i) {
    agg_path->line_to(path.points[i].x, path.points[i].y);
  }
  if (path.closed) {
    agg_path->close_polygon();
  }
}

void setup_font_engine(AggGoCPPFont* font) {
  font->font_engine->height(font->font_size);
  font->font_engine->width(font->font_size);
  font->font_engine->hinting(font->hinting_enabled);
  font->font_engine->flip_y(font->flip_y);
  font->font_engine->char_map(FT_ENCODING_UNICODE);
}

// map_comp_op translates the engine blend_mode enum into the AGG compositing
// operator used by the comp-op pixfmt. The mapping is 1:1 with AGG 2.6's
// comp_op_e (agg_pixfmt_rgba.h), covering every Porter-Duff and separable blend
// operator. kBlendAlpha aliases src-over; kBlendAdd maps to comp_op_plus (AGG's
// comp_op_minus is disabled upstream, so plus is followed directly by multiply).
agg::comp_op_e map_comp_op(int blend_mode) {
  switch (blend_mode) {
    case kBlendClear:
      return agg::comp_op_clear;
    case kBlendSrc:
      return agg::comp_op_src;
    case kBlendDst:
      return agg::comp_op_dst;
    case kBlendDstOver:
      return agg::comp_op_dst_over;
    case kBlendSrcIn:
      return agg::comp_op_src_in;
    case kBlendDstIn:
      return agg::comp_op_dst_in;
    case kBlendSrcOut:
      return agg::comp_op_src_out;
    case kBlendDstOut:
      return agg::comp_op_dst_out;
    case kBlendSrcAtop:
      return agg::comp_op_src_atop;
    case kBlendDstAtop:
      return agg::comp_op_dst_atop;
    case kBlendXor:
      return agg::comp_op_xor;
    case kBlendAdd:
      return agg::comp_op_plus;
    case kBlendMultiply:
      return agg::comp_op_multiply;
    case kBlendScreen:
      return agg::comp_op_screen;
    case kBlendOverlay:
      return agg::comp_op_overlay;
    case kBlendDarken:
      return agg::comp_op_darken;
    case kBlendLighten:
      return agg::comp_op_lighten;
    case kBlendColorDodge:
      return agg::comp_op_color_dodge;
    case kBlendColorBurn:
      return agg::comp_op_color_burn;
    case kBlendHardLight:
      return agg::comp_op_hard_light;
    case kBlendSoftLight:
      return agg::comp_op_soft_light;
    case kBlendDifference:
      return agg::comp_op_difference;
    case kBlendExclusion:
      return agg::comp_op_exclusion;
    case kBlendSrcOver:
    case kBlendAlpha:
    default:
      return agg::comp_op_src_over;
  }
}

// render_with_comp_op renders a configured rasterizer directly onto the
// destination image through a comp-op-aware pixfmt, so the compositing operator
// is applied per span with anti-aliased coverage (matching AGG's Agg2D). This
// is the faithful path that replaces "render to a transparent layer, then
// composite the whole layer": the latter applied the operator across the entire
// clip rectangle, so operators like src/clear wiped the untouched background.
// configure_stroke applies width, cap, join, and miter limit to any conv_stroke
// instantiation, mirroring the inline setup used by the non-comp stroke path.
template <class Stroke>
void configure_stroke(Stroke& stroke, float width, int line_cap, int line_join, float miter_limit) {
  stroke.width(width);
  switch (line_cap) {
    case AggGoCPPLineCapRound:
      stroke.line_cap(agg::round_cap);
      break;
    case AggGoCPPLineCapSquare:
      stroke.line_cap(agg::square_cap);
      break;
    default:
      stroke.line_cap(agg::butt_cap);
      break;
  }
  switch (line_join) {
    case AggGoCPPLineJoinRound:
      stroke.line_join(agg::round_join);
      break;
    case AggGoCPPLineJoinBevel:
      stroke.line_join(agg::bevel_join);
      break;
    default:
      stroke.line_join(agg::miter_join);
      stroke.miter_limit(miter_limit);
      break;
  }
}

#ifdef AGG_GO_CPP_REAL
// matrix_is_identity reports whether a (possibly null) matrix is the identity
// transform. A null or identity matrix means the path is already in device
// space (or no transform is active) and the stroke is rasterized as-is.
bool matrix_is_identity(const AggGoCPPMatrix* matrix) {
  return matrix == nullptr || (matrix->a == 1.0 && matrix->b == 0.0 && matrix->c == 0.0 &&
                               matrix->d == 1.0 && matrix->e == 0.0 && matrix->f == 0.0);
}

// add_stroke_to_ras rasterizes a configured stroke (any conv_stroke
// instantiation, possibly over conv_dash). When matrix is non-identity the
// stroke output is wrapped in conv_transform so the dash period and stroke
// width are measured in user space and the resulting outline is mapped to
// device space last. This is faithful to AGG's Agg2D pipeline
// (path -> dash -> stroke -> transform) and the Go port's addStrokeToRasterizer
// (conv_transform(conv_stroke(...), m)). Pre-transforming the path and stroking
// in device space instead would leave the dash lengths and width unscaled by
// the matrix, diverging from the port under any non-identity transform.
template <class VertexSource>
void add_stroke_to_ras(agg::rasterizer_scanline_aa<>& ras, VertexSource& stroke,
                       const AggGoCPPMatrix* matrix) {
  if (matrix_is_identity(matrix)) {
    ras.add_path(stroke);
    return;
  }
  agg::trans_affine mtx(matrix->a, matrix->b, matrix->c, matrix->d, matrix->e, matrix->f);
  agg::conv_transform<VertexSource, agg::trans_affine> trans(stroke, mtx);
  ras.add_path(trans);
}
#endif

// comp_op_adaptor_rgba_plain bridges AGG's premultiplied compositing math to a
// *straight* (non-premultiplied) destination buffer, matching the Go port's
// CompositeBlenderPlain. AGG's stock comp_op_adaptor_rgba treats the buffer as
// premultiplied: it stores premultiplied results, which read back too dark for
// operators that leave a translucent result (e.g. src with a translucent
// colour). This adaptor premultiplies the straight destination on read, runs the
// operator in premultiplied space, then demultiplies back to straight on write.
// For opaque destinations and opaque results the round-trip is identity, so the
// faithful src-over/clear/solid paths are unaffected.
template <class ColorT, class Order>
struct comp_op_adaptor_rgba_plain {
  typedef ColorT color_type;
  typedef Order order_type;
  typedef typename color_type::value_type value_type;

  static AGG_INLINE void blend_pix(unsigned op, value_type* p, value_type r, value_type g,
                                   value_type b, value_type a, agg::cover_type cover) {
    agg::multiplier_rgba<ColorT, Order>::premultiply(p);
    agg::comp_op_table_rgba<ColorT, Order>::g_comp_op_func[op](
        p, color_type::multiply(r, a), color_type::multiply(g, a), color_type::multiply(b, a), a,
        cover);
    agg::multiplier_rgba<ColorT, Order>::demultiply(p);
  }
};

template <class Rasterizer>
void render_with_comp_op(AggGoCPPImage* image, Rasterizer& ras, int blend_mode, int clip_x1,
                         int clip_y1, int clip_x2, int clip_y2, uint8_t r, uint8_t g, uint8_t b,
                         uint8_t a) {
  typedef comp_op_adaptor_rgba_plain<agg::rgba8, agg::order_rgba> blender_type;
  typedef agg::pixfmt_custom_blend_rgba<blender_type, agg::rendering_buffer> pixfmt_comp_type;
  pixfmt_comp_type pf(*image->rendering_buf);
  pf.comp_op(map_comp_op(blend_mode));
  agg::renderer_base<pixfmt_comp_type> rb(pf);
  rb.clip_box(clip_x1, clip_y1, clip_x2, clip_y2);
  agg::scanline_u8 sl;
  agg::renderer_scanline_aa_solid<agg::renderer_base<pixfmt_comp_type> > ren(rb);
  ren.color(agg::rgba8(r, g, b, a));
  agg::render_scanlines(ras, sl, ren);
}
#endif

bool next_utf8(const char*& p, uint32_t& cp) {
  unsigned char c = static_cast<unsigned char>(*p);
  if (c < 0x80) {
    cp = c;
    ++p;
    return true;
  }
  if ((c >> 5) == 0x6) {
    if (!p[1]) return false;
    cp = ((c & 0x1F) << 6) | (static_cast<unsigned char>(p[1]) & 0x3F);
    p += 2;
    return true;
  }
  if ((c >> 4) == 0xE) {
    if (!p[1] || !p[2]) return false;
    cp = ((c & 0x0F) << 12) | ((static_cast<unsigned char>(p[1]) & 0x3F) << 6) |
         (static_cast<unsigned char>(p[2]) & 0x3F);
    p += 3;
    return true;
  }
  if ((c >> 3) == 0x1E) {
    if (!p[1] || !p[2] || !p[3]) return false;
    cp = ((c & 0x07) << 18) | ((static_cast<unsigned char>(p[1]) & 0x3F) << 12) |
         ((static_cast<unsigned char>(p[2]) & 0x3F) << 6) |
         (static_cast<unsigned char>(p[3]) & 0x3F);
    p += 4;
    return true;
  }
  return false;
}

uint8_t clamp_byte(double value) {
  if (value <= 0.0) {
    return 0;
  }
  if (value >= 255.0) {
    return 255;
  }
  return static_cast<uint8_t>(std::lround(value));
}

// supported_blend_mode gates the composite_pixel paths, which only implement the
// five operators with a direct straight-alpha CPU formula. The whole-rect plain
// composite (the default text-layer blit) and the stub build use it; the scaled/
// quad image blits route through blend_image_pixel and so honour the full operator
// set in the real build (supported_comp_op_mode).
bool supported_blend_mode(int blend_mode) {
  switch (blend_mode) {
    case kBlendAlpha:
    case kBlendClear:
    case kBlendSrc:
    case kBlendDst:
    case kBlendSrcOver:
      return true;
    default:
      return false;
  }
}

// supported_comp_op_mode gates the vector fill/stroke and gradient-coverage
// paths, which render through AGG's comp-op pixfmt (g_comp_op_func[op]) and so
// support every operator map_comp_op knows about. In the stub build those paths
// fall back to a plain solid fill that ignores the operator, so the stub is
// held to the same narrow set it can actually honour.
bool supported_comp_op_mode(int blend_mode) {
#ifdef AGG_GO_CPP_REAL
  return blend_mode >= kBlendAlpha && blend_mode <= kBlendExclusion;
#else
  return supported_blend_mode(blend_mode);
#endif
}

RectI clamp_clip_rect(const AggGoCPPImage* image, int clip_x1, int clip_y1, int clip_x2, int clip_y2) {
  if (clip_x2 < clip_x1) {
    std::swap(clip_x1, clip_x2);
  }
  if (clip_y2 < clip_y1) {
    std::swap(clip_y1, clip_y2);
  }
  return RectI{
      std::max(0, std::min(static_cast<int>(image->width), clip_x1)),
      std::max(0, std::min(static_cast<int>(image->height), clip_y1)),
      std::max(0, std::min(static_cast<int>(image->width), clip_x2)),
      std::max(0, std::min(static_cast<int>(image->height), clip_y2)),
  };
}

bool rect_empty(const RectI& rect) { return rect.x1 >= rect.x2 || rect.y1 >= rect.y2; }

void composite_pixel(uint8_t* dst, const uint8_t* src, int blend_mode) {
  switch (blend_mode) {
    case kBlendClear:
      dst[0] = 0;
      dst[1] = 0;
      dst[2] = 0;
      dst[3] = 0;
      return;
    case kBlendDst:
      return;
    case kBlendSrc:
      dst[0] = src[0];
      dst[1] = src[1];
      dst[2] = src[2];
      dst[3] = src[3];
      return;
    case kBlendAlpha:
    case kBlendSrcOver: {
      const double sr = static_cast<double>(src[0]) / 255.0;
      const double sg = static_cast<double>(src[1]) / 255.0;
      const double sb = static_cast<double>(src[2]) / 255.0;
      const double sa = static_cast<double>(src[3]) / 255.0;
      const double dr = static_cast<double>(dst[0]) / 255.0;
      const double dg = static_cast<double>(dst[1]) / 255.0;
      const double db = static_cast<double>(dst[2]) / 255.0;
      const double da = static_cast<double>(dst[3]) / 255.0;
      const double out_a = sa + da * (1.0 - sa);
      if (out_a <= 1e-12) {
        dst[0] = 0;
        dst[1] = 0;
        dst[2] = 0;
        dst[3] = 0;
        return;
      }
      const double out_r = (sr * sa + dr * da * (1.0 - sa)) / out_a;
      const double out_g = (sg * sa + dg * da * (1.0 - sa)) / out_a;
      const double out_b = (sb * sa + db * da * (1.0 - sa)) / out_a;
      dst[0] = clamp_byte(out_r * 255.0);
      dst[1] = clamp_byte(out_g * 255.0);
      dst[2] = clamp_byte(out_b * 255.0);
      dst[3] = clamp_byte(out_a * 255.0);
      return;
    }
    default:
      return;
  }
}

// blend_image_pixel composites one straight-RGBA source pixel onto the straight-
// RGBA destination through the blend operator. It is used by the image blits whose
// loops only ever visit pixels the image actually covers, so a full rasterizer
// cover (255) is correct and no untouched background is disturbed. In the real AGG
// build it routes through comp_op_adaptor_rgba_plain — the same premultiply →
// comp_op → demultiply path the gradient cover blit and AGG's comp-op pixfmt use —
// so image draw honours the full operator set, matching the port's comp-op image
// renderer (renBaseCompPre). The stub build has only the five composite_pixel
// operators and is never advertised as a valid backend.
void blend_image_pixel(uint8_t* d, const uint8_t* s, int blend_mode) {
#ifdef AGG_GO_CPP_REAL
  const unsigned op = static_cast<unsigned>(map_comp_op(blend_mode));
  comp_op_adaptor_rgba_plain<agg::rgba8, agg::order_rgba>::blend_pix(op, d, s[0], s[1], s[2], s[3],
                                                                     255);
#else
  composite_pixel(d, s, blend_mode);
#endif
}

bool point_in_path_even_odd(const AggGoCPPPath& path, float px, float py) {
  bool inside = false;
  const auto count = path.points.size();
  for (size_t i = 0, j = count - 1; i < count; j = i++) {
    const auto& a = path.points[j];
    const auto& b = path.points[i];
    const bool intersects = ((a.y > py) != (b.y > py)) &&
                            (px < (b.x - a.x) * (py - a.y) / ((b.y - a.y) + 1e-12f) + a.x);
    if (intersects) {
      inside = !inside;
    }
  }
  return inside;
}

int winding_direction(const Point& a, const Point& b, float px, float py) {
  const float cross = (b.x - a.x) * (py - a.y) - (px - a.x) * (b.y - a.y);
  if (cross > 0.0f) {
    return 1;
  }
  if (cross < 0.0f) {
    return -1;
  }
  return 0;
}

bool point_in_path_non_zero(const AggGoCPPPath& path, float px, float py) {
  int winding = 0;
  const auto count = path.points.size();
  for (size_t i = 0, j = count - 1; i < count; j = i++) {
    const auto& a = path.points[j];
    const auto& b = path.points[i];
    if (a.y <= py) {
      if (b.y > py && winding_direction(a, b, px, py) > 0) {
        ++winding;
      }
    } else if (b.y <= py && winding_direction(a, b, px, py) < 0) {
      --winding;
    }
  }
  return winding != 0;
}

bool point_in_path(const AggGoCPPPath& path, int fill_rule, float px, float py) {
  if (fill_rule == AggGoCPPFillRuleEvenOdd) {
    return point_in_path_even_odd(path, px, py);
  }
  return point_in_path_non_zero(path, px, py);
}

float distance_squared(float x1, float y1, float x2, float y2) {
  const float dx = x2 - x1;
  const float dy = y2 - y1;
  return dx * dx + dy * dy;
}

float distance_to_segment_squared(const Point& a, const Point& b, float px, float py, int line_cap,
                                  float radius) {
  const float vx = b.x - a.x;
  const float vy = b.y - a.y;
  const float seg_len_sq = vx * vx + vy * vy;
  if (seg_len_sq <= 1e-12f) {
    return distance_squared(a.x, a.y, px, py);
  }

  float t = ((px - a.x) * vx + (py - a.y) * vy) / seg_len_sq;
  if (line_cap == AggGoCPPLineCapSquare) {
    const float seg_len = std::sqrt(seg_len_sq);
    const float extend = radius / seg_len;
    t = std::max(-extend, std::min(1.0f + extend, t));
  } else {
    t = std::max(0.0f, std::min(1.0f, t));
  }

  const float proj_x = a.x + t * vx;
  const float proj_y = a.y + t * vy;
  float dist_sq = distance_squared(proj_x, proj_y, px, py);

  if (line_cap == AggGoCPPLineCapRound) {
    dist_sq = std::min(dist_sq, distance_squared(a.x, a.y, px, py));
    dist_sq = std::min(dist_sq, distance_squared(b.x, b.y, px, py));
  }
  return dist_sq;
}

double signed_area2(const PointD& a, const PointD& b, const PointD& c) {
  return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x);
}

bool barycentric_weights(const PointD& p, const PointD& a, const PointD& b, const PointD& c,
                         double* w0, double* w1, double* w2) {
  const double denom = signed_area2(a, b, c);
  if (std::abs(denom) <= 1e-12) {
    return false;
  }
  *w0 = signed_area2(p, b, c) / denom;
  *w1 = signed_area2(a, p, c) / denom;
  *w2 = signed_area2(a, b, p) / denom;
  constexpr double kEps = -1e-9;
  return *w0 >= kEps && *w1 >= kEps && *w2 >= kEps;
}

void matrix_set_identity(AggGoCPPMatrix* matrix) {
  matrix->a = 1.0;
  matrix->b = 0.0;
  matrix->c = 0.0;
  matrix->d = 1.0;
  matrix->e = 0.0;
  matrix->f = 0.0;
}

void matrix_multiply(AggGoCPPMatrix* lhs, const AggGoCPPMatrix& rhs) {
  AggGoCPPMatrix out;
  out.a = lhs->a * rhs.a + lhs->c * rhs.b;
  out.b = lhs->b * rhs.a + lhs->d * rhs.b;
  out.c = lhs->a * rhs.c + lhs->c * rhs.d;
  out.d = lhs->b * rhs.c + lhs->d * rhs.d;
  out.e = lhs->a * rhs.e + lhs->c * rhs.f + lhs->e;
  out.f = lhs->b * rhs.e + lhs->d * rhs.f + lhs->f;
  *lhs = out;
}

// matrix_premultiply applies `primitive` in the matrix's output space, i.e.
// matrix := primitive ∘ matrix. This matches agg::trans_affine::translate /
// rotate / scale, where each incremental op composes so that the most recent
// call is applied last (outermost) to a point. matrix_multiply alone composes
// the other way (primitive innermost), which would reverse a Translate→Rotate→
// Scale sequence relative to the faithful AGG port — see internal/transform.
void matrix_premultiply(AggGoCPPMatrix* matrix, const AggGoCPPMatrix& primitive) {
  AggGoCPPMatrix combined = primitive;
  matrix_multiply(&combined, *matrix);
  *matrix = combined;
}

Point matrix_transform_point(const AggGoCPPMatrix& matrix, const Point& point) {
  return Point{
      static_cast<float>(matrix.a * point.x + matrix.c * point.y + matrix.e),
      static_cast<float>(matrix.b * point.x + matrix.d * point.y + matrix.f),
  };
}

}  // namespace

extern "C" int agg_go_cpp_bridge_probe(void) {
#ifdef AGG_GO_CPP_REAL
  last_error_storage().clear();
  return 0;
#else
  last_error_storage() = "stub bridge: no AGG-backed implementation linked";
  return -1;
#endif
}

extern "C" int agg_go_cpp_bridge_is_stub(void) {
#ifdef AGG_GO_CPP_REAL
  return 0;
#else
  return 1;
#endif
}

extern "C" const char* agg_go_cpp_bridge_build_id(void) {
#ifdef AGG_GO_CPP_REAL
  return "agogo-agg-real-v1";
#else
  return "agogo-primitives-stub-v2";
#endif
}

extern "C" const char* agg_go_cpp_bridge_last_error(void) {
  return last_error_storage().c_str();
}

extern "C" AggGoCPPImage* agg_go_cpp_image_create(uint32_t width, uint32_t height) {
  if (width == 0 || height == 0) {
    set_last_error("image dimensions must be positive");
    return nullptr;
  }
  auto* image = new AggGoCPPImage();
  image->width = width;
  image->height = height;
  image->stride = width * 4;
  image->pixels.assign(static_cast<size_t>(image->stride) * height, 0);
#ifdef AGG_GO_CPP_REAL
  init_rendering(image);
#endif
  return image;
}

extern "C" void agg_go_cpp_image_free(AggGoCPPImage* image) { delete image; }

extern "C" uint32_t agg_go_cpp_image_width(const AggGoCPPImage* image) {
  return image ? image->width : 0;
}

extern "C" uint32_t agg_go_cpp_image_height(const AggGoCPPImage* image) {
  return image ? image->height : 0;
}

extern "C" uint32_t agg_go_cpp_image_stride(const AggGoCPPImage* image) {
  return image ? image->stride : 0;
}

extern "C" const uint8_t* agg_go_cpp_image_pixels(const AggGoCPPImage* image) {
  if (!valid_image(image) || image->pixels.empty()) {
    return nullptr;
  }
  return image->pixels.data();
}

extern "C" uint8_t* agg_go_cpp_image_pixels_mut(AggGoCPPImage* image) {
  if (!valid_image(image) || image->pixels.empty()) {
    return nullptr;
  }
  return image->pixels.data();
}

extern "C" int agg_go_cpp_image_clear(AggGoCPPImage* image, uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
#ifdef AGG_GO_CPP_REAL
  image->ren_base->clear(agg::rgba8(r, g, b, a));
  return 0;
#else
  for (size_t i = 0; i < image->pixels.size(); i += 4) {
    image->pixels[i + 0] = r;
    image->pixels[i + 1] = g;
    image->pixels[i + 2] = b;
    image->pixels[i + 3] = a;
  }
  return 0;
#endif
}

extern "C" int agg_go_cpp_image_blit(AggGoCPPImage* dst, const AggGoCPPImage* src, int dst_x, int dst_y,
                                     int src_x, int src_y, uint32_t width, uint32_t height) {
  if (!valid_image(dst)) {
    set_last_error("destination image is nil");
    return -1;
  }
  if (!valid_image(src)) {
    set_last_error("source image is nil");
    return -1;
  }
  if (width == 0 || height == 0) {
    set_last_error("blit dimensions must be positive");
    return -1;
  }
  if (src_x < 0 || src_y < 0 || dst_x < 0 || dst_y < 0) {
    set_last_error("blit coordinates must be non-negative");
    return -1;
  }
  if (static_cast<uint32_t>(src_x) + width > src->width || static_cast<uint32_t>(src_y) + height > src->height) {
    set_last_error("source blit region exceeds image bounds");
    return -1;
  }
  if (static_cast<uint32_t>(dst_x) + width > dst->width || static_cast<uint32_t>(dst_y) + height > dst->height) {
    set_last_error("destination blit region exceeds image bounds");
    return -1;
  }

  for (uint32_t row = 0; row < height; ++row) {
    const size_t src_offset = static_cast<size_t>(src_y + static_cast<int>(row)) * src->stride +
                              static_cast<size_t>(src_x) * 4;
    const size_t dst_offset = static_cast<size_t>(dst_y + static_cast<int>(row)) * dst->stride +
                              static_cast<size_t>(dst_x) * 4;
    std::memcpy(&dst->pixels[dst_offset], &src->pixels[src_offset], static_cast<size_t>(width) * 4);
  }
  return 0;
}

extern "C" int agg_go_cpp_image_composite(AggGoCPPImage* dst, const AggGoCPPImage* src, int dst_x, int dst_y,
                                          int clip_x1, int clip_y1, int clip_x2, int clip_y2,
                                          int blend_mode) {
  if (!valid_image(dst)) {
    set_last_error("destination image is nil");
    return -1;
  }
  if (!valid_image(src)) {
    set_last_error("source image is nil");
    return -1;
  }
  if (!supported_blend_mode(blend_mode)) {
    set_last_error("unsupported blend mode");
    return -1;
  }

  const RectI clip = clamp_clip_rect(dst, clip_x1, clip_y1, clip_x2, clip_y2);
  if (rect_empty(clip)) {
    return 0;
  }

  const int visible_x1 = std::max(dst_x, clip.x1);
  const int visible_y1 = std::max(dst_y, clip.y1);
  const int visible_x2 = std::min(dst_x + static_cast<int>(src->width), clip.x2);
  const int visible_y2 = std::min(dst_y + static_cast<int>(src->height), clip.y2);
  if (visible_x1 >= visible_x2 || visible_y1 >= visible_y2) {
    return 0;
  }

  for (int y = visible_y1; y < visible_y2; ++y) {
    for (int x = visible_x1; x < visible_x2; ++x) {
      const int src_px = x - dst_x;
      const int src_py = y - dst_y;
      const size_t src_offset =
          static_cast<size_t>(src_py) * src->stride + static_cast<size_t>(src_px) * 4;
      const size_t dst_offset =
          static_cast<size_t>(y) * dst->stride + static_cast<size_t>(x) * 4;
      composite_pixel(&dst->pixels[dst_offset], &src->pixels[src_offset], blend_mode);
    }
  }
  return 0;
}

// agg_go_cpp_image_composite_cover composites a straight-RGBA source layer onto
// the destination through the comp-op operator, using a separate per-pixel
// coverage mask as the rasterizer cover. This is the faithful path for gradient
// (and other per-pixel-coloured) fills under a non-src-over blend: the operator
// is applied only where the shape has geometric coverage (cover > 0), exactly as
// AGG's renderer_scanline_aa + span_gradient + comp-op pixfmt would, so pixels
// outside the shape (cover == 0) leave the destination untouched. The plain
// whole-rect composite path applies the operator everywhere, which makes src and
// clear wipe the untouched background. src and dst share the canvas dimensions
// and are aligned at (0,0); cover is a width*height 8-bit buffer.
extern "C" int agg_go_cpp_image_composite_cover(AggGoCPPImage* dst, const AggGoCPPImage* src,
                                                const uint8_t* cover, int cover_stride, int clip_x1,
                                                int clip_y1, int clip_x2, int clip_y2, int blend_mode) {
  if (!valid_image(dst)) {
    set_last_error("destination image is nil");
    return -1;
  }
  if (!valid_image(src)) {
    set_last_error("source image is nil");
    return -1;
  }
  if (cover == nullptr) {
    set_last_error("cover buffer is nil");
    return -1;
  }
  if (!supported_comp_op_mode(blend_mode)) {
    set_last_error("unsupported blend mode");
    return -1;
  }

  const RectI clip = clamp_clip_rect(dst, clip_x1, clip_y1, clip_x2, clip_y2);
  if (rect_empty(clip)) {
    return 0;
  }

#ifdef AGG_GO_CPP_REAL
  const unsigned op = static_cast<unsigned>(map_comp_op(blend_mode));
#endif
  for (int y = clip.y1; y < clip.y2; ++y) {
    for (int x = clip.x1; x < clip.x2; ++x) {
      const uint8_t cov = cover[static_cast<size_t>(y) * cover_stride + static_cast<size_t>(x)];
      if (cov == 0) {
        continue;  // no geometric coverage: leave the destination untouched
      }
      const size_t src_offset = static_cast<size_t>(y) * src->stride + static_cast<size_t>(x) * 4;
      const size_t dst_offset = static_cast<size_t>(y) * dst->stride + static_cast<size_t>(x) * 4;
      const uint8_t* s = &src->pixels[src_offset];
      uint8_t* d = &dst->pixels[dst_offset];
#ifdef AGG_GO_CPP_REAL
      comp_op_adaptor_rgba_plain<agg::rgba8, agg::order_rgba>::blend_pix(op, d, s[0], s[1], s[2], s[3],
                                                                        static_cast<agg::cover_type>(cov));
#else
      // Stub fallback (never advertised as a valid backend): apply the operator
      // only on covered pixels, scaling the source alpha by the coverage so the
      // untouched background is preserved.
      uint8_t scaled[4] = {s[0], s[1], s[2],
                           static_cast<uint8_t>((static_cast<int>(s[3]) * cov) / 255)};
      composite_pixel(d, scaled, blend_mode);
#endif
    }
  }
  return 0;
}

extern "C" int agg_go_cpp_image_composite_scaled(AggGoCPPImage* dst, const AggGoCPPImage* src, int src_x,
                                                 int src_y, uint32_t src_width, uint32_t src_height,
                                                 int dst_x, int dst_y, uint32_t dst_width,
                                                 uint32_t dst_height, int clip_x1, int clip_y1,
                                                 int clip_x2, int clip_y2, int blend_mode) {
  if (!valid_image(dst)) {
    set_last_error("destination image is nil");
    return -1;
  }
  if (!valid_image(src)) {
    set_last_error("source image is nil");
    return -1;
  }
  if (!supported_comp_op_mode(blend_mode)) {
    set_last_error("unsupported blend mode");
    return -1;
  }
  if (src_width == 0 || src_height == 0) {
    set_last_error("source dimensions must be positive");
    return -1;
  }
  if (dst_width == 0 || dst_height == 0) {
    set_last_error("destination dimensions must be positive");
    return -1;
  }
  if (src_x < 0 || src_y < 0) {
    set_last_error("source coordinates must be non-negative");
    return -1;
  }
  if (static_cast<uint32_t>(src_x) + src_width > src->width ||
      static_cast<uint32_t>(src_y) + src_height > src->height) {
    set_last_error("source region exceeds image bounds");
    return -1;
  }

  const RectI clip = clamp_clip_rect(dst, clip_x1, clip_y1, clip_x2, clip_y2);
  if (rect_empty(clip)) {
    return 0;
  }

  const int visible_x1 = std::max(dst_x, clip.x1);
  const int visible_y1 = std::max(dst_y, clip.y1);
  const int visible_x2 = std::min(dst_x + static_cast<int>(dst_width), clip.x2);
  const int visible_y2 = std::min(dst_y + static_cast<int>(dst_height), clip.y2);
  if (visible_x1 >= visible_x2 || visible_y1 >= visible_y2) {
    return 0;
  }

  for (int y = visible_y1; y < visible_y2; ++y) {
    const uint32_t src_py = static_cast<uint32_t>(
        (static_cast<uint64_t>(y - dst_y) * src_height) / dst_height);
    for (int x = visible_x1; x < visible_x2; ++x) {
      const uint32_t src_px = static_cast<uint32_t>(
          (static_cast<uint64_t>(x - dst_x) * src_width) / dst_width);
      const size_t src_offset = static_cast<size_t>(src_y + static_cast<int>(src_py)) * src->stride +
                                static_cast<size_t>(src_x + static_cast<int>(src_px)) * 4;
      const size_t dst_offset =
          static_cast<size_t>(y) * dst->stride + static_cast<size_t>(x) * 4;
      blend_image_pixel(&dst->pixels[dst_offset], &src->pixels[src_offset], blend_mode);
    }
  }
  return 0;
}

extern "C" int agg_go_cpp_image_composite_quad(AggGoCPPImage* dst, const AggGoCPPImage* src, int src_x,
                                               int src_y, uint32_t src_width, uint32_t src_height,
                                               const double* quad_xy, int clip_x1, int clip_y1,
                                               int clip_x2, int clip_y2, int blend_mode) {
  if (!valid_image(dst)) {
    set_last_error("destination image is nil");
    return -1;
  }
  if (!valid_image(src)) {
    set_last_error("source image is nil");
    return -1;
  }
  if (!supported_comp_op_mode(blend_mode)) {
    set_last_error("unsupported blend mode");
    return -1;
  }
  if (src_width == 0 || src_height == 0) {
    set_last_error("source dimensions must be positive");
    return -1;
  }
  if (quad_xy == nullptr) {
    set_last_error("quad is nil");
    return -1;
  }
  if (src_x < 0 || src_y < 0) {
    set_last_error("source coordinates must be non-negative");
    return -1;
  }
  if (static_cast<uint32_t>(src_x) + src_width > src->width ||
      static_cast<uint32_t>(src_y) + src_height > src->height) {
    set_last_error("source region exceeds image bounds");
    return -1;
  }

  const PointD q0{quad_xy[0], quad_xy[1]};
  const PointD q1{quad_xy[2], quad_xy[3]};
  const PointD q2{quad_xy[4], quad_xy[5]};
  const PointD q3{quad_xy[6], quad_xy[7]};
  if (std::abs(signed_area2(q0, q1, q2)) <= 1e-12 || std::abs(signed_area2(q0, q2, q3)) <= 1e-12) {
    set_last_error("quad is degenerate");
    return -1;
  }

  const RectI clip = clamp_clip_rect(dst, clip_x1, clip_y1, clip_x2, clip_y2);
  if (rect_empty(clip)) {
    return 0;
  }

  const double min_x = std::min(std::min(q0.x, q1.x), std::min(q2.x, q3.x));
  const double max_x = std::max(std::max(q0.x, q1.x), std::max(q2.x, q3.x));
  const double min_y = std::min(std::min(q0.y, q1.y), std::min(q2.y, q3.y));
  const double max_y = std::max(std::max(q0.y, q1.y), std::max(q2.y, q3.y));

  const int start_x = std::max(clip.x1, static_cast<int>(std::floor(min_x)));
  const int end_x = std::min(clip.x2, static_cast<int>(std::ceil(max_x)));
  const int start_y = std::max(clip.y1, static_cast<int>(std::floor(min_y)));
  const int end_y = std::min(clip.y2, static_cast<int>(std::ceil(max_y)));

  const PointD s0{static_cast<double>(src_x), static_cast<double>(src_y)};
  const PointD s1{static_cast<double>(src_x) + static_cast<double>(src_width),
                  static_cast<double>(src_y)};
  const PointD s2{static_cast<double>(src_x) + static_cast<double>(src_width),
                  static_cast<double>(src_y) + static_cast<double>(src_height)};
  const PointD s3{static_cast<double>(src_x),
                  static_cast<double>(src_y) + static_cast<double>(src_height)};

  for (int y = start_y; y < end_y; ++y) {
    for (int x = start_x; x < end_x; ++x) {
      const PointD p{static_cast<double>(x) + 0.5, static_cast<double>(y) + 0.5};
      double w0 = 0.0;
      double w1 = 0.0;
      double w2 = 0.0;
      double sample_x = 0.0;
      double sample_y = 0.0;

      if (barycentric_weights(p, q0, q1, q2, &w0, &w1, &w2)) {
        sample_x = w0 * s0.x + w1 * s1.x + w2 * s2.x;
        sample_y = w0 * s0.y + w1 * s1.y + w2 * s2.y;
      } else if (barycentric_weights(p, q0, q2, q3, &w0, &w1, &w2)) {
        sample_x = w0 * s0.x + w1 * s2.x + w2 * s3.x;
        sample_y = w0 * s0.y + w1 * s2.y + w2 * s3.y;
      } else {
        continue;
      }

      int src_px = static_cast<int>(std::floor(sample_x));
      int src_py = static_cast<int>(std::floor(sample_y));
      src_px = std::max(src_x, std::min(src_x + static_cast<int>(src_width) - 1, src_px));
      src_py = std::max(src_y, std::min(src_y + static_cast<int>(src_height) - 1, src_py));

      const size_t src_offset =
          static_cast<size_t>(src_py) * src->stride + static_cast<size_t>(src_px) * 4;
      const size_t dst_offset =
          static_cast<size_t>(y) * dst->stride + static_cast<size_t>(x) * 4;
      blend_image_pixel(&dst->pixels[dst_offset], &src->pixels[src_offset], blend_mode);
    }
  }
  return 0;
}

extern "C" AggGoCPPPath* agg_go_cpp_path_create(void) { return new AggGoCPPPath(); }

extern "C" void agg_go_cpp_path_free(AggGoCPPPath* path) { delete path; }

extern "C" int agg_go_cpp_path_reset(AggGoCPPPath* path) {
  if (path == nullptr) {
    set_last_error("path is nil");
    return -1;
  }
  path->points.clear();
  path->closed = false;
  return 0;
}

extern "C" int agg_go_cpp_path_move_to(AggGoCPPPath* path, float x, float y) {
  if (path == nullptr) {
    set_last_error("path is nil");
    return -1;
  }
  path->points.clear();
  path->points.push_back({x, y});
  path->closed = false;
  return 0;
}

extern "C" int agg_go_cpp_path_line_to(AggGoCPPPath* path, float x, float y) {
  if (path == nullptr) {
    set_last_error("path is nil");
    return -1;
  }
  path->points.push_back({x, y});
  return 0;
}

extern "C" int agg_go_cpp_path_close(AggGoCPPPath* path) {
  if (path == nullptr) {
    set_last_error("path is nil");
    return -1;
  }
  path->closed = true;
  return 0;
}

extern "C" AggGoCPPMatrix* agg_go_cpp_matrix_create(void) {
  auto* matrix = new AggGoCPPMatrix();
  matrix_set_identity(matrix);
  return matrix;
}

extern "C" void agg_go_cpp_matrix_free(AggGoCPPMatrix* matrix) { delete matrix; }

extern "C" int agg_go_cpp_matrix_identity(AggGoCPPMatrix* matrix) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  matrix_set_identity(matrix);
  return 0;
}

extern "C" int agg_go_cpp_matrix_translate(AggGoCPPMatrix* matrix, float tx, float ty) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  AggGoCPPMatrix translation{1.0, 0.0, 0.0, 1.0, tx, ty};
  matrix_premultiply(matrix, translation);
  return 0;
}

extern "C" int agg_go_cpp_matrix_scale(AggGoCPPMatrix* matrix, float sx, float sy) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  AggGoCPPMatrix scale{sx, 0.0, 0.0, sy, 0.0, 0.0};
  matrix_premultiply(matrix, scale);
  return 0;
}

extern "C" int agg_go_cpp_matrix_rotate(AggGoCPPMatrix* matrix, float angle) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  const double sin_v = std::sin(static_cast<double>(angle));
  const double cos_v = std::cos(static_cast<double>(angle));
  AggGoCPPMatrix rotation{cos_v, sin_v, -sin_v, cos_v, 0.0, 0.0};
  matrix_premultiply(matrix, rotation);
  return 0;
}

extern "C" int agg_go_cpp_matrix_transform_point(const AggGoCPPMatrix* matrix, double* x, double* y) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  if (x == nullptr || y == nullptr) {
    set_last_error("point coordinates are nil");
    return -1;
  }
  const Point out = matrix_transform_point(*matrix, Point{static_cast<float>(*x), static_cast<float>(*y)});
  *x = out.x;
  *y = out.y;
  return 0;
}

extern "C" int agg_go_cpp_matrix_store(const AggGoCPPMatrix* matrix, double* out) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  if (out == nullptr) {
    set_last_error("output buffer is nil");
    return -1;
  }
  out[0] = matrix->a;
  out[1] = matrix->b;
  out[2] = matrix->c;
  out[3] = matrix->d;
  out[4] = matrix->e;
  out[5] = matrix->f;
  return 0;
}

extern "C" AggGoCPPPath* agg_go_cpp_path_transform(const AggGoCPPPath* path, const AggGoCPPMatrix* matrix) {
  if (path == nullptr) {
    set_last_error("path is nil");
    return nullptr;
  }
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return nullptr;
  }
  auto* out = new AggGoCPPPath();
  out->closed = path->closed;
  out->points.reserve(path->points.size());
  for (const auto& point : path->points) {
    out->points.push_back(matrix_transform_point(*matrix, point));
  }
  return out;
}

extern "C" int agg_go_cpp_render_fill_path(AggGoCPPImage* image, const AggGoCPPPath* path, int fill_rule,
                                           uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
  if (!valid_path(path)) {
    set_last_error("path must contain at least three points");
    return -1;
  }

#ifdef AGG_GO_CPP_REAL
  agg::path_storage agg_path;
  convert_path_to_agg(*path, &agg_path);
  agg::rasterizer_scanline_aa<> ras;
  if (fill_rule == AggGoCPPFillRuleEvenOdd) {
    ras.filling_rule(agg::fill_even_odd);
  } else {
    ras.filling_rule(agg::fill_non_zero);
  }
  agg::scanline_u8 sl;
  ras.add_path(agg_path);
  agg::renderer_scanline_aa_solid<agg::renderer_base<agg::pixfmt_rgba32>> ren(*image->ren_base);
  ren.color(agg::rgba8(r, g, b, a));
  agg::render_scanlines(ras, sl, ren);
  return 0;
#else
  float min_x = path->points[0].x;
  float max_x = path->points[0].x;
  float min_y = path->points[0].y;
  float max_y = path->points[0].y;
  for (const auto& point : path->points) {
    min_x = std::min(min_x, point.x);
    max_x = std::max(max_x, point.x);
    min_y = std::min(min_y, point.y);
    max_y = std::max(max_y, point.y);
  }

  const int start_x = std::max(0, static_cast<int>(std::floor(min_x)));
  const int end_x = std::min(static_cast<int>(image->width), static_cast<int>(std::ceil(max_x)));
  const int start_y = std::max(0, static_cast<int>(std::floor(min_y)));
  const int end_y = std::min(static_cast<int>(image->height), static_cast<int>(std::ceil(max_y)));

  for (int y = start_y; y < end_y; ++y) {
    for (int x = start_x; x < end_x; ++x) {
      const float px = static_cast<float>(x) + 0.5f;
      const float py = static_cast<float>(y) + 0.5f;
      if (!point_in_path(*path, fill_rule, px, py)) {
        continue;
      }
      const size_t offset = static_cast<size_t>(y) * image->stride + static_cast<size_t>(x) * 4;
      image->pixels[offset + 0] = r;
      image->pixels[offset + 1] = g;
      image->pixels[offset + 2] = b;
      image->pixels[offset + 3] = a;
    }
  }

  return 0;
#endif
}

extern "C" int agg_go_cpp_render_stroke_path(AggGoCPPImage* image, const AggGoCPPPath* path, float width,
                                             int line_cap, int line_join, float miter_limit, uint8_t r,
                                             uint8_t g, uint8_t b, uint8_t a,
                                             const AggGoCPPMatrix* matrix) {
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
  if (!valid_stroke_path(path)) {
    set_last_error("path must contain at least two points for stroking");
    return -1;
  }
  if (!(width > 0.0f)) {
    set_last_error("stroke width must be positive");
    return -1;
  }
  if (line_cap < AggGoCPPLineCapButt || line_cap > AggGoCPPLineCapSquare) {
    set_last_error("invalid line cap");
    return -1;
  }
  if (line_join < AggGoCPPLineJoinMiter || line_join > AggGoCPPLineJoinBevel) {
    set_last_error("invalid line join");
    return -1;
  }
  if (line_join == AggGoCPPLineJoinMiter && miter_limit < 1.0f) {
    set_last_error("miter limit must be >= 1 for miter joins");
    return -1;
  }

#ifdef AGG_GO_CPP_REAL
  agg::path_storage agg_path;
  convert_path_to_agg(*path, &agg_path);
  agg::conv_stroke<agg::path_storage> stroke(agg_path);
  stroke.width(width);
  switch (line_cap) {
    case AggGoCPPLineCapRound:
      stroke.line_cap(agg::round_cap);
      break;
    case AggGoCPPLineCapSquare:
      stroke.line_cap(agg::square_cap);
      break;
    default:
      stroke.line_cap(agg::butt_cap);
      break;
  }
  switch (line_join) {
    case AggGoCPPLineJoinRound:
      stroke.line_join(agg::round_join);
      break;
    case AggGoCPPLineJoinBevel:
      stroke.line_join(agg::bevel_join);
      break;
    default:
      stroke.line_join(agg::miter_join);
      stroke.miter_limit(miter_limit);
      break;
  }
  agg::rasterizer_scanline_aa<> ras;
  agg::scanline_u8 sl;
  add_stroke_to_ras(ras, stroke, matrix);
  agg::renderer_scanline_aa_solid<agg::renderer_base<agg::pixfmt_rgba32>> ren(*image->ren_base);
  ren.color(agg::rgba8(r, g, b, a));
  agg::render_scanlines(ras, sl, ren);
  return 0;
#else
  float min_x = path->points[0].x;
  float max_x = path->points[0].x;
  float min_y = path->points[0].y;
  float max_y = path->points[0].y;
  for (const auto& point : path->points) {
    min_x = std::min(min_x, point.x);
    max_x = std::max(max_x, point.x);
    min_y = std::min(min_y, point.y);
    max_y = std::max(max_y, point.y);
  }
  const float radius = width * 0.5f;
  min_x -= radius;
  max_x += radius;
  min_y -= radius;
  max_y += radius;

  const int start_x = std::max(0, static_cast<int>(std::floor(min_x)));
  const int end_x = std::min(static_cast<int>(image->width), static_cast<int>(std::ceil(max_x)));
  const int start_y = std::max(0, static_cast<int>(std::floor(min_y)));
  const int end_y = std::min(static_cast<int>(image->height), static_cast<int>(std::ceil(max_y)));
  const float radius_sq = radius * radius;

  for (int y = start_y; y < end_y; ++y) {
    for (int x = start_x; x < end_x; ++x) {
      const float px = static_cast<float>(x) + 0.5f;
      const float py = static_cast<float>(y) + 0.5f;
      bool hit = false;
      for (size_t i = 1; i < path->points.size(); ++i) {
        const auto& a_point = path->points[i - 1];
        const auto& b_point = path->points[i];
        if (distance_to_segment_squared(a_point, b_point, px, py, line_cap, radius) <= radius_sq) {
          hit = true;
          break;
        }
      }
      if (!hit && path->closed) {
        const auto& a_point = path->points.back();
        const auto& b_point = path->points.front();
        hit = distance_to_segment_squared(a_point, b_point, px, py, line_cap, radius) <= radius_sq;
      }
      if (!hit) {
        continue;
      }
      const size_t offset = static_cast<size_t>(y) * image->stride + static_cast<size_t>(x) * 4;
      image->pixels[offset + 0] = r;
      image->pixels[offset + 1] = g;
      image->pixels[offset + 2] = b;
      image->pixels[offset + 3] = a;
    }
  }

  return 0;
#endif
}

extern "C" int agg_go_cpp_render_stroke_path_dashed(AggGoCPPImage* image, const AggGoCPPPath* path,
                                                    float width, int line_cap, int line_join,
                                                    float miter_limit, const float* dashes,
                                                    int dash_pair_count, float dash_start, uint8_t r,
                                                    uint8_t g, uint8_t b, uint8_t a,
                                                    const AggGoCPPMatrix* matrix) {
  if (dash_pair_count < 0 || (dash_pair_count > 0 && dashes == nullptr)) {
    set_last_error("invalid dash pattern");
    return -1;
  }
  // No dashes: identical to a solid stroke.
  if (dash_pair_count == 0) {
    return agg_go_cpp_render_stroke_path(image, path, width, line_cap, line_join, miter_limit, r, g,
                                         b, a, matrix);
  }
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
  if (!valid_stroke_path(path)) {
    set_last_error("path must contain at least two points for stroking");
    return -1;
  }
  if (!(width > 0.0f)) {
    set_last_error("stroke width must be positive");
    return -1;
  }
  if (line_cap < AggGoCPPLineCapButt || line_cap > AggGoCPPLineCapSquare) {
    set_last_error("invalid line cap");
    return -1;
  }
  if (line_join < AggGoCPPLineJoinMiter || line_join > AggGoCPPLineJoinBevel) {
    set_last_error("invalid line join");
    return -1;
  }
  if (line_join == AggGoCPPLineJoinMiter && miter_limit < 1.0f) {
    set_last_error("miter limit must be >= 1 for miter joins");
    return -1;
  }
  for (int i = 0; i < dash_pair_count * 2; ++i) {
    if (!(dashes[i] >= 0.0f)) {
      set_last_error("dash lengths must be non-negative");
      return -1;
    }
  }

#ifdef AGG_GO_CPP_REAL
  agg::path_storage agg_path;
  convert_path_to_agg(*path, &agg_path);
  agg::conv_dash<agg::path_storage> dash(agg_path);
  for (int i = 0; i < dash_pair_count; ++i) {
    dash.add_dash(dashes[i * 2], dashes[i * 2 + 1]);
  }
  dash.dash_start(dash_start);
  agg::conv_stroke<agg::conv_dash<agg::path_storage>> stroke(dash);
  stroke.width(width);
  switch (line_cap) {
    case AggGoCPPLineCapRound:
      stroke.line_cap(agg::round_cap);
      break;
    case AggGoCPPLineCapSquare:
      stroke.line_cap(agg::square_cap);
      break;
    default:
      stroke.line_cap(agg::butt_cap);
      break;
  }
  switch (line_join) {
    case AggGoCPPLineJoinRound:
      stroke.line_join(agg::round_join);
      break;
    case AggGoCPPLineJoinBevel:
      stroke.line_join(agg::bevel_join);
      break;
    default:
      stroke.line_join(agg::miter_join);
      stroke.miter_limit(miter_limit);
      break;
  }
  agg::rasterizer_scanline_aa<> ras;
  agg::scanline_u8 sl;
  add_stroke_to_ras(ras, stroke, matrix);
  agg::renderer_scanline_aa_solid<agg::renderer_base<agg::pixfmt_rgba32>> ren(*image->ren_base);
  ren.color(agg::rgba8(r, g, b, a));
  agg::render_scanlines(ras, sl, ren);
  return 0;
#else
  // The stub backend is never advertised as an available engine, so it does not
  // implement dash segmentation or matrix transforms; stroke solid so tagged
  // primitive tests still exercise the path geometry.
  return agg_go_cpp_render_stroke_path(image, path, width, line_cap, line_join, miter_limit, r, g, b,
                                       a, matrix);
#endif
}

extern "C" int agg_go_cpp_render_fill_path_comp(AggGoCPPImage* image, const AggGoCPPPath* path,
                                                int fill_rule, int blend_mode, int clip_x1,
                                                int clip_y1, int clip_x2, int clip_y2, uint8_t r,
                                                uint8_t g, uint8_t b, uint8_t a) {
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
  if (!valid_path(path)) {
    set_last_error("path must contain at least three points");
    return -1;
  }
  if (!supported_comp_op_mode(blend_mode)) {
    set_last_error("unsupported blend mode");
    return -1;
  }
  // The clip rectangle arrives half-open ([x1,x2) x [y1,y2)); an empty box means
  // nothing is drawn.
  if (clip_x2 <= clip_x1 || clip_y2 <= clip_y1) {
    return 0;
  }

#ifdef AGG_GO_CPP_REAL
  agg::path_storage agg_path;
  convert_path_to_agg(*path, &agg_path);
  agg::rasterizer_scanline_aa<> ras;
  ras.filling_rule(fill_rule == AggGoCPPFillRuleEvenOdd ? agg::fill_even_odd : agg::fill_non_zero);
  ras.add_path(agg_path);
  // renderer_base::clip_box takes an inclusive max corner.
  render_with_comp_op(image, ras, blend_mode, clip_x1, clip_y1, clip_x2 - 1, clip_y2 - 1, r, g, b,
                      a);
  return 0;
#else
  // The stub backend is never advertised; fall back to a plain unclipped fill so
  // tagged primitive tests still exercise the geometry.
  return agg_go_cpp_render_fill_path(image, path, fill_rule, r, g, b, a);
#endif
}

extern "C" int agg_go_cpp_render_stroke_path_comp(AggGoCPPImage* image, const AggGoCPPPath* path,
                                                  float width, int line_cap, int line_join,
                                                  float miter_limit, const float* dashes,
                                                  int dash_pair_count, float dash_start,
                                                  int blend_mode, int clip_x1, int clip_y1,
                                                  int clip_x2, int clip_y2, uint8_t r, uint8_t g,
                                                  uint8_t b, uint8_t a,
                                                  const AggGoCPPMatrix* matrix) {
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
  if (!valid_stroke_path(path)) {
    set_last_error("path must contain at least two points for stroking");
    return -1;
  }
  if (!(width > 0.0f)) {
    set_last_error("stroke width must be positive");
    return -1;
  }
  if (line_cap < AggGoCPPLineCapButt || line_cap > AggGoCPPLineCapSquare) {
    set_last_error("invalid line cap");
    return -1;
  }
  if (line_join < AggGoCPPLineJoinMiter || line_join > AggGoCPPLineJoinBevel) {
    set_last_error("invalid line join");
    return -1;
  }
  if (line_join == AggGoCPPLineJoinMiter && miter_limit < 1.0f) {
    set_last_error("miter limit must be >= 1 for miter joins");
    return -1;
  }
  if (dash_pair_count < 0 || (dash_pair_count > 0 && dashes == nullptr)) {
    set_last_error("invalid dash pattern");
    return -1;
  }
  if (!supported_comp_op_mode(blend_mode)) {
    set_last_error("unsupported blend mode");
    return -1;
  }
  if (clip_x2 <= clip_x1 || clip_y2 <= clip_y1) {
    return 0;
  }

#ifdef AGG_GO_CPP_REAL
  agg::path_storage agg_path;
  convert_path_to_agg(*path, &agg_path);
  agg::rasterizer_scanline_aa<> ras;
  if (dash_pair_count > 0) {
    agg::conv_dash<agg::path_storage> dash(agg_path);
    for (int i = 0; i < dash_pair_count; ++i) {
      dash.add_dash(dashes[i * 2], dashes[i * 2 + 1]);
    }
    dash.dash_start(dash_start);
    agg::conv_stroke<agg::conv_dash<agg::path_storage> > stroke(dash);
    configure_stroke(stroke, width, line_cap, line_join, miter_limit);
    add_stroke_to_ras(ras, stroke, matrix);
  } else {
    agg::conv_stroke<agg::path_storage> stroke(agg_path);
    configure_stroke(stroke, width, line_cap, line_join, miter_limit);
    add_stroke_to_ras(ras, stroke, matrix);
  }
  render_with_comp_op(image, ras, blend_mode, clip_x1, clip_y1, clip_x2 - 1, clip_y2 - 1, r, g, b,
                      a);
  return 0;
#else
  return agg_go_cpp_render_stroke_path_dashed(image, path, width, line_cap, line_join, miter_limit,
                                              dashes, dash_pair_count, dash_start, r, g, b, a,
                                              matrix);
#endif
}

extern "C" AggGoCPPFont* agg_go_cpp_font_create(const char* font_path) {
  if (font_path == nullptr || *font_path == '\0') {
    set_last_error("font path is empty");
    return nullptr;
  }
  auto* font = new AggGoCPPFont();
  font->font_path = font_path;
#ifdef AGG_GO_CPP_REAL
  font->font_engine = std::make_unique<AggGoCPPFont::font_engine_type>();
  font->font_manager = std::make_unique<AggGoCPPFont::font_manager_type>(*font->font_engine);
  setup_font_engine(font);
  if (!font->font_engine->load_font(font_path, 0, agg::glyph_ren_outline)) {
    set_last_error("failed to load font");
    delete font;
    return nullptr;
  }
  setup_font_engine(font);
  return font;
#else
  set_last_error("font support is unavailable in stub backend");
  delete font;
  return nullptr;
#endif
}

extern "C" void agg_go_cpp_font_free(AggGoCPPFont* font) { delete font; }

extern "C" int agg_go_cpp_font_set_size(AggGoCPPFont* font, float size) {
  if (!valid_font(font) || !(size > 0.0f)) {
    set_last_error("invalid font size");
    return -1;
  }
  font->font_size = size;
#ifdef AGG_GO_CPP_REAL
  setup_font_engine(font);
#endif
  return 0;
}

extern "C" int agg_go_cpp_font_set_hinting(AggGoCPPFont* font, int enabled) {
  if (!valid_font(font)) {
    set_last_error("font is nil");
    return -1;
  }
  font->hinting_enabled = enabled != 0;
#ifdef AGG_GO_CPP_REAL
  setup_font_engine(font);
#endif
  return 0;
}

extern "C" int agg_go_cpp_font_set_flip_y(AggGoCPPFont* font, int flip_y) {
  if (!valid_font(font)) {
    set_last_error("font is nil");
    return -1;
  }
  font->flip_y = flip_y != 0;
#ifdef AGG_GO_CPP_REAL
  setup_font_engine(font);
#endif
  return 0;
}

extern "C" int agg_go_cpp_render_text(AggGoCPPImage* image, AggGoCPPFont* font, const char* text, float x, float y,
                                      uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
  if (!valid_image(image) || !valid_font(font) || text == nullptr) {
    set_last_error("text rendering inputs are nil");
    return -1;
  }
#ifdef AGG_GO_CPP_REAL
  if (!image->ren_base || !font->font_engine || !font->font_manager) {
    set_last_error("font rendering backend is unavailable");
    return -1;
  }
  try {
    agg::renderer_scanline_aa_solid<agg::renderer_base<agg::pixfmt_rgba32>> ren(*image->ren_base);
    ren.color(agg::rgba8(r, g, b, a));
    agg::rasterizer_scanline_aa<> ras;
    agg::scanline_u8 sl;
    double pen_x = x;
    double pen_y = y;
    const char* p = text;
    while (*p) {
      uint32_t codepoint;
      if (!next_utf8(p, codepoint)) {
        break;
      }
      const agg::glyph_cache* glyph = font->font_manager->glyph(codepoint);
      if (!glyph) {
        continue;
      }
      font->font_manager->add_kerning(&pen_x, &pen_y);
      font->font_manager->init_embedded_adaptors(glyph, pen_x, pen_y);
      AggGoCPPFont::path_adaptor_type path_adaptor = font->font_manager->path_adaptor();
      AggGoCPPFont::conv_curves_type curves(path_adaptor);
      curves.approximation_scale(1.0);
      ras.reset();
      ras.add_path(curves);
      agg::render_scanlines(ras, sl, ren);
      pen_x += glyph->advance_x;
      pen_y += glyph->advance_y;
    }
    return 0;
  } catch (...) {
    set_last_error("font rendering failed");
    return -1;
  }
#else
  set_last_error("font support is unavailable in stub backend");
  return -1;
#endif
}

extern "C" float agg_go_cpp_text_width(AggGoCPPFont* font, const char* text) {
  if (!valid_font(font) || text == nullptr) {
    return 0.0f;
  }
#ifdef AGG_GO_CPP_REAL
  if (!font->font_engine || !font->font_manager) {
    return 0.0f;
  }
  float width = 0.0f;
  const char* p = text;
  while (*p) {
    uint32_t codepoint;
    if (!next_utf8(p, codepoint)) {
      break;
    }
    const agg::glyph_cache* glyph = font->font_manager->glyph(codepoint);
    if (glyph) {
      width += glyph->advance_x;
    }
  }
  return width;
#else
  return 0.0f;
#endif
}

extern "C" float agg_go_cpp_text_height(AggGoCPPFont* font) {
  if (!valid_font(font)) {
    return 0.0f;
  }
#ifdef AGG_GO_CPP_REAL
  if (!font->font_engine) {
    return font->font_size;
  }
  return font->font_engine->height();
#else
  return 0.0f;
#endif
}
