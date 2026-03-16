// Package span implements AGG's span-generation layer.
//
// Span generators sit between coordinate interpolation and pixel-format
// blending. Given a start pixel and run length, they expand one scanline span
// into concrete colors that renderer_scanline_aa then passes to pixfmt methods.
//
// The package follows AGG's original family split:
//
//   - interpolators map destination pixel positions back into source space
//   - gradient generators convert transformed coordinates into color lookup indices
//   - image filters and resamplers sample source images with nearest, bilinear,
//     kernel, or resampling logic
//   - pattern generators repeat source buffers with configurable offsets
//   - Gouraud generators interpolate vertex colors across triangles
//
// The upstream C++ references are primarily agg_span_interpolator*.h,
// agg_span_gradient*.h, agg_span_image_filter*.h, agg_span_pattern_*.h, and
// agg_span_gouraud*.h.
package span
