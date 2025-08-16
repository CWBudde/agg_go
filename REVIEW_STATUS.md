# AGG 2.6 Go Port - File Checklist

This is a comprehensive checklist of files that need to be ported from the original AGG 2.6 C++ codebase to Go. Please always check the original C++ implementation for reference in ../agg-2.6

## Core Header Files (include/)

### Basic Types and Configuration

#### agg_basics.h - Core types, enums, path commands, geometry utilities

Go files:
- internal/basics/types.go
- internal/basics/constants.go
- internal/basics/path.go
- internal/basics/math.go

**Allocators (Templates → Go Generics)**

- [ ] pod_allocator<T> → Generic allocator interface
- [ ] obj_allocator<T> → Object allocator with constructors

**Basic Types**

- [ ] int8, int8u, int16, int16u, int32, int32u, int64, int64u type definitions
- [ ] cover_type (unsigned char)
- [ ] Enums: cover_scale_e, poly_subpixel_scale_e, filling_rule_e
- [ ] Path command enums: path_commands_e, path_flags_e

**Rounding Functions**

- [ ] iround(), uround(), ifloor(), ufloor(), iceil(), uceil()
- [ ] Platform-specific optimizations (FISTP, QIFIST)

**Template Structs → Go Generics**

- [x] saturation<Limit> → Saturation[T] with limit parameter
- [x] mul_one<Shift> → MulOne with shift parameter
- [ ] rect_base<T> → Rect[T] generic struct
- [ ] point_base<T> → Point[T] generic struct
- [ ] vertex_base<T> → Vertex[T] generic struct
- [ ] row_info<T> → RowInfo[T] generic struct
- [ ] const_row_info<T> → ConstRowInfo[T] generic struct

**Geometry Functions**

- [ ] intersect_rectangles(), unite_rectangles()
- [ ] is_equal_eps() epsilon comparison

**Path Utility Functions**

- [ ] is_vertex(), is_drawing(), is_stop(), is_move_to()
- [ ] is_line_to(), is_curve(), is_curve3(), is_curve4()
- [ ] is_end_poly(), is_close(), is_next_poly()
- [ ] is_cw(), is_ccw(), is_oriented(), is_closed()
- [ ] get_close_flag(), clear_orientation(), get_orientation(), set_orientation()

**Constants**

- [ ] pi constant
- [x] deg2rad(), rad2deg() conversions

#### agg_config.h - Configuration definitions

Go files:
- internal/config/config.go

- [ ] Configuration constants (mostly compile-time in C++)
- [ ] Type overrides mechanism for Go

#### agg_array.h - Dynamic array implementation

Go files:
- internal/array/interfaces.go
- internal/array/algorithms.go
- internal/array/comparators.go
- internal/array/pod_arrays.go
- internal/array/pod_bvector.go
- internal/array/block_allocator.go

**POD Array Types (Templates → Go Generics)**

- [ ] pod_array_adaptor<T> → PodArrayAdaptor[T]
- [ ] pod_auto_array<T, Size> → PodAutoArray[T] with size
- [ ] pod_auto_vector<T, Size> → PodAutoVector[T] with size
- [ ] pod_array<T> → PodArray[T] dynamic array
- [ ] pod_vector<T> → PodVector[T] growable vector
- [ ] pod_bvector<T, S> → PodBVector[T] block vector

**Block Allocator**

- [ ] block_allocator class → BlockAllocator struct
- [ ] allocate() with alignment support
- [ ] block management

**Algorithms (Templates → Go Generics)**

- [ ] quick_sort<Array, Less> → QuickSort[T] with comparator
- [ ] swap_elements<T> → SwapElements[T]
- [ ] remove_duplicates<Array, Equal> → RemoveDuplicates[T]
- [ ] invert_container<Array> → InvertContainer[T]
- [ ] binary_search_pos<Array, Value, Less> → BinarySearchPos[T]
- [ ] range_adaptor<Array> → RangeAdaptor[T]

**Comparison Functions**

- [ ] int_less(), int_greater()
- [ ] unsigned_less(), unsigned_greater()

#### agg_math.h - Mathematical functions and constants

Go files:
- internal/basics/math.go
- internal/basics/constants.go

**Constants**

- [x] vertex_dist_epsilon
- [x] intersection_epsilon

**Geometric Calculations**

- [x] cross_product()
- [x] point_in_triangle()
- [x] calc_distance()
- [x] calc_sq_distance()
- [ ] calc_line_point_distance()
- [ ] calc_segment_point_u()
- [x] calc_segment_point_sq_distance() (2 overloads)
- [x] calc_intersection()
- [ ] intersection_exists()
- [ ] calc_orthogonal()
- [ ] dilate_triangle()
- [x] calc_triangle_area()
- [x] calc_polygon_area<Storage>() → CalcPolygonArea[T]()

**Fast Math**

- [x] fast_sqrt() with lookup tables
- [ ] g_sqrt_table[1024] lookup table
- [ ] g_elder_bit_table[256] lookup table
- [x] besj() Bessel function

---

### Color and Pixel Formats

#### agg_color_gray.h - Grayscale color handling

Go files:
- internal/color/gray.go
- internal/color/conversion.go

**Template Types → Go Generics**

- [ ] gray8T<Colorspace> → Gray8[CS] generic struct
- [ ] gray16T<Colorspace> → Gray16[CS] generic struct
- [ ] gray32T<Colorspace> → Gray32[CS] generic struct

**Core Gray8 Methods**

- [ ] luminance(rgba) - ITU-R BT.709 calculation
- [ ] luminance(rgba8) - Optimized 8-bit version
- [ ] convert() methods between colorspaces
- [ ] convert() from/to RGBA types
- [ ] convert_from_sRGB() → ConvertFromSRGB()
- [ ] convert_to_sRGB() → ConvertToSRGB()
- [ ] make_rgba8(), make_srgba8(), make_rgba16(), make_rgba32()
- [ ] Constructors and operators

**Gray16 and Gray32 Variants**

- [ ] gray16 type with 16-bit precision
- [ ] gray32 type with 32-bit precision
- [ ] Conversion methods for each precision

#### agg_color_rgba.h - RGBA color handling

Go files:
- internal/color/rgba.go
- internal/color/rgb.go
- internal/color/conversion.go

**Order Structs (Component Ordering)**

- [ ] order_rgb → OrderRGB constants
- [ ] order_bgr → OrderBGR constants  
- [ ] order_rgba → OrderRGBA constants
- [ ] order_argb → OrderARGB constants
- [ ] order_abgr → OrderABGR constants
- [ ] order_bgra → OrderBGRA constants

**Colorspace Tags**

- [ ] linear struct → Linear type tag
- [ ] sRGB struct → SRGB type tag

**Base RGBA Type (float64)**

- [ ] rgba struct → RGBA base type
- [ ] clear(), transparent(), opacity() methods
- [ ] premultiply(), demultiply() methods
- [ ] gradient() interpolation
- [ ] Operators: +=, _=, +, _
- [ ] no_color() static method
- [ ] from_wavelength() static method

**Template Types → Go Generics**

- [ ] rgba8T<Colorspace> → RGBA8[CS] generic struct
- [ ] rgba16T<Colorspace> → RGBA16[CS] generic struct
- [ ] rgba32T<Colorspace> → RGBA32[CS] generic struct

**RGBA8 Core Methods**

- [ ] convert() between colorspaces (linear ↔ sRGB)
- [ ] convert() to/from float rgba
- [ ] premultiply(), demultiply() operations
- [ ] gradient() interpolation
- [ ] clear(), transparent() methods
- [ ] add(), subtract(), multiply() blend operations
- [ ] apply_gamma_dir(), apply_gamma_inv()

**RGBA16 and RGBA32 Variants**

- [ ] 16-bit and 32-bit precision versions
- [ ] Corresponding conversion methods

