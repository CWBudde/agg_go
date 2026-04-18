//go:build agogo && cgo

#include "cpp_native.h"

#include <algorithm>
#include <cmath>
#include <cstdint>
#include <cstring>
#include <string>
#include <vector>

struct Point {
  float x;
  float y;
};

struct AggGoCPPImage {
  uint32_t width;
  uint32_t height;
  uint32_t stride;
  std::vector<uint8_t> pixels;
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

namespace {

constexpr int kBlendAlpha = 0;
constexpr int kBlendClear = 1;
constexpr int kBlendSrc = 2;
constexpr int kBlendDst = 3;
constexpr int kBlendSrcOver = 4;

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

void set_last_error(const char* value) { last_error_storage() = value; }

uint8_t clamp_byte(double value) {
  if (value <= 0.0) {
    return 0;
  }
  if (value >= 255.0) {
    return 255;
  }
  return static_cast<uint8_t>(std::lround(value));
}

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

Point matrix_transform_point(const AggGoCPPMatrix& matrix, const Point& point) {
  return Point{
      static_cast<float>(matrix.a * point.x + matrix.c * point.y + matrix.e),
      static_cast<float>(matrix.b * point.x + matrix.d * point.y + matrix.f),
  };
}

}  // namespace

extern "C" int agg_go_cpp_bridge_probe(void) {
  last_error_storage() = "stub bridge: no AGG-backed implementation linked";
  return -1;
}

extern "C" int agg_go_cpp_bridge_is_stub(void) { return 1; }

extern "C" const char* agg_go_cpp_bridge_build_id(void) {
  return "agogo-primitives-stub-v2";
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

extern "C" int agg_go_cpp_image_clear(AggGoCPPImage* image, uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
  if (!valid_image(image)) {
    set_last_error("image is nil");
    return -1;
  }
  for (size_t i = 0; i < image->pixels.size(); i += 4) {
    image->pixels[i + 0] = r;
    image->pixels[i + 1] = g;
    image->pixels[i + 2] = b;
    image->pixels[i + 3] = a;
  }
  return 0;
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
  if (!supported_blend_mode(blend_mode)) {
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
      composite_pixel(&dst->pixels[dst_offset], &src->pixels[src_offset], blend_mode);
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
  matrix_multiply(matrix, translation);
  return 0;
}

extern "C" int agg_go_cpp_matrix_scale(AggGoCPPMatrix* matrix, float sx, float sy) {
  if (!valid_matrix(matrix)) {
    set_last_error("matrix is nil");
    return -1;
  }
  AggGoCPPMatrix scale{sx, 0.0, 0.0, sy, 0.0, 0.0};
  matrix_multiply(matrix, scale);
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
  matrix_multiply(matrix, rotation);
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
}

extern "C" int agg_go_cpp_render_stroke_path(AggGoCPPImage* image, const AggGoCPPPath* path, float width,
                                             int line_cap, int line_join, float miter_limit, uint8_t r,
                                             uint8_t g, uint8_t b, uint8_t a) {
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
}