**Helper Functions**

- [ ] rgba_pre() → RGBAPre() premultiplied constructor
- [ ] rgb_conv_rgba8() → RGBConvRGBA8()
- [ ] rgb_conv_rgba16() → RGBConvRGBA16()

**sRGB Conversion Tables**

- [ ] sRGB_conv<T> → SRGBConv[T] conversion utilities
- [ ] Lookup tables for sRGB ↔ linear conversion

#### agg_pixfmt_base.h - Base pixel format definitions

Go files:
- internal/pixfmt/base.go

**Pixel Format Tags**

- [ ] pixfmt_gray_tag → PixFmtGrayTag
- [ ] pixfmt_rgb_tag → PixFmtRGBTag
- [ ] pixfmt_rgba_tag → PixFmtRGBATag

**Base Blender Template → Go Generic**

- [ ] blender_base<ColorT, Order> → BlenderBase[C, O]
- [ ] get() methods for pixel extraction
- [ ] set() methods for pixel setting

#### agg_pixfmt_gray.h - Grayscale pixel formats

Go files:
- internal/pixfmt/pixfmt_gray.go
- internal/pixfmt/pixfmt_gray16.go
- internal/pixfmt/pixfmt_gray32.go
- internal/pixfmt/blender_gray.go
- internal/pixfmt/blender_gray16.go
- internal/pixfmt/blender_gray32.go
- internal/pixfmt/gamma_gray.go

**Blender Types**

- [ ] blender_gray<ColorT> → BlenderGray[C]
- [ ] blender_gray_pre<ColorT> → BlenderGrayPre[C]
- [ ] blend_pix() methods for both

**Gamma Application**

- [ ] apply_gamma_dir_gray<ColorT, GammaLut> → ApplyGammaDirGray[C]
- [ ] apply_gamma_inv_gray<ColorT, GammaLut> → ApplyGammaInvGray[C]

**Main Pixel Format Template**

- [ ] pixfmt_alpha_blend_gray<Blender, RenBuf> → PixFmtAlphaBlendGray[B]
- [ ] Core pixel operations (copy_pixel, blend_pixel, etc.)
- [ ] Span operations (copy_hline, blend_hline, etc.)
- [ ] copy_from() for buffer copying

**Concrete Types**

- [ ] pixfmt_gray8 → PixFmtGray8
- [ ] pixfmt_sgray8 → PixFmtSGray8
- [ ] pixfmt_gray16 → PixFmtGray16
- [ ] pixfmt_gray32 → PixFmtGray32

#### agg_pixfmt_rgb.h - RGB pixel formats

Go files:
- internal/pixfmt/pixfmt_rgb.go
- internal/pixfmt/blender_rgb.go

**Gamma Application Classes**

- [ ] apply_gamma_dir_rgb<ColorT, Order, GammaLut>
- [ ] apply_gamma_inv_rgb<ColorT, Order, GammaLut>

**Blender Types**

- [ ] blender_rgb<ColorT, Order> → BlenderRGB[C, O]
- [ ] blender_rgb_pre<ColorT, Order> → BlenderRGBPre[C, O]
- [ ] blender_rgb_gamma<ColorT, Order, Gamma> → BlenderRGBGamma[C, O]

**Main Pixel Format Template**

- [ ] pixfmt_alpha_blend_rgb<Blender, RenBuf, Step, Offset>
- [ ] pixel_type nested struct → RGBPixelType
- [ ] row_data(), make_pix(), copy_pixel(), blend_pixel()
- [ ] Hline operations (copy_hline, blend_hline, etc.)
- [ ] Solid color operations (fill, blend*solid*\*)
- [ ] copy_from(), blend_from() for compositing

**Concrete RGB24 Types**

- [ ] pixfmt_rgb24 → PixFmtRGB24
- [ ] pixfmt_bgr24 → PixFmtBGR24  
- [ ] pixfmt_srgb24 → PixFmtSRGB24
- [ ] pixfmt_sbgr24 → PixFmtSBGR24

**RGB48 Types (16-bit per channel)**

- [ ] pixfmt_rgb48 → PixFmtRGB48
- [ ] pixfmt_bgr48 → PixFmtBGR48

**Gamma Variants**

- [ ] pixfmt_rgb24_gamma<Gamma> → PixFmtRGB24Gamma[G]
- [ ] Similar for all RGB formats

#### agg_pixfmt_rgb_packed.h - Packed RGB pixel formats

**Packed Formats (555, 565, etc.)**

- [ ] pixfmt_rgb555 → PixFmtRGB555
- [ ] pixfmt_rgb565 → PixFmtRGB565
- [ ] pixfmt_bgr555 → PixFmtBGR555
- [ ] pixfmt_bgr565 → PixFmtBGR565
- [ ] Packing/unpacking utilities
- [ ] Bit-shift operations for packed formats

#### agg_pixfmt_rgba.h - RGBA pixel formats

Go files:
- internal/pixfmt/pixfmt_rgba.go
- internal/pixfmt/blender_rgba.go
- internal/pixfmt/gamma_rgba.go

**Blender Types**

- [ ] blender_rgba<ColorT, Order> → BlenderRGBA[C, O]
- [ ] blender_rgba_pre<ColorT, Order> → BlenderRGBAPre[C, O]
- [ ] blender_rgba_plain<ColorT, Order> → BlenderRGBAPlain[C, O]
- [ ] Composite blenders (multiply, screen, overlay, etc.)

**Main RGBA Pixel Format**

- [ ] pixfmt_alpha_blend_rgba<Blender, RenBuf>
- [ ] Full alpha channel support
- [ ] Premultiplied/non-premultiplied operations

**Concrete RGBA32 Types**

- [ ] pixfmt_rgba32 → PixFmtRGBA32
- [ ] pixfmt_argb32 → PixFmtARGB32
- [ ] pixfmt_abgr32 → PixFmtABGR32
- [ ] pixfmt_bgra32 → PixFmtBGRA32

**RGBA64 Types (16-bit per channel)**

- [ ] pixfmt_rgba64 → PixFmtRGBA64
- [ ] pixfmt_argb64 → PixFmtARGB64
- [ ] Similar variants

#### agg_pixfmt_transposer.h - Pixel format transposer

**Transposer Wrapper**

- [ ] pixfmt_transposer<PixFmt> → PixFmtTransposer[P]
- [ ] Transposes x/y coordinates
- [ ] Wraps another pixel format

#### agg_pixfmt_amask_adaptor.h - Alpha mask adaptor

**Alpha Mask Adaptor**

- [ ] pixfmt_amask_adaptor<PixFmt, AlphaMask> → PixFmtAMaskAdaptor[P, A]
- [ ] Applies alpha mask to pixel format operations
- [ ] combine_pixel() with mask

---

### Rendering Buffer

#### agg_rendering_buffer.h

Go files:
- internal/buffer/rendering_buffer.go
- internal/buffer/rendering_buffer_dynarow.go
- internal/config/config.go (type selection)

**row_accessor<T> template class:**

- [ ] Default constructor
- [ ] Parameterized constructor (buf, width, height, stride)
- [ ] attach() method
- [ ] buf() accessor methods (const and non-const)
- [ ] width() accessor method
- [ ] height() accessor method
- [ ] stride() accessor method
- [ ] stride_abs() accessor method
- [ ] row_ptr(int, int y, unsigned) method
- [ ] row_ptr(int y) method (const and non-const)
- [ ] row() method returning row_data
- [ ] copy_from() template method
- [ ] clear() method
- [ ] Private member variables (m_buf, m_start, m_width, m_height, m_stride)

**row_ptr_cache<T> template class:**

- [ ] Default constructor
- [ ] Parameterized constructor (buf, width, height, stride)
- [ ] attach() method with row pointer caching
- [ ] buf() accessor methods (const and non-const)
- [ ] width() accessor method
- [ ] height() accessor method
- [ ] stride() accessor method
- [ ] stride_abs() accessor method
- [ ] row_ptr(int, int y, unsigned) method
- [ ] row_ptr(int y) method (const and non-const)
- [ ] row() method returning row_data
- [ ] rows() method returning row pointer array
- [ ] copy_from() template method
- [ ] clear() method
- [ ] Private member variables (m_buf, m_rows, m_width, m_height, m_stride)

**Type definitions:**

- [ ] rendering_buffer typedef (configurable between row_accessor and row_ptr_cache)

#### agg_rendering_buffer_dynarow.h

Go files:
- internal/buffer/rendering_buffer_dynarow.go

**rendering_buffer_dynarow class:**

- [ ] Destructor
- [ ] Default constructor
- [ ] Parameterized constructor (width, height, byte_width)
- [ ] init() method with memory management
- [ ] width() accessor method
- [ ] height() accessor method
- [ ] byte_width() accessor method
- [ ] row_ptr(int x, int y, unsigned len) method with dynamic allocation
- [ ] row_ptr(int y) const method
- [ ] row_ptr(int y) non-const method
- [ ] row(int y) method returning row_data
- [ ] Private member variables (m_rows, m_width, m_height, m_byte_width)
- [ ] Copy constructor and assignment operator (prohibited)

**Template adaptation considerations:**

- [ ] Design Go generics approach for template types
- [ ] Consider interface-based design for type flexibility
- [ ] Implement concrete types for common use cases (uint8)

---

### Scanlines

#### agg_scanline_bin.h

Go files:
- internal/scanline/scanline_bin.go

**scanline_bin class:**

- [ ] span struct (x, len members)
- [ ] coord_type typedef
- [ ] const_iterator typedef
- [ ] Default constructor
- [ ] reset(min_x, max_x) method
- [ ] add_cell(x, cover) method
- [ ] add_span(x, len, cover) method
- [ ] add_cells(x, len, covers) method
- [ ] finalize(y) method
- [ ] reset_spans() method
- [ ] y() accessor method
- [ ] num_spans() accessor method
- [ ] begin() accessor method
- [ ] Private members (m_last_x, m_y, m_spans, m_cur_span)
- [ ] Copy constructor and assignment operator (prohibited)

**scanline32_bin class:**

- [ ] span struct with constructors
- [ ] coord_type typedef
- [ ] span_array_type typedef
- [ ] const_iterator nested class
- [ ] Default constructor
- [ ] reset(min_x, max_x) method
- [ ] add_cell(x, cover) method
- [ ] add_span(x, len, cover) method
- [ ] add_cells(x, len, covers) method
- [ ] finalize(y) method
- [ ] reset_spans() method
- [ ] y() accessor method
- [ ] num_spans() accessor method
- [ ] begin() accessor method
- [ ] Private members (m_max_len, m_last_x, m_y, m_spans)
- [ ] Copy constructor and assignment operator (prohibited)

#### agg_scanline_p.h

Go files:
- internal/scanline/scanline_p8.go

**scanline_p8 class:**

- [ ] self_type typedef
- [ ] cover_type typedef (int8u)
- [ ] coord_type typedef (int16)
- [ ] span struct (x, len, covers pointer)
- [ ] iterator and const_iterator typedefs
- [ ] Default constructor
- [ ] reset(min_x, max_x) method with memory allocation
- [ ] add_cell(x, cover) method
- [ ] add_cells(x, len, covers) method with memcpy
- [ ] add_span(x, len, cover) method for solid spans
- [ ] finalize(y) method
- [ ] reset_spans() method
- [ ] y() accessor method
- [ ] num_spans() accessor method
- [ ] begin() accessor method
- [ ] Private members (m_last_x, m_y, m_covers, m_cover_ptr, m_spans, m_cur_span)
- [ ] Copy constructor and assignment operator (prohibited)

**scanline32_p8 class:**

- [ ] self_type typedef
- [ ] cover_type typedef (int8u)
- [ ] coord_type typedef (int32)
- [ ] span struct with constructors
- [ ] span_array_type typedef
- [ ] const_iterator nested class
- [ ] Default constructor
- [ ] reset(min_x, max_x) method
- [ ] add_cell(x, cover) method
- [ ] add_cells(x, len, covers) method
- [ ] add_span(x, len, cover) method
- [ ] finalize(y) method
- [ ] reset_spans() method
- [ ] y() accessor method
- [ ] num_spans() accessor method
- [ ] begin() accessor method
- [ ] Private members for 32-bit coordinate handling
- [ ] Copy constructor and assignment operator (prohibited)

#### agg_scanline_u.h

**scanline_u8 class:**

- [ ] self_type typedef
- [ ] cover_type typedef (int8u)
- [ ] coord_type typedef (int16)
- [ ] span struct (x, len, covers array pointer)
- [ ] iterator and const_iterator typedefs
- [ ] Default constructor
- [ ] reset(min_x, max_x) method
- [ ] add_cell(x, cover) method
- [ ] add_cells(x, len, covers) method
- [ ] add_span(x, len, cover) method
- [ ] finalize(y) method
- [ ] reset_spans() method
- [ ] y() accessor method
- [ ] num_spans() accessor method
- [ ] begin() accessor method
- [ ] Private members (m_min_x, m_last_x, m_y, m_covers, m_spans, m_cur_span)

**scanline32_u8 class:**

- [ ] Similar structure adapted for 32-bit coordinates
- [ ] 32-bit typedefs and member variables
- [ ] All corresponding methods adapted for larger coordinate space

#### agg_scanline_storage_aa.h

**scanline_cell_storage<T> template class:**

- [ ] extra_span struct (len, ptr members)
- [ ] value_type typedef
- [ ] Destructor with memory cleanup
- [ ] Default constructor
- [ ] Copy constructor with deep copy
- [ ] Assignment operator with proper cleanup
- [ ] remove_all() method
- [ ] add_cells(cells, num_cells) method with dynamic allocation
- [ ] operator[] const overload for cell access
- [ ] operator[] non-const overload for cell access
- [ ] copy_extra_storage() private helper method
- [ ] Private members (m_cells, m_extra_storage)

**scanline_storage_aa class:**

- [ ] Embedded span and scanline structs
- [ ] Constructor and destructor
- [ ] min_x(), max_x() accessor methods
- [ ] reset(min_x, max_x) method
- [ ] add_cells() method
- [ ] finalize() method
- [ ] size() accessor method
- [ ] operator[] for scanline access
- [ ] Memory management methods

**scanline_storage_aa8 typedef:**

- [ ] Concrete instantiation for int8u cover type

#### agg_scanline_storage_bin.h

**scanline_storage_bin class:**

- [ ] Similar structure to AA storage but for binary scanlines
- [ ] span and scanline structs for binary data
- [ ] Constructor and destructor
- [ ] reset() method
- [ ] add_span() method
- [ ] finalize() method
- [ ] Access methods for binary scanline data

#### agg_scanline_boolean_algebra.h

**Boolean operation functors (all template-based):**

- [ ] sbool_combine_spans_bin template functor
- [ ] sbool_combine_spans_empty template functor
- [ ] sbool_add_span_empty template functor
- [ ] sbool_add_span_bin template functor
- [ ] sbool_add_span_aa template functor
- [ ] sbool_intersect_spans_aa template functor with cover_scale_e enum
- [ ] sbool_unite_spans_aa template functor
- [ ] sbool_xor_spans_aa template functor
- [ ] sbool_subtract_spans_aa template functor
- [ ] Additional boolean operation functors

**Main algorithm templates:**

- [ ] sbool_intersect_shapes template function
- [ ] sbool_unite_shapes template function
- [ ] sbool_xor_shapes template function
- [ ] sbool_subtract_shapes template function

**Template adaptation considerations:**

- [ ] Convert C++ functors to Go function types or interfaces
- [ ] Adapt template parameters to Go generics or concrete types
- [ ] Handle iterator patterns with Go-idiomatic approaches
- [ ] Memory management adaptation for Go's garbage collector

---

### Rasterizers

#### agg_rasterizer_cells_aa.h

**rasterizer_cells_aa<Cell> template class:**

- [ ] cell_block_scale_e enum (cell_block_shift, cell_block_size, cell_block_mask, cell_block_pool)
- [ ] sorted_y struct (start, num members)
- [ ] cell_type typedef
- [ ] self_type typedef
- [ ] Destructor with block memory cleanup
- [ ] Constructor with cell_block_limit parameter
- [ ] reset() method
- [ ] style(style_cell) method
- [ ] line(x1, y1, x2, y2) method for line rasterization
- [ ] min_x() accessor method
- [ ] min_y() accessor method
- [ ] max_x() accessor method
- [ ] max_y() accessor method
- [ ] sort_cells() method
- [ ] total_cells() accessor method
- [ ] scanline_num_cells(y) method
- [ ] scanline_cells(y) method
- [ ] sorted() accessor method
- [ ] set_curr_cell(x, y) private method
- [ ] add_curr_cell() private method
- [ ] render_hline() private method
- [ ] allocate_block() private method
- [ ] Private members (m_num_blocks, m_max_blocks, m_curr_block, m_num_cells, etc.)
- [ ] Copy constructor and assignment operator (prohibited)

#### agg_rasterizer_scanline_aa.h

**rasterizer_scanline_aa<Clip> template class:**

- [ ] status enum (status_initial, status_move_to, status_line_to, status_closed)
- [ ] clip_type typedef
- [ ] conv_type typedef
- [ ] coord_type typedef
- [ ] aa_scale_e enum (aa_shift, aa_scale, aa_mask, aa_scale2, aa_mask2)
- [ ] Default constructor with cell_block_limit
- [ ] Template constructor with gamma function
- [ ] reset() method
- [ ] reset_clipping() method
- [ ] clip_box(x1, y1, x2, y2) method
- [ ] filling_rule(filling_rule) method
- [ ] auto_close(flag) method
- [ ] gamma() template method for gamma correction
- [ ] apply_gamma(cover) method
- [ ] move_to(x, y) method (integer coordinates)
- [ ] line_to(x, y) method (integer coordinates)
- [ ] move_to_d(x, y) method (double coordinates)
- [ ] line_to_d(x, y) method (double coordinates)
- [ ] close_polygon() method
- [ ] add_path() template method
- [ ] add_vertex(x, y, cmd) method
- [ ] edge(x1, y1, x2, y2) method
- [ ] edge_d(x1, y1, x2, y2) method
- [ ] sort() method
- [ ] rewind_scanlines() method
- [ ] calculate_alpha() method
- [ ] sweep_scanline() template method
- [ ] navigate_scanline(y) method
- [ ] hit_test(tx, ty) method
- [ ] Private members (m_outline, m_clipper, m_filling_rule, m_gamma, etc.)

#### agg_rasterizer_scanline_aa_nogamma.h

**rasterizer_scanline_aa_nogamma<Clip> template class:**

- [ ] Similar structure to rasterizer_scanline_aa but without gamma correction
- [ ] Simplified apply_gamma() method (no gamma table)
- [ ] All other methods matching rasterizer_scanline_aa interface
- [ ] Performance-optimized implementation

#### agg_rasterizer_compound_aa.h

**cell_style_aa struct:**

- [ ] Position members (x, y)
- [ ] Coverage members (cover, area)
- [ ] Style members (left, right)
- [ ] initial() method
- [ ] style(c) method
- [ ] not_equal(ex, ey, c) method

**layer_order_e enum:**

- [ ] layer_unsorted constant
- [ ] layer_direct constant
- [ ] layer_inverse constant

**rasterizer_compound_aa<Clip> template class:**

- [ ] style_info struct (start_cell, num_cells, last_x)
- [ ] cell_info struct (x, area, cover)
- [ ] clip_type typedef
- [ ] conv_type typedef
- [ ] coord_type typedef
- [ ] aa_scale_e enum constants
- [ ] Default constructor
- [ ] reset() method
- [ ] reset_clipping() method
- [ ] clip_box(x1, y1, x2, y2) method
- [ ] filling_rule(filling_rule) method
- [ ] layer_order(order) method
- [ ] styles(left, right) method
- [ ] move_to(x, y) method
- [ ] line_to(x, y) method
- [ ] move_to_d(x, y) method
- [ ] line_to_d(x, y) method
- [ ] add_vertex(x, y, cmd) method
- [ ] edge(x1, y1, x2, y2) method
- [ ] edge_d(x1, y1, x2, y2) method
- [ ] sort() method
- [ ] navigate_scanline(y) method
- [ ] hit_test(tx, ty) method
- [ ] allocate_master_alpha() method
- [ ] sweep_styles() method
- [ ] scanline_start() method
- [ ] scanline_length() method
- [ ] style(style_id) method
- [ ] Private members for style and layer management

#### agg_rasterizer_sl_clip.h

**Coordinate conversion structs:**

- [ ] ras_conv_int struct (coord_type typedef, mul_div, xi, yi, upscale, downscale methods)
- [ ] ras_conv_int_sat struct (saturated integer conversion)
- [ ] ras_conv_int_3x struct (3x integer conversion for sub-pixel rendering)
- [ ] ras_conv_dbl struct (double precision conversion)
- [ ] ras_conv_dbl_3x struct (3x double conversion)

**Clipping template classes:**

- [ ] rasterizer_sl_no_clip<Conv> template class
- [ ] rasterizer_sl_clip_int<Conv> template class
- [ ] rasterizer_sl_clip_int_sat<Conv> template class
- [ ] rasterizer_sl_clip_int_3x<Conv> template class
- [ ] rasterizer_sl_clip_dbl<Conv> template class
- [ ] rasterizer_sl_clip_dbl_3x<Conv> template class

**Each clipping class includes:**

- [ ] conv_type typedef
- [ ] coord_type typedef
- [ ] Constructor with clipping bounds
- [ ] reset_clipping() method
- [ ] clip_box() method
- [ ] move_to() method
- [ ] line_to() method
- [ ] Private clipping implementation

#### agg_rasterizer_outline.h

**rasterizer_outline<Renderer> template class:**

- [ ] renderer_type typedef
- [ ] coord_type typedef
- [ ] Constructor with renderer
- [ ] attach(renderer) method
- [ ] filling_rule(filling_rule) method
- [ ] gamma() method
- [ ] reset() method
- [ ] move_to(x, y) method
- [ ] line_to(x, y) method
- [ ] move_to_d(x, y) method
- [ ] line_to_d(x, y) method
- [ ] close_polygon() method
- [ ] add_path() template method
- [ ] add_vertex(x, y, cmd) method
- [ ] Private outline rendering implementation

#### agg_rasterizer_outline_aa.h

**rasterizer_outline_aa<Renderer> template class:**

- [ ] Similar structure to rasterizer_outline but with anti-aliasing
- [ ] Enhanced line rendering with coverage calculation
- [ ] Anti-aliased endpoint handling
- [ ] Smooth line joining algorithms

**Template adaptation considerations:**

- [ ] Convert Cell template parameters to Go generics or interfaces
- [ ] Adapt Clip template parameters to interface-based design
- [ ] Convert Renderer template parameters to interface types
- [ ] Replace C++ functors with Go function types
- [ ] Adapt memory management for Go's garbage collector
- [ ] Convert enums to Go constants or typed constants
- [ ] Handle coordinate conversion with Go methods or interfaces

---

### Renderers

#### agg_renderer_base.h - Base renderer template class

**Template Class renderer_base<PixelFormat>**

- [ ] pixfmt_type, color_type, row_data typedefs
- [ ] Default constructor
- [ ] Parameterized constructor with pixel format
- [ ] attach() method for pixel format attachment

**Pixel Format Access**

- [ ] ren() const method - pixel format accessor
- [ ] ren() non-const method - pixel format accessor
- [ ] width() const method
- [ ] height() const method

**Clipping Operations**

- [ ] clip_box(x1, y1, x2, y2) method with bounds checking
- [ ] reset_clipping(visibility) method
- [ ] clip_box_naked(x1, y1, x2, y2) method - no bounds checking
- [ ] inbox(x, y) const method - point-in-clip test

**Clipping Accessors**

- [ ] clip_box() const method
- [ ] xmin(), ymin(), xmax(), ymax() accessors
- [ ] bounding_clip_box() const method
- [ ] bounding_xmin(), bounding_ymin(), bounding_xmax(), bounding_ymax() accessors

**Buffer Operations**

- [ ] clear(color) method - clear entire buffer
- [ ] fill(color) method - fill with blending

**Pixel Operations**

- [ ] copy_pixel(x, y, color) method
- [ ] blend_pixel(x, y, color, cover) method
- [ ] pixel(x, y) const method - get pixel color

**Line Operations**

- [ ] copy_hline(x1, y, x2, color) method
- [ ] copy_vline(x, y1, y2, color) method
- [ ] blend_hline(x1, y, x2, color, cover) method
- [ ] blend_vline(x, y1, y2, color, cover) method

**Rectangle Operations**

- [ ] copy_bar(x1, y1, x2, y2, color) method
- [ ] blend_bar(x1, y1, x2, y2, color, cover) method

**Span Operations**

- [ ] blend_solid_hspan(x, y, len, color, covers) method
- [ ] blend_solid_vspan(x, y, len, color, covers) method
- [ ] copy_color_hspan(x, y, len, colors) method
- [ ] copy_color_vspan(x, y, len, colors) method
- [ ] blend_color_hspan(x, y, len, colors, covers, cover) method
- [ ] blend_color_vspan(x, y, len, colors, covers, cover) method

**Buffer Copying**

- [ ] copy_from() template method for buffer-to-buffer copying

#### agg_renderer_scanline.h - Scanline rendering functions and classes

**Free Functions**

- [ ] render_scanline_aa_solid<Scanline, BaseRenderer, ColorT>() function
- [ ] render_scanlines_aa_solid<Rasterizer, Scanline, BaseRenderer, ColorT>() function
- [ ] render_scanline_aa<Scanline, BaseRenderer, SpanAllocator, SpanGenerator>() function
- [ ] render_scanlines_aa<Rasterizer, Scanline, BaseRenderer, SpanAllocator, SpanGenerator>() function
- [ ] render_scanline_bin_solid<Scanline, BaseRenderer, ColorT>() function
- [ ] render_scanlines_bin_solid<Rasterizer, Scanline, BaseRenderer, ColorT>() function

**Template Class renderer_scanline_aa_solid<BaseRenderer>**

- [ ] base_ren_type, color_type typedefs
- [ ] Constructor with base renderer
- [ ] attach(base_ren) method
- [ ] color(color) setter method
- [ ] color() const getter method
- [ ] prepare() method
- [ ] render<Scanline>(scanline) template method

**Template Class renderer_scanline_aa<BaseRenderer, SpanAllocator, SpanGenerator>**

- [ ] base_ren_type, alloc_type, span_gen_type typedefs
- [ ] Constructor with base renderer
- [ ] attach(base_ren, span_allocator, span_generator) method
- [ ] prepare() method
- [ ] render<Scanline>(scanline) template method

**Template Class renderer_scanline_bin_solid<BaseRenderer>**

- [ ] base_ren_type, color_type typedefs
- [ ] Constructor and attach method
- [ ] color management methods
- [ ] prepare() method
- [ ] render<Scanline>(scanline) template method for binary scanlines

**Template Class renderer_scanline_bin<BaseRenderer, SpanAllocator, SpanGenerator>**

- [ ] Similar structure to renderer_scanline_aa but for binary scanlines
- [ ] Base renderer and span generator management
- [ ] Binary scanline rendering

#### agg_renderer_primitives.h - Primitive drawing operations

**Template Class renderer_primitives<BaseRenderer>**

- [ ] base_ren_type, color_type typedefs
- [ ] Constructor with base renderer
- [ ] attach(base_ren) method

**Color Management**

- [ ] fill_color(color) setter method
- [ ] line_color(color) setter method
- [ ] fill_color() const getter method
- [ ] line_color() const getter method

**Rectangle Operations**

- [ ] rectangle(x1, y1, x2, y2) method - outlined rectangle
- [ ] solid_rectangle(x1, y1, x2, y2) method - filled rectangle
- [ ] outlined_rectangle(x1, y1, x2, y2) method - outlined with different line color

**Ellipse Operations**

- [ ] ellipse(x, y, rx, ry) method - outlined ellipse with Bresenham algorithm
- [ ] solid_ellipse(x, y, rx, ry) method - filled ellipse
- [ ] outlined_ellipse(x, y, rx, ry) method - outlined with different line color

**Line Drawing**

- [ ] line(x1, y1, x2, y2, last) method using DDA algorithm
- [ ] move_to(x, y) method for path building
- [ ] line_to(x, y, last) method for path building

**Accessors**

- [ ] ren() const method - base renderer accessor
- [ ] rbuf() const method - rendering buffer accessor

**Private Members**

- [ ] m_ren pointer to base renderer
- [ ] m_fill_color member
- [ ] m_line_color member
- [ ] m_curr_x, m_curr_y current position members

#### agg_renderer_markers.h - Marker shape rendering

**Template Class renderer_markers<BaseRenderer> (inherits from renderer_primitives)**

- [ ] base_type, base_ren_type, color_type typedefs
- [ ] Inheritance from renderer_primitives<BaseRenderer>

**Visibility and Basic Operations**

- [ ] visible(x, y, r) const method - visibility test within bounds

**Basic Shape Markers**

- [ ] square(x, y, r) method - solid square marker
- [ ] diamond(x, y, r) method - solid diamond marker
- [ ] circle(x, y, r) method - solid circle marker using ellipse algorithm
- [ ] crossed_circle(x, y, r) method - circle with cross pattern

**Semi-ellipse Markers (Direction-specific)**

- [ ] semiellipse_left(x, y, r) method
- [ ] semiellipse_right(x, y, r) method
- [ ] semiellipse_up(x, y, r) method
- [ ] semiellipse_down(x, y, r) method

**Triangle Markers (Direction-specific)**

- [ ] triangle_left(x, y, r) method
- [ ] triangle_right(x, y, r) method
- [ ] triangle_up(x, y, r) method
- [ ] triangle_down(x, y, r) method

**Ray and Line Markers**

- [ ] four_rays(x, y, r) method - plus sign pattern
- [ ] cross(x, y, r) method - diagonal cross pattern
- [ ] x(x, y, r) method - X pattern
- [ ] dash(x, y, r) method - horizontal dash
- [ ] dot(x, y, r) method - small filled circle
- [ ] pixel(x, y, color) method - single pixel marker

#### agg_renderer_outline_aa.h - Anti-aliased outline rendering

**Template Class renderer_outline_aa<BaseRenderer>**

- [ ] base_ren_type, color_type, coord_type typedefs
- [ ] Constructor with base renderer
- [ ] attach(base_ren) method

**Line Pattern Support**

- [ ] pattern(line_pattern) method
- [ ] pattern() const getter method
- [ ] pattern_scale() setter method
- [ ] pattern_scale() const getter method
- [ ] pattern_start() setter method

**Line Join and Cap Settings**

- [ ] line_join(join_type) method - miter, round, bevel
- [ ] line_cap(cap_type) method - butt, square, round
- [ ] inner_join(join_type) method
- [ ] width(line_width) setter method
- [ ] width() const getter method

**Rendering Methods**

- [ ] move_to(x, y) method
- [ ] line_to(x, y) method
- [ ] move_to_d(x, y) method - double precision
- [ ] line_to_d(x, y) method - double precision
- [ ] close_polygon() method
- [ ] add_path<VertexSource>(vs, path_id) template method
- [ ] add_vertex(x, y, cmd) method

**Accuracy Control**

- [ ] accuracy(approximation_scale) setter method
- [ ] accuracy() const getter method

#### agg_renderer_outline_image.h - Image-based outline rendering

**Template Class renderer_outline_image<BaseRenderer, ImagePattern>**

- [ ] base_ren_type, color_type, order_type typedefs
- [ ] pattern_type typedef
- [ ] Constructor with base renderer and pattern
- [ ] attach(base_ren) method

**Pattern Management**

- [ ] pattern(image_pattern) setter method
- [ ] pattern() const getter method
- [ ] pattern_scale_x(), pattern_scale_y() setters
- [ ] pattern_scale() unified setter method

**Rendering Methods**

- [ ] move_to(x, y) method
- [ ] line_to(x, y) method
- [ ] move_to_d(x, y) method
- [ ] line_to_d(x, y) method
- [ ] Pattern-based line stroke rendering

**Image Pattern Application**

- [ ] Subpixel pattern positioning
- [ ] Pattern scaling and rotation
- [ ] Pattern tiling along line path

#### agg_renderer_mclip.h - Multi-clipping renderer

**Template Class renderer_mclip<PixelFormat>**

- [ ] pixfmt_type, color_type typedefs
- [ ] base_ren_type typedef
- [ ] Constructor with pixel format
- [ ] attach(pixfmt) method

**Clipping Region Management**

- [ ] first_clip_box() method
- [ ] add_clip_box(x1, y1, x2, y2) method
- [ ] remove_last_clip_box() method
- [ ] clear_clip_boxes() method
- [ ] clip_box_count() const method

**Multi-region Clipping Operations**

- [ ] copy_pixel(x, y, color) method with multi-clip
- [ ] blend_pixel(x, y, color, cover) method with multi-clip
- [ ] copy_hline(x1, y, x2, color) method with multi-clip
- [ ] blend_hline(x1, y, x2, color, cover) method with multi-clip
- [ ] copy_vline(x, y1, y2, color) method with multi-clip
- [ ] blend_vline(x, y1, y2, color, cover) method with multi-clip

**Clipping Logic**

- [ ] inbox_all(x, y) const method - point in all clips
- [ ] inbox_any(x, y) const method - point in any clip
- [ ] Intersection and union clipping operations

#### agg_renderer_raster_text.h - Raster text rendering

**Template Class renderer_raster_text<BaseRenderer, GlyphRasterizer>**

- [ ] base_ren_type, color_type typedefs
- [ ] glyph_ras_type, glyph_type typedefs
- [ ] Constructor with base renderer and glyph rasterizer
- [ ] attach(base_ren, glyph_ras) method

**Text Rendering**

- [ ] render_text(x, y, text_string) method
- [ ] render_glyph(x, y, glyph) method
- [ ] Character positioning and advancement

**Font and Style Management**

- [ ] color(text_color) setter method
- [ ] color() const getter method
- [ ] font_size(size) setter method
- [ ] font_height() const getter method
- [ ] baseline() const getter method

**Text Positioning**

- [ ] move_to(x, y) method
- [ ] text_out(text_string) method
- [ ] Horizontal and vertical alignment support
- [ ] Character and line spacing controls

**Glyph Operations**

- [ ] Embedded raster font support
- [ ] Glyph caching and reuse
- [ ] Unicode text support

---

### Geometric Primitives

#### agg_arc.h - Arc generation

**arc class**

- [ ] Default constructor: `arc()`
- [ ] Parameterized constructor: `arc(double x, double y, double rx, double ry, double a1, double a2, bool ccw=true)`
- [ ] init() method: `init(double x, double y, double rx, double ry, double a1, double a2, bool ccw=true)`
- [ ] approximation_scale(double s) setter method
- [ ] approximation_scale() const getter method
- [ ] rewind(unsigned) method for path rewinding
- [ ] vertex(double* x, double* y) method for vertex generation
- [ ] normalize(double a1, double a2, bool ccw) private method for angle normalization

**Member Variables**

- [ ] Position and radius: m_x, m_y, m_rx, m_ry (double)
- [ ] Angle parameters: m_angle, m_start, m_end, m_da (double)
- [ ] Scale parameter: m_scale (double)
- [ ] State flags: m_ccw, m_initialized (bool)
- [ ] Path command: m_path_cmd (unsigned)

#### agg_ellipse.h - Ellipse generation

**ellipse class**

- [ ] Default constructor: `ellipse()`
- [ ] Parameterized constructor: `ellipse(double x, double y, double rx, double ry, unsigned num_steps=0, bool cw=false)`
- [ ] init() inline method: `init(double x, double y, double rx, double ry, unsigned num_steps=0, bool cw=false)`
- [ ] approximation_scale(double scale) inline setter method
- [ ] rewind(unsigned path_id) inline method
- [ ] vertex(double* x, double* y) inline method
- [ ] calc_num_steps() private inline method for step calculation

**Member Variables**

- [ ] Position and radius: m_x, m_y, m_rx, m_ry (double)
- [ ] Scale parameter: m_scale (double)
- [ ] Step tracking: m_num, m_step (unsigned)
- [ ] Direction flag: m_cw (bool)

#### agg_ellipse_bresenham.h - Bresenham ellipse algorithm

**ellipse_bresenham_interpolator class**

- [ ] Constructor: `ellipse_bresenham_interpolator(int rx, int ry)`
- [ ] dx() const getter method
- [ ] dy() const getter method
- [ ] operator++() increment operator for pixel stepping

**Member Variables**

- [ ] Radius squared: m_rx2, m_ry2 (int)
- [ ] Double radius squared: m_two_rx2, m_two_ry2 (int)
- [ ] Current deltas: m_dx, m_dy (int)
- [ ] Increments: m_inc_x, m_inc_y (int)
- [ ] Current function value: m_cur_f (int)

#### agg_rounded_rect.h - Rounded rectangle generation

**rounded_rect class**

- [ ] Default constructor: `rounded_rect()`
- [ ] Parameterized constructor: `rounded_rect(double x1, double y1, double x2, double y2, double r)`
- [ ] rect(double x1, double y1, double x2, double y2) method
- [ ] radius(double r) single radius setter
- [ ] radius(double rx, double ry) x/y radius setter
- [ ] radius(double rx_bottom, double ry_bottom, double rx_top, double ry_top) top/bottom setter
- [ ] radius(double rx1, double ry1, double rx2, double ry2, double rx3, double ry3, double rx4, double ry4) individual corner setter
- [ ] normalize_radius() method
- [ ] approximation_scale(double s) inline setter
- [ ] approximation_scale() const inline getter
- [ ] rewind(unsigned) method
- [ ] vertex(double* x, double* y) method with state machine

**Member Variables**

- [ ] Rectangle bounds: m_x1, m_y1, m_x2, m_y2 (double)
- [ ] Corner radii: m_rx1, m_ry1, m_rx2, m_ry2, m_rx3, m_ry3, m_rx4, m_ry4 (double)
- [ ] State tracking: m_status (unsigned)
- [ ] Composed arc object: m_arc (arc)

#### agg_arrowhead.h - Arrowhead generation

**arrowhead class**

- [ ] Default constructor: `arrowhead()`
- [ ] head(double d1, double d2, double d3, double d4) inline head configuration
- [ ] head() inline getter
- [ ] no_head() inline method to disable head
- [ ] tail(double d1, double d2, double d3, double d4) inline tail configuration
- [ ] tail() inline getter
- [ ] no_tail() inline method to disable tail
- [ ] rewind(unsigned path_id) method
- [ ] vertex(double* x, double* y) method

**Member Variables**

- [ ] Head parameters: m_head_d1, m_head_d2, m_head_d3, m_head_d4 (double)
- [ ] Tail parameters: m_tail_d1, m_tail_d2, m_tail_d3, m_tail_d4 (double)
- [ ] Enable flags: m_head_flag, m_tail_flag (bool)
- [ ] Coordinate array: m_coord[16] (double)
- [ ] Command array: m_cmd[8] (unsigned)
- [ ] Current state: m_curr_id, m_curr_coord (unsigned)

#### Implementation Files (.cpp)

**agg_arc.cpp**

- [ ] Constructor implementation with angle normalization
- [ ] init() method with parameter validation
- [ ] approximation_scale() with scale setting
- [ ] rewind() path reset logic
- [ ] vertex() trigonometric vertex calculation
- [ ] normalize() private angle normalization algorithm

**agg_arrowhead.cpp**

- [ ] Constructor with coordinate and command array initialization
- [ ] rewind() path selection logic (head/tail/both)
- [ ] vertex() coordinate lookup and command generation

**agg_rounded_rect.cpp**

- [ ] Constructor with radius initialization
- [ ] rect() bounds setting with validation
- [ ] radius() methods with various parameter combinations
- [ ] normalize_radius() clamping to valid ranges
- [ ] rewind() state machine initialization
- [ ] vertex() complex state machine for corner generation using composed arc

#### Special Porting Considerations

- [ ] All classes implement the same vertex source interface (rewind/vertex pattern)
- [ ] No templates used - direct struct/class conversion
- [ ] Mathematical dependencies on <cmath> functions (sin, cos, atan2, etc.)
- [ ] State machines in vertex() methods need careful Go translation
- [ ] Inline methods should be regular Go methods
- [ ] Double precision floating point throughout
- [ ] Path command constants from agg_basics.h dependency

---

### Curves and Paths

- [ ] agg_curves.h - Curve approximation
- [ ] agg_bezier_arc.h - Bezier arc
- [ ] agg_bspline.h - B-spline curves
- [ ] agg_path_storage.h - Path storage
- [ ] agg_path_storage_integer.h - Integer path storage
- [ ] agg_path_length.h - Path length calculation

---

### Transformations

- [ ] agg_trans_affine.h - Affine transformations
- [ ] agg_trans_bilinear.h - Bilinear transformations
- [ ] agg_trans_perspective.h - Perspective transformations
- [ ] agg_trans_viewport.h - Viewport transformations
- [ ] agg_trans_single_path.h - Single path transformation
- [ ] agg_trans_double_path.h - Double path transformation
- [ ] agg_trans_warp_magnifier.h - Warp magnifier transformation

---

### Converters

- [ ] agg_conv_adaptor_vcgen.h - Vertex generator adaptor
- [ ] agg_conv_adaptor_vpgen.h - Vertex processor adaptor
- [ ] agg_conv_bspline.h - B-spline converter
- [ ] agg_conv_clip_polygon.h - Polygon clipping converter
- [ ] agg_conv_clip_polyline.h - Polyline clipping converter
- [ ] agg_conv_close_polygon.h - Polygon closing converter
- [ ] agg_conv_concat.h - Path concatenation converter
- [ ] agg_conv_contour.h - Contour converter
- [ ] agg_conv_curve.h - Curve converter
- [ ] agg_conv_dash.h - Dash converter
- [ ] agg_conv_gpc.h - General Polygon Clipper converter
- [ ] agg_conv_marker.h - Marker converter
- [ ] agg_conv_marker_adaptor.h - Marker adaptor converter
- [ ] agg_conv_segmentator.h - Segmentator converter
- [ ] agg_conv_shorten_path.h - Path shortening converter
- [ ] agg_conv_smooth_poly1.h - Polygon smoothing converter
- [ ] agg_conv_stroke.h - Stroke converter
- [ ] agg_conv_transform.h - Transform converter
- [ ] agg_conv_unclose_polygon.h - Polygon unclosing converter

---

### Vertex Generators

- [ ] agg_vcgen_bspline.h - B-spline vertex generator
- [ ] agg_vcgen_contour.h - Contour vertex generator
- [ ] agg_vcgen_dash.h - Dash vertex generator
- [ ] agg_vcgen_markers_term.h - Terminal markers vertex generator
- [ ] agg_vcgen_smooth_poly1.h - Polygon smoothing vertex generator
- [ ] agg_vcgen_stroke.h - Stroke vertex generator
- [ ] agg_vcgen_vertex_sequence.h - Vertex sequence generator

---

### Vertex Processors

- [ ] agg_vpgen_clip_polygon.h - Polygon clipping vertex processor
- [ ] agg_vpgen_clip_polyline.h - Polyline clipping vertex processor
- [ ] agg_vpgen_segmentator.h - Segmentator vertex processor

---

### Spans and Gradients

- [ ] agg_span_allocator.h - Span allocator
- [ ] agg_span_converter.h - Span converter
- [ ] agg_span_solid.h - Solid color span
- [ ] agg_span_gradient.h - Gradient span
- [ ] agg_span_gradient_alpha.h - Alpha gradient span
- [ ] agg_span_gradient_contour.h - Contour gradient span
- [ ] agg_span_gradient_image.h - Image gradient span
- [ ] agg_span_gouraud.h - Gouraud shading span
- [ ] agg_span_gouraud_gray.h - Grayscale Gouraud span
- [ ] agg_span_gouraud_rgba.h - RGBA Gouraud span

---

### Image Processing

- [ ] agg_image_accessors.h - Image accessors
- [ ] agg_image_filters.h - Image filters
- [ ] agg_span_image_filter.h - Image filter span
- [ ] agg_span_image_filter_gray.h - Grayscale image filter span
- [ ] agg_span_image_filter_rgb.h - RGB image filter span
- [ ] agg_span_image_filter_rgba.h - RGBA image filter span

---

### Pattern Processing

- [ ] agg_pattern_filters_rgba.h - RGBA pattern filters
- [ ] agg_span_pattern_gray.h - Grayscale pattern span
- [ ] agg_span_pattern_rgb.h - RGB pattern span
- [ ] agg_span_pattern_rgba.h - RGBA pattern span

---

### Interpolators

- [ ] agg_span_interpolator_adaptor.h - Interpolator adaptor
- [ ] agg_span_interpolator_linear.h - Linear interpolator
- [ ] agg_span_interpolator_persp.h - Perspective interpolator
- [ ] agg_span_interpolator_trans.h - Transform interpolator
- [ ] agg_span_subdiv_adaptor.h - Subdivision adaptor

---

### Utility and Math

- [ ] agg_alpha_mask_u8.h - 8-bit alpha mask
- [ ] agg_bitset_iterator.h - Bitset iterator
- [ ] agg_blur.h - Blur effects
- [ ] agg_bounding_rect.h - Bounding rectangle calculation
- [ ] agg_clip_liang_barsky.h - Liang-Barsky clipping algorithm
- [ ] agg_dda_line.h - DDA line algorithm
- [ ] agg_gamma_functions.h - Gamma correction functions
- [ ] agg_gamma_lut.h - Gamma lookup table
- [ ] agg_gradient_lut.h - Gradient lookup table
- [ ] agg_line_aa_basics.h - Anti-aliased line basics
- [ ] agg_math_stroke.h - Stroke mathematics
- [ ] agg_shorten_path.h - Path shortening
- [ ] agg_simul_eq.h - Simultaneous equations solver
- [ ] agg_vertex_sequence.h - Vertex sequence

---

### Text and Fonts

- [ ] agg_embedded_raster_fonts.h - Embedded raster fonts
- [ ] agg_font_cache_manager.h - Font cache manager
- [ ] agg_font_cache_manager2.h - Font cache manager v2
- [ ] agg_glyph_raster_bin.h - Binary glyph rasterizer
- [ ] agg_gsv_text.h - GSV text rendering

---

### Controls (ctrl/)

- [ ] agg_ctrl.h - Base control class
- [ ] agg_bezier_ctrl.h - Bezier curve control
- [ ] agg_cbox_ctrl.h - Checkbox control
- [ ] agg_gamma_ctrl.h - Gamma control
- [ ] agg_gamma_spline.h - Gamma spline
- [ ] agg_polygon_ctrl.h - Polygon control
- [ ] agg_rbox_ctrl.h - Radio button control
- [ ] agg_scale_ctrl.h - Scale control
- [ ] agg_slider_ctrl.h - Slider control
- [ ] agg_spline_ctrl.h - Spline control

---

### Platform Support (platform/)

- [ ] agg_platform_support.h - Platform support interface

---

### Utilities (util/)

- [ ] agg_color_conv.h - Color conversion utilities
- [ ] agg_color_conv_rgb16.h - 16-bit RGB color conversion
- [ ] agg_color_conv_rgb8.h - 8-bit RGB color conversion

## Core Implementation Files (src/)

### Basic Implementations

- [ ] agg_arc.cpp - Arc generation implementation
- [ ] agg_arrowhead.cpp - Arrowhead implementation
- [ ] agg_bezier_arc.cpp - Bezier arc implementation
- [ ] agg_bspline.cpp - B-spline implementation
- [ ] agg_color_rgba.cpp - RGBA color implementation
- [ ] agg_curves.cpp - Curve approximation implementation
- [ ] agg_embedded_raster_fonts.cpp - Embedded fonts implementation
- [ ] agg_gsv_text.cpp - GSV text implementation
- [ ] agg_image_filters.cpp - Image filters implementation
- [ ] agg_line_aa_basics.cpp - Anti-aliased line basics
- [ ] agg_line_profile_aa.cpp - Anti-aliased line profile
- [ ] agg_rounded_rect.cpp - Rounded rectangle implementation
- [ ] agg_sqrt_tables.cpp - Square root tables
- [ ] agg_trans_affine.cpp - Affine transformation implementation
- [ ] agg_trans_double_path.cpp - Double path transformation
- [ ] agg_trans_single_path.cpp - Single path transformation
- [ ] agg_trans_warp_magnifier.cpp - Warp magnifier implementation

### Vertex Generators

- [ ] agg_vcgen_bspline.cpp - B-spline vertex generator
- [ ] agg_vcgen_contour.cpp - Contour vertex generator
- [ ] agg_vcgen_dash.cpp - Dash vertex generator
- [ ] agg_vcgen_markers_term.cpp - Terminal markers implementation
- [ ] agg_vcgen_smooth_poly1.cpp - Polygon smoothing implementation
- [ ] agg_vcgen_stroke.cpp - Stroke vertex generator

### Vertex Processors

- [ ] agg_vpgen_clip_polygon.cpp - Polygon clipping implementation
- [ ] agg_vpgen_clip_polyline.cpp - Polyline clipping implementation
- [ ] agg_vpgen_segmentator.cpp - Segmentator implementation

### Controls Implementation (src/ctrl/)

- [ ] agg_bezier_ctrl.cpp - Bezier control implementation
- [ ] agg_cbox_ctrl.cpp - Checkbox control implementation
- [ ] agg_gamma_ctrl.cpp - Gamma control implementation
- [ ] agg_gamma_spline.cpp - Gamma spline implementation
- [ ] agg_polygon_ctrl.cpp - Polygon control implementation
- [ ] agg_rbox_ctrl.cpp - Radio button implementation
- [ ] agg_scale_ctrl.cpp - Scale control implementation
- [ ] agg_slider_ctrl.cpp - Slider control implementation
- [ ] agg_spline_ctrl.cpp - Spline control implementation

## AGG2D High-Level Interface

- [ ] agg2d.h - AGG2D header
- [ ] agg2d.cpp - AGG2D implementation

## Font Support

### FreeType Integration

- [ ] agg_font_freetype.h - FreeType font support header
- [ ] agg_font_freetype.cpp - FreeType font implementation
- [ ] agg_font_freetype2.h - FreeType v2 support header
- [ ] agg_font_freetype2.cpp - FreeType v2 implementation

### Win32 TrueType Support

- [ ] agg_font_win32_tt.h - Win32 TrueType header
- [ ] agg_font_win32_tt.cpp - Win32 TrueType implementation

## General Polygon Clipper (GPC)

- [ ] gpc.h - GPC header
- [ ] gpc.c - GPC implementation

## Platform Support Implementations

### Cross-Platform

- [ ] agg_platform_support.cpp (generic interface)

### Platform-Specific (Optional - for examples)

- [ ] src/platform/X11/agg_platform_support.cpp - X11 support
- [ ] src/platform/win32/agg_platform_support.cpp - Win32 support
- [ ] src/platform/win32/agg_win32_bmp.cpp - Win32 bitmap support
- [ ] src/platform/mac/agg_platform_support.cpp - macOS support
- [ ] src/platform/mac/agg_mac_pmap.cpp - macOS pixmap support
- [ ] src/platform/BeOS/agg_platform_support.cpp - BeOS support
- [ ] src/platform/AmigaOS/agg_platform_support.cpp - AmigaOS support
- [ ] src/platform/sdl/agg_platform_support.cpp - SDL support
- [ ] src/platform/sdl2/agg_platform_support.cpp - SDL2 support
- [ ] src/platform/nano-X/agg_platform_support.cpp - nano-X support

## Priority Order

### Phase 1: Core Foundation

1. Basic types and configuration (agg_basics.h, agg_config.h, agg_array.h, agg_math.h)
2. Color handling (agg_color_rgba.h/.cpp, agg_color_gray.h)
3. Rendering buffer (agg_rendering_buffer.h)
4. Basic pixel formats (agg_pixfmt_base.h, agg_pixfmt_rgba.h, agg_pixfmt_rgb.h)

### Phase 2: Core Rendering

1. Scanlines (agg_scanline_u.h, agg_scanline_p.h, agg_scanline_bin.h)
2. Rasterizers (agg_rasterizer_scanline_aa.h, agg_rasterizer_cells_aa.h)
3. Basic renderers (agg_renderer_base.h, agg_renderer_scanline.h)
4. Spans (agg_span_solid.h, agg_span_allocator.h)

### Phase 3: Geometric Primitives

1. Path storage (agg_path_storage.h)
2. Basic shapes (agg_ellipse.h, agg_rounded_rect.h/.cpp, agg_arc.h/.cpp)
3. Transformations (agg_trans_affine.h/.cpp)
4. Basic converters (agg_conv_transform.h, agg_conv_stroke.h)

### Phase 4: Advanced Features

1. Curves (agg_curves.h/.cpp, agg_bezier_arc.h/.cpp, agg_bspline.h/.cpp)
2. Complex converters and vertex generators
3. Image processing and filters
4. Gradients and patterns
5. Text rendering
6. High-level AGG2D interface

## Notes

- Files marked with `.h` are header files that define interfaces and templates
- Files marked with `.cpp` are implementation files
- Some headers are template-only and may not have corresponding .cpp files
- Platform-specific files can be implemented as needed for target platforms
- The GPC library may need special licensing consideration
- Font support files are optional depending on text rendering requirements
