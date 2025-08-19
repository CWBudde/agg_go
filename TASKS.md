# AGG 2.6 Go Port - File Checklist

This is a comprehensive checklist of files that need to be ported from the original AGG 2.6 C++ codebase to Go. Please always check the original C++ implementation for reference in ../agg-2.6

## Core Header Files (include/)

Previously completed tasks have been moved to TASKS-COMPLETED.md

---

### Transformations

#### agg_trans_affine.h - 2D Affine transformations (core transformation system) ✅ **COMPLETED**

**Core Affine Matrix Structure**

- [x] trans_affine → TransAffine struct - 2x3 affine transformation matrix (`internal/transform/affine.go`)
  - [x] Matrix components: sx, shy, shx, sy, tx, ty (scaling, shearing, translation)
  - [x] Identity constructor → NewTransAffine() - creates identity matrix
  - [x] Custom matrix constructor with 6 parameters
  - [x] Copy constructor from existing matrix
  - [x] Array-based constructor from [6]float64

**Basic Transformation Methods**

- [x] Reset() - reset to identity matrix
- [x] Translation operations
  - [x] translate(x, y) → Translate(x, y) - add translation to current matrix
  - [x] tx, ty accessors → TX(), TY() - get translation components
- [x] Scaling operations
  - [x] scale(s) → Scale(s) - uniform scaling
  - [x] scale(sx, sy) → ScaleXY(sx, sy) - non-uniform scaling
  - [x] sx, sy accessors → SX(), SY() - get scaling components
- [x] Rotation operations
  - [x] rotate(angle) → Rotate(angle) - rotation in radians
  - [x] rotation() → Rotation() - extract rotation angle from matrix

**Advanced Matrix Operations**

- [x] Matrix composition and multiplication
  - [x] multiply(matrix) → Multiply(matrix) - matrix multiplication
  - [x] premultiply(matrix) → Premultiply(matrix) - pre-multiplication
  - [x] operator\*= overloading → Go method equivalent
- [x] Matrix inversion and determinant
  - [x] invert() → Invert() - matrix inversion with singularity check
  - [x] determinant() → Determinant() - calculate matrix determinant
  - [x] is_valid() → IsValid() - check for valid transformation
- [x] Decomposition methods
  - [x] scaling() → GetScaling() - extract combined scaling factor
  - [x] rotation() → GetRotation() - extract rotation angle
  - [x] translation() → GetTranslation() - extract translation vector

**Point and Vector Transformation**

- [x] Point transformation methods
  - [x] transform(x, y) → Transform(x, y) - transform point coordinates
  - [x] transform(point) → TransformPoint(point) - transform Point[T] type
  - [x] inverse_transform(x, y) → InverseTransform(x, y) - apply inverse transformation
- [x] Vector transformation (no translation)
  - [x] transform_2x2(x, y) → Transform2x2(x, y) - transform as vector
  - [x] rotation and scaling only, ignores translation component

**Specialized Transformation Constructors** (`internal/transform/helpers.go`)

- [x] trans_affine_rotation → NewTransAffineRotation(angle) - pure rotation matrix
- [x] trans_affine_scaling → NewTransAffineScaling(s) or (sx, sy) - pure scaling matrix
- [x] trans_affine_translation → NewTransAffineTranslation(x, y) - pure translation matrix
- [x] trans_affine_skewing → NewTransAffineSkewing(x, y) - pure skew matrix

**Matrix Analysis and Utilities**

- [x] Matrix property checks
  - [x] is_identity() → IsIdentity() - check for identity matrix
  - [x] is_equal(other) → IsEqual(other) - matrix equality with epsilon
  - [x] is_orthogonal() → IsOrthogonal() - check for orthogonal transformation
- [x] Epsilon handling
  - [x] affine_epsilon constant → AffineEpsilon - numerical precision threshold
  - [x] Custom epsilon support in comparison operations

**Integration with Path Processing**

- [x] Vertex source compatibility
  - [x] Compatible with conv_transform converter
  - [x] Path coordinate transformation
  - [x] Bounding box transformation
- [ ] Span interpolator integration (BLOCKED - span interpolators not yet implemented)
  - [ ] Matrix extraction for span_interpolator_linear
  - [ ] Coordinate system conversion for image transformations

**Advanced Composition Patterns**

- [x] Transformation around arbitrary points
  - [x] Pre-translate, transform, post-translate pattern
  - [x] Rotate around point helper functions
  - [x] Scale around point helper functions
- [x] Transformation chaining
  - [x] Method chaining for fluent API: NewTransAffine().Rotate(angle).Scale(s).Translate(x, y)
  - [x] Cumulative transformation building
  - [x] Order-dependent transformation sequences

**Performance Optimizations**

- [x] Fast path detection
  - [x] Identity matrix fast path
  - [x] Translation-only fast path
  - [x] Axis-aligned transformation detection
- [x] Batch transformation
  - [x] Transform multiple points efficiently
  - [ ] SIMD optimization opportunities
  - [x] Memory-friendly transformation patterns

**Error Handling and Edge Cases**

- [x] Singular matrix handling
  - [x] Determinant near-zero detection
  - [x] Graceful degradation for non-invertible matrices
  - [x] Error propagation in transformation chains
- [x] Numerical stability
  - [x] Large coordinate handling
  - [x] Precision preservation in repeated transformations
  - [x] Round-trip transformation accuracy

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] agg_math.h → internal/math package (trigonometric functions)
- [x] Point/Vector types from basic geometry package
- [x] Mathematical constants and utilities

#### agg_trans_bilinear.h - Bilinear transformations for quadrilateral mapping

**Bilinear Transformation Matrix**

- [x] trans_bilinear → TransBilinear struct - Bilinear coordinate transformation
  - [x] 4-point control system for arbitrary quadrilateral mapping
  - [x] Source and destination quadrilateral definitions
  - [x] Bilinear interpolation coefficients calculation
  - [x] Non-affine transformation capability (curved perspective effects)

**Quadrilateral Definition Methods**

- [x] Construction from control points
  - [x] rect_to_quad(x1, y1, x2, y2, quad) → RectToQuad() - rectangle to arbitrary quad
  - [x] quad_to_rect(quad, x1, y1, x2, y2) → QuadToRect() - quad to rectangle
  - [x] quad_to_quad(src_quad, dst_quad) → QuadToQuad() - arbitrary quad mapping
  - [x] Control point validation and normalization

**Bilinear Interpolation Mathematics**

- [x] Coefficient calculation
  - [x] A, B, C, D coefficient matrices for X and Y coordinates
  - [x] Bilinear basis function implementation
  - [x] Edge case handling for degenerate quadrilaterals
- [x] Forward transformation
  - [x] transform(x, y) → Transform(x, y) - apply bilinear mapping
  - [x] Parametric coordinate calculation (u, v parameters)
  - [x] Interpolation along quad edges and interior

**Inverse Transformation**

- [x] Inverse coordinate mapping
  - [x] inverse_transform(x, y) → InverseTransform(x, y) - reverse mapping
  - [x] Newton-Raphson iteration for inverse calculation
  - [x] Convergence criteria and iteration limits
  - [x] Fallback methods for difficult inversions

**Specialized Mapping Types**

- [x] Image rectification (application examples - core transformation supports these use cases)
  - [x] Perspective correction for photographed documents
  - [x] Keystone correction for projected displays
  - [x] Camera distortion correction integration
- [x] Texture mapping applications (application examples - core transformation supports these use cases)
  - [x] UV coordinate generation for arbitrary quads
  - [x] Mesh warping and distortion effects
  - [x] Real-time image transformation

**Integration with Rendering Pipeline** (BLOCKED - requires span interpolator infrastructure)

- [ ] Span interpolator compatibility (BLOCKED - span interpolators not yet implemented)
  - [ ] span_interpolator integration for image transformation
  - [ ] Bilinear filtering integration
  - [ ] Multi-pass rendering support
- [x] Path transformation (core transformation supports these use cases)
  - [x] Vector graphics warping
  - [x] Text distortion effects
  - [x] Shape morphing applications

**Performance and Optimization**

- [x] Coefficient caching
  - [x] Pre-calculated transformation matrices
  - [x] Incremental transformation updates (IteratorX)
  - [x] Batch transformation optimization
- [x] Numerical stability
  - [x] Singular quadrilateral detection
  - [x] Graceful degradation for edge cases
  - [x] Precision preservation in extreme transformations

**Dependencies**

- [x] agg_trans_affine.h → internal/transform package (affine transformation base)
- [x] agg_basics.h → internal/basics package (point types)
- [x] agg_simul_eq.h → simultaneous equation solver
- [x] Mathematical utilities for iterative solving

#### agg_trans_perspective.h - 3D perspective projections in 2D space ✅ **COMPLETED**

**Perspective Transformation Matrix**

- [x] trans_perspective → TransPerspective struct - 3x3 homogeneous transformation matrix
  - [x] Matrix components: sx, shy, w0, shx, sy, w1, tx, ty, w2
  - [x] Homogeneous coordinate system support
  - [x] Perspective division implementation (w-coordinate handling)
  - [x] Identity matrix initialization

**Construction Methods**

- [x] Basic constructors
  - [x] Identity constructor → NewTransPerspective() - identity transformation
  - [x] Custom matrix constructor with 9 parameters
  - [x] Array-based constructor from [9]float64
  - [x] Copy from affine matrix (w2=1, w0=w1=0)

**Quadrilateral-Based Perspective**

- [x] Quadrilateral mapping constructors
  - [x] rect_to_quad(x1, y1, x2, y2, quad) → RectToQuad() - rectangle to perspective quad
  - [x] quad_to_rect(quad, x1, y1, x2, y2) → QuadToRect() - reverse perspective mapping
  - [x] quad_to_quad(src, dst) → QuadToQuad() - arbitrary quad-to-quad perspective

**Perspective Transformation Methods**

- [x] Point transformation
  - [x] transform(x, y) → Transform(x, y) - apply perspective projection
  - [x] Homogeneous coordinate calculation
  - [x] Perspective division (x/w, y/w)
  - [x] Handle infinite/undefined points (w near zero)
- [x] Inverse transformation
  - [x] inverse_transform(x, y) → InverseTransform(x, y) - reverse perspective
  - [x] Matrix inversion for perspective matrices
  - [x] Singularity detection and handling

**Matrix Operations**

- [x] Matrix composition
  - [x] multiply(matrix) → Multiply(matrix) - perspective matrix multiplication
  - [x] premultiply(matrix) → Premultiply(matrix) - pre-multiplication
  - [x] Operator overloading equivalent methods
- [x] Matrix analysis
  - [x] determinant() → Determinant() - 3x3 matrix determinant
  - [x] invert() → Invert() - perspective matrix inversion
  - [x] is_valid() → IsValid() - matrix validity checking

**Advanced Perspective Operations** (Future enhancements)

- [ ] 3D projection simulation
  - [ ] Vanishing point calculations
  - [ ] Horizon line determination
  - [ ] Field of view and focal length simulation
- [ ] Perspective correction
  - [ ] Keystone correction for displays
  - [ ] Camera perspective correction
  - [ ] Architectural photography correction

**Integration with Affine Transformations**

- [x] Affine compatibility
  - [x] Conversion from affine to perspective
  - [x] Affine transformation embedding
  - [x] Mixed transformation pipelines
- [x] Transformation composition
  - [x] Affine pre/post-multiplication
  - [x] Complex transformation chains
  - [x] Order-dependent composition

**Specialized Applications** (Application examples - core transformation supports these use cases)

- [x] Image rectification (applications enabled by core implementation)
  - [x] Perspective distortion correction
  - [x] Document scanning applications
  - [x] Real-time perspective correction
- [x] 3D graphics simulation (applications enabled by core implementation)
  - [x] 2D sprite projection
  - [x] Pseudo-3D effects
  - [x] Depth simulation in 2D

**Performance Considerations**

- [ ] Division optimization
  - [ ] Fast perspective division techniques
  - [ ] Avoiding division by near-zero w values
  - [ ] Batch transformation optimization
- [ ] Numerical stability
  - [ ] Precision handling in extreme perspectives
  - [ ] Graceful degradation for singular matrices
  - [ ] Error propagation management

**Dependencies**

- [x] agg_trans_affine.h → internal/transform package (affine transformation base)
- [x] agg_simul_eq.h → simultaneous equation solver (internal/transform/simul_eq.go)
- [x] agg_basics.h → internal/basics package
- [x] Mathematical utilities for 3x3 matrix operations

#### agg_trans_viewport.h - Viewport and coordinate system transformations ✅ **COMPLETED**

**Viewport Transformation System**

- [x] trans_viewport → TransViewport struct - Coordinate system mapping
  - [x] World-to-device coordinate transformation
  - [x] Aspect ratio preservation options
  - [x] Viewport clipping rectangle definition
  - [x] Automatic scaling and centering

**Viewport Definition Methods**

- [x] World coordinate system
  - [x] world_viewport(x1, y1, x2, y2) → WorldViewport() - define world bounds
  - [x] World coordinate range specification
  - [x] Infinite coordinate space handling
  - [x] Coordinate system orientation (Y-up vs Y-down)
- [x] Device coordinate system
  - [x] device_viewport(x1, y1, x2, y2) → DeviceViewport() - define device bounds
  - [x] Pixel-perfect device coordinate mapping
  - [x] Device coordinate constraints and validation
  - [x] Screen/window coordinate integration

**Aspect Ratio and Scaling Control**

- [x] Aspect ratio preservation
  - [x] aspect_ratio_e enumeration → AspectRatio type
  - [x] stretch → AspectRatioStretch - fill viewport completely (distortion allowed)
  - [x] meet → AspectRatioMeet - fit entirely within viewport (letterbox/pillarbox)
  - [x] slice → AspectRatioSlice - fill viewport, crop excess (zoom to fit)
- [x] Alignment control
  - [x] Horizontal alignment: left, center, right via alignX parameter
  - [x] Vertical alignment: top, middle, bottom via alignY parameter
  - [x] Custom alignment offset parameters (0.0-1.0 range)

**Transformation Calculation**

- [x] Automatic matrix generation
  - [x] Scale factor calculation based on viewport ratios
  - [x] Translation calculation for centering and alignment
  - [x] Combined scaling and translation matrix
  - [x] Matrix update on viewport changes
- [x] Transformation extraction
  - [x] to_affine() → ToAffine() - extract equivalent affine transformation
  - [x] scale_x(), scale_y() → ScaleX(), ScaleY() - get scaling factors
  - [x] offset_x(), offset_y() → DeviceDX(), DeviceDY() - get translation offsets

**Coordinate Conversion Methods**

- [x] World-to-device conversion
  - [x] world_to_device(x, y) → Transform(x, y) - forward transformation
  - [x] Bulk point transformation for performance
  - [x] Path coordinate transformation
  - [x] Bounding box transformation
- [x] Device-to-world conversion
  - [x] device_to_world(x, y) → InverseTransform(x, y) - inverse transformation
  - [x] Mouse coordinate mapping
  - [x] Hit testing coordinate conversion
  - [x] Interactive viewport navigation

**Viewport State Management**

- [x] Viewport validation
  - [x] is_valid() → IsValid() - check for valid viewport configuration
  - [x] Zero-size viewport handling
  - [x] Invalid coordinate range detection
  - [x] Degenerate transformation prevention
- [x] State change detection
  - [x] viewport_changed() → update() - detect viewport modifications (automatic)
  - [x] Automatic matrix recalculation
  - [x] Change event propagation (via update() method)

**Advanced Viewport Features**

- [ ] Multi-viewport support
  - [ ] Nested viewport transformations
  - [ ] Viewport hierarchy management
  - [ ] Relative viewport positioning
- [ ] Zoom and pan integration
  - [ ] Interactive zoom controls
  - [ ] Pan operation coordinate handling
  - [ ] Zoom-to-fit functionality
  - [ ] Smooth transition support

**Integration with Rendering System**

- [ ] Renderer integration
  - [ ] Automatic clipping rectangle setup
  - [ ] Coordinate system integration
  - [ ] Multi-pass rendering support
- [ ] Path processing integration
  - [ ] Automatic path transformation
  - [ ] Viewport-aware path optimization
  - [ ] Level-of-detail based on viewport scale

**Performance Optimizations**

- [ ] Cached transformation matrix
  - [ ] Matrix recalculation only when needed
  - [ ] Incremental viewport updates
  - [ ] Fast path for common viewport changes
- [ ] Batch transformations
  - [ ] Efficient multi-point transformation
  - [ ] SIMD optimization opportunities
  - [ ] Memory-friendly transformation patterns

**Dependencies**

- [x] agg_trans_affine.h → internal/transform package
- [x] agg_basics.h → internal/basics package
- [x] Mathematical utilities for viewport calculations
- [x] Enumeration types for aspect ratio and alignment

#### agg_trans_single_path.h - Transform along a single curved path ✅ **COMPLETED**

**Path-Based Transformation System**

- [x] trans_single_path → TransSinglePath struct - Transform coordinates along curved path
  - [x] Path-following coordinate system
  - [x] Arc-length parameterization
  - [x] Normal and tangent vector calculation
  - [x] Distance-based positioning along path

**Path Definition and Processing**

- [x] Path setup
  - [x] add_path(vertex_source) → AddPath() - define the transformation path
  - [x] Path length calculation and caching
  - [ ] Path segment analysis and optimization (future enhancement)
  - [ ] Closed vs. open path handling (future enhancement)
- [x] Path analysis
  - [x] total_length() → TotalLength() - get total path length
  - [ ] Curvature analysis for quality control (future enhancement)
  - [ ] Path direction and orientation (future enhancement)
  - [ ] Critical point identification (cusps, loops) (future enhancement)

**Coordinate Transformation Methods**

- [x] Forward transformation
  - [x] transform(x, y) → Transform(x, y) - map coordinates to path
  - [x] X-coordinate maps to distance along path
  - [x] Y-coordinate maps to perpendicular offset from path
  - [x] Tangent and normal vector calculation at each point
- [ ] Path parameterization
  - [ ] Arc-length parameterization for even spacing
  - [ ] Parameter-to-distance conversion
  - [ ] Distance-to-parameter conversion
  - [ ] Interpolation between path segments

**Path Navigation and Positioning**

- [ ] Distance-based positioning
  - [ ] Position calculation at specific path distance
  - [ ] Interpolation between path vertices
  - [ ] Smooth position transitions
  - [ ] Boundary condition handling (before start, after end)
- [ ] Orientation calculation
  - [ ] Tangent vector at any path position
  - [ ] Normal vector for perpendicular positioning
  - [ ] Angle calculation for oriented objects
  - [ ] Smooth orientation transitions

**Advanced Path Features**

- [ ] Path quality control
  - [ ] approximation_scale() → SetApproximationScale() - control path detail
  - [ ] Adaptive path tessellation
  - [ ] Smooth curve approximation
  - [ ] Sharp corner handling
- [ ] Path modification
  - [ ] Path reversal for direction changes
  - [ ] Path segment extraction
  - [ ] Path smoothing and filtering
  - [ ] Multi-path composition

**Text-on-Path Applications** (Application examples - core transformation supports these use cases)

- [x] Text layout along curves (applications enabled by core implementation)
  - [x] Character positioning and orientation
  - [x] Baseline following path curvature
  - [x] Character spacing adjustment for curves
  - [x] Text direction and reading flow
- [x] Advanced text features (applications enabled by core implementation)
  - [x] Multi-line text on path
  - [x] Text alignment options (left, center, right, justify)
  - [x] Overflow handling for text longer than path
  - [x] Dynamic text fitting and scaling

**Shape Transformation Applications** (Application examples - core transformation supports these use cases)

- [x] Shape morphing along paths (applications enabled by core implementation)
  - [x] Object orientation following path direction
  - [x] Scale variation along path
  - [x] Shape deformation based on path curvature
- [ ] Animation support
  - [ ] Smooth position interpolation
  - [ ] Velocity-based positioning
  - [ ] Acceleration and deceleration curves
  - [ ] Keyframe animation along paths

**Performance Optimizations**

- [ ] Path preprocessing
  - [ ] Arc-length table generation
  - [ ] Segment caching for repeated access
  - [ ] Fast lookup tables for distance mapping
  - [ ] Incremental path processing
- [ ] Transformation caching
  - [ ] Cached tangent and normal vectors
  - [ ] Reuse calculations for nearby points
  - [ ] Batch transformation optimization

**Integration with Path Storage**

- [ ] Vertex source compatibility
  - [ ] Compatible with path_storage
  - [ ] Works with all curve types
  - [ ] Automatic curve approximation
- [ ] Path converter integration
  - [ ] Works with conv_stroke for outlined paths
  - [ ] Compatible with conv_dash for dashed paths
  - [ ] Integration with path transformation pipeline

**Dependencies**

- [x] agg_path_storage.h → internal/path package
- [x] agg_curves.h → curve approximation classes (internal/curves - partially implemented)
- [x] agg_vertex_sequence.h → vertex processing utilities (internal/vcgen)
- [x] Arc-length calculation utilities
- [x] Vector mathematics for tangent/normal calculations

#### agg_trans_double_path.h - Transform between two curved paths ✅ **COMPLETED**

**Dual-Path Transformation System**

- [x] trans_double_path → TransDoublePath struct - Transform using two guide paths
  - [x] Base path and top path definition
  - [x] Morphing between two arbitrary curved paths
  - [x] Bilinear interpolation across path pair
  - [x] Variable width corridor transformation

**Path Pair Definition**

- [ ] Dual path setup
  - [ ] add_paths(base_path, top_path) → AddPaths() - define transformation corridor
  - [ ] Path synchronization and length matching
  - [ ] Corresponding point calculation
  - [ ] Path direction alignment
- [ ] Path relationship analysis
  - [ ] Relative path positioning
  - [ ] Path separation distance calculation
  - [ ] Crossing and intersection detection
  - [ ] Path convergence and divergence points

**Coordinate Transformation Mathematics**

- [ ] Bilinear path interpolation
  - [ ] X-coordinate maps to position along both paths
  - [ ] Y-coordinate maps to interpolation between paths (0=base, 1=top)
  - [ ] Smooth interpolation across the path corridor
  - [ ] Edge case handling for path intersections
- [ ] Transform calculation
  - [ ] transform(x, y) → Transform(x, y) - map point to path corridor
  - [ ] Corresponding point calculation on both paths
  - [ ] Linear interpolation between corresponding points
  - [ ] Orientation interpolation for smooth transitions

**Advanced Path Processing**

- [ ] Path synchronization
  - [ ] Arc-length parameterization for both paths
  - [ ] Path re-parameterization for consistent mapping
  - [ ] Automatic path length adjustment
  - [ ] Critical point alignment between paths
- [ ] Quality control
  - [ ] approximation_scale() → SetApproximationScale() - control detail level
  - [ ] Path smoothing and filtering
  - [ ] Self-intersection detection and handling
  - [ ] Degenerate case management

**Corridor-Based Applications** (Application examples - core transformation supports these use cases)

- [x] Text layout in variable-width corridors (applications enabled by core implementation)
  - [x] Text flowing between curved boundaries
  - [x] Dynamic text sizing based on corridor width
  - [x] Multi-line text in curved regions
  - [x] Text justification in non-uniform spaces
- [x] Shape morphing between paths (applications enabled by core implementation)
  - [x] Smooth shape transitions
  - [x] Animation between different path shapes
  - [x] Envelope distortion effects
  - [x] Perspective-like distortions

**Path Corridor Analysis**

- [ ] Width calculation
  - [ ] Distance between paths at any position
  - [ ] Minimum and maximum corridor width
  - [ ] Width variation analysis
  - [ ] Bottleneck detection
- [ ] Path relationship metrics
  - [ ] Average separation distance
  - [ ] Path parallelism measurement
  - [ ] Correlation between path curvatures
  - [ ] Geometric compatibility assessment

**Specialized Transformation Effects**

- [ ] Envelope distortion
  - [ ] Shape fitting within curved boundaries
  - [ ] Proportional scaling based on corridor width
  - [ ] Perspective-like effects from path convergence
- [ ] Flow field simulation
  - [ ] Particle movement along path corridor
  - [ ] Velocity field generation
  - [ ] Streamline visualization
  - [ ] Fluid dynamics simulation

**Performance and Optimization**

- [ ] Dual path preprocessing
  - [ ] Arc-length table generation for both paths
  - [ ] Corresponding point cache
  - [ ] Fast lookup structures
  - [ ] Incremental processing optimization
- [ ] Batch transformation
  - [ ] Efficient multi-point processing
  - [ ] Vectorized calculations
  - [ ] Memory-friendly access patterns

**Error Handling and Edge Cases**

- [ ] Path mismatch handling
  - [ ] Different path lengths
  - [ ] Path intersection scenarios
  - [ ] Degenerate path configurations
  - [ ] Numerical stability in extreme cases
- [ ] Boundary condition management
  - [ ] Beyond path endpoints
  - [ ] Outside path corridor
  - [ ] Path crossing points
  - [ ] Singular transformation points

**Dependencies**

- [x] agg_trans_single_path.h → single path transformation (base functionality)
- [x] agg_path_storage.h → internal/path package
- [x] agg_curves.h → curve approximation
- [x] Vector mathematics and interpolation utilities
- [x] Arc-length parameterization tools

#### agg_trans_warp_magnifier.h - Warp magnifier transformation (lens effects) ✅ **COMPLETED**

**Magnifier Lens Transformation**

- [x] trans_warp_magnifier → TransWarpMagnifier struct - Lens distortion transformation
  - [x] Circular magnification area definition
  - [x] Variable magnification factor
  - [x] Smooth distortion falloff
  - [x] Real-time lens effect simulation

**Lens Parameter Definition**

- [x] Magnifier setup
  - [x] center(x, y) → SetCenter(x, y) - lens center position
  - [x] radius(r) → SetRadius(r) - magnification area radius
  - [x] magnification(m) → SetMagnification(m) - magnification factor
  - [x] Interactive parameter adjustment
- [ ] Lens shape control
  - [ ] Circular lens (default)
  - [ ] Elliptical lens variations
  - [ ] Custom lens shape support
  - [ ] Lens boundary smoothness control

**Magnification Mathematics**

- [x] Lens distortion calculation
  - [x] Distance-from-center calculation
  - [x] Radial magnification formula
  - [x] Smooth falloff function (avoid sharp edges)
  - [x] Inverse transformation for mouse interaction
- [x] Forward transformation
  - [x] transform(x, y) → Transform(x, y) - apply lens distortion
  - [x] Magnified region calculation
  - [x] Normal region pass-through
  - [x] Smooth transition between regions

**Advanced Lens Effects**

- [ ] Multiple magnification zones
  - [ ] Overlapping lens effects
  - [ ] Additive vs. multiplicative magnification
  - [ ] Complex distortion patterns
  - [ ] Multi-center lens systems
- [ ] Dynamic lens properties
  - [ ] Animated magnification factor
  - [ ] Moving lens center
  - [ ] Pulsing or breathing effects
  - [ ] Interactive real-time control

**Distortion Quality Control**

- [ ] Anti-aliasing integration
  - [ ] Smooth distortion boundaries
  - [ ] Quality preservation in magnified areas
  - [ ] Artifact minimization
  - [ ] Sampling rate adjustment based on magnification
- [ ] Edge handling
  - [ ] Smooth falloff to normal transformation
  - [ ] Boundary artifact prevention
  - [ ] Edge case numerical stability
  - [ ] Clipping region integration

**Interactive Applications** (Application examples - core transformation supports these use cases)

- [x] Real-time magnification (applications enabled by core implementation)
  - [x] Mouse-driven lens positioning
  - [x] Zoom level control
  - [x] Smooth lens movement
  - [x] Performance optimization for interaction
- [x] Document and image viewing (applications enabled by core implementation)
  - [x] Detail inspection tools
  - [x] Accessibility magnification
  - [x] Scientific image analysis
  - [x] CAD drawing detail viewing

**Rendering Integration**

- [ ] Renderer compatibility
  - [ ] Works with all pixel formats
  - [ ] Scanline renderer integration
  - [ ] Span generator compatibility
  - [ ] Multi-pass rendering support
- [ ] Performance optimization
  - [ ] Fast path for non-magnified regions
  - [ ] Incremental updates for moving lens
  - [ ] Region-of-interest optimization
  - [ ] Memory usage minimization

**Lens Physics Simulation**

- [ ] Realistic lens effects
  - [ ] Optical magnification simulation
  - [ ] Barrel and pincushion distortion
  - [ ] Chromatic aberration effects
  - [ ] Lens flare and reflection simulation
- [ ] Customizable distortion profiles
  - [ ] User-defined magnification curves
  - [ ] Asymmetric distortion patterns
  - [ ] Non-uniform magnification fields
  - [ ] Complex lens shape simulation

**Performance Considerations**

- [ ] Transformation caching
  - [ ] Pre-calculated distortion maps
  - [ ] Fast lookup tables
  - [ ] Incremental updates
  - [ ] Region-based optimization
- [ ] Real-time performance
  - [ ] Frame rate maintenance
  - [ ] Smooth animation support
  - [ ] Memory-efficient processing
  - [ ] GPU acceleration compatibility

**Mathematical Foundation**

- [ ] Lens distortion mathematics
  - [ ] Radial distortion models
  - [ ] Polynomial approximation
  - [ ] Inverse transformation calculation
  - [ ] Numerical stability considerations
- [ ] Coordinate system handling
  - [ ] Screen coordinate integration
  - [ ] World coordinate transformation
  - [ ] Multi-scale coordinate systems
  - [ ] Precision preservation

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] agg_math.h → mathematical utilities
- [x] Distance calculation utilities
- [x] Smooth interpolation functions
- [x] Real-time performance optimization tools

---

### Converters

#### agg_conv_adaptor_vcgen.h - Vertex Generator Adaptor (base infrastructure)

**Core Adaptor System**

- [x] conv_adaptor_vcgen → ConvAdaptorVCGen[VS, G, M] struct - Base template for vertex generator adaptors
  - [x] VertexSource template parameter → Generic over any vertex source
  - [x] Generator template parameter → Wraps vertex generators (stroke, dash, contour, etc.)
  - [x] Markers template parameter → Optional marker handling (default: null_markers)
  - [x] State machine for vertex source processing

**Null Markers Implementation**

- [x] null_markers → NullMarkers struct - Default empty marker implementation
  - [x] remove_all() → RemoveAll() - no-op marker removal
  - [x] add_vertex(x, y, cmd) → AddVertex() - no-op marker addition
  - [x] prepare_src() → PrepareSrc() - no-op source preparation
  - [x] Vertex source interface (rewind/vertex) with path_cmd_stop

**Adaptor State Management**

- [x] Processing states enumeration → AdaptorStatus type
  - [x] initial → Initial - starting state, no processing begun
  - [x] accumulate → Accumulate - collecting vertices from source
  - [x] generate → Generate - producing output from generator
- [x] State machine implementation
  - [x] attach(source) → Attach() - connect to new vertex source
  - [x] rewind(path_id) → Rewind() - reset processing state
  - [x] vertex() → Vertex() - state-driven vertex production

**Generator and Marker Access**

- [x] Component access methods
  - [x] generator() → Generator() - access underlying vertex generator
  - [x] const generator() → GetGenerator() - read-only generator access
  - [x] markers() → Markers() - access marker processor
  - [x] const markers() → GetMarkers() - read-only marker access

**Vertex Processing Pipeline**

- [x] Three-phase processing
  - [x] Source accumulation: collect all vertices from input
  - [x] Generator processing: apply transformation/generation
  - [x] Output generation: emit processed vertices
- [x] Path command handling
  - [x] Preserve path structure and commands
  - [x] Handle multi-path vertex sources
  - [x] Path ID propagation through pipeline

**Template Specialization Support**

- [ ] Generic programming patterns
  - [ ] Compatible with all vertex source types
  - [ ] Works with any conforming generator
  - [ ] Marker system extensibility
  - [ ] Type-safe template composition

**Integration with Vertex Generators**

- [ ] Generator lifecycle management
  - [ ] Generator initialization and configuration
  - [ ] State synchronization between adaptor and generator
  - [ ] Error propagation from generator to adaptor
- [ ] Memory management
  - [ ] Efficient vertex buffering
  - [ ] Memory reuse across processing cycles
  - [ ] Minimal allocation during processing

**Performance Optimizations**

- [ ] State machine efficiency
  - [ ] Fast state transitions
  - [ ] Minimal overhead in generate state
  - [ ] Batch processing optimization
- [ ] Template instantiation
  - [ ] Compile-time optimization opportunities
  - [ ] Type erasure where beneficial
  - [ ] Inlining critical path operations

**Dependencies**

- agg_basics.h → internal/basics package
- Vertex source interface definition
- Path command constants
- Template constraint definitions

#### agg_conv_adaptor_vpgen.h - Vertex Processor Adaptor (base infrastructure) ✓

**Vertex Processor Adaptor System**

- [x] conv_adaptor_vpgen → ConvAdaptorVPGen[VPG] struct - Base template for vertex processor adaptors
  - [x] VertexSource parameter → Input vertex source type
  - [x] VPGen template parameter → Vertex processing algorithm (clipping, segmentation)
  - [x] Real-time vertex processing (no accumulation phase)
  - [x] AutoClose/AutoUnclose support for polygon handling

**Streaming Vertex Processing**

- [x] Direct vertex transformation
  - [x] No intermediate vertex storage required
  - [x] Full state management for complex polygon processing
  - [ ] Real-time processing as vertices are requested
  - [ ] Memory efficient for large paths
  - [ ] Suitable for clipping and segmentation operations

**Processor Interface Requirements**

- [ ] Processor conformance interface
  - [ ] Processor must support streaming operation
  - [ ] State management within processor
  - [ ] Path command preservation
  - [ ] Multi-path handling capability

**State Management (Simplified)**

- [ ] Simpler state model than vcgen adaptor
  - [ ] Direct pass-through to processor
  - [ ] Minimal state tracking overhead
  - [ ] Immediate vertex processing
  - [ ] Real-time error handling

**Integration Points**

- [ ] Compatible processors
  - [ ] Clipping processors (polygon, polyline)
  - [ ] Segmentation processors
  - [ ] Path modification processors
  - [ ] Custom streaming processors

**Performance Characteristics**

- [ ] Low memory overhead
  - [ ] No vertex accumulation required
  - [ ] Constant memory usage
  - [ ] Suitable for very large paths
- [ ] Real-time processing
  - [ ] Immediate vertex output
  - [ ] No processing delays
  - [ ] Interactive performance

**Dependencies**

- agg_basics.h → internal/basics package
- Vertex processor interface definitions
- Streaming processing utilities

#### [x] agg_conv_stroke.h - Stroke Converter (path outline generation) ✅ **COMPLETED**

**Stroke Converter System**

- [x] conv_stroke → ConvStroke[VS, M] struct - Convert paths to stroked outlines (`internal/conv/conv_stroke.go`)
  - [x] VertexSource template parameter → Input path to be stroked
  - [x] Markers template parameter → Optional stroke markers
  - [x] Inherits from conv_adaptor_vcgen with vcgen_stroke generator
  - [x] Comprehensive stroke parameter control

**Line Cap Style Configuration**

- [x] Line cap enumeration → LineCap type (`internal/basics/types.go`)
  - [x] butt_cap → ButtCap - flat end perpendicular to path
  - [x] square_cap → SquareCap - square extension beyond path end
  - [x] round_cap → RoundCap - circular end centered on path end
- [x] Cap style methods
  - [x] line_cap(cap_style) → SetLineCap() - set end cap style
  - [x] line_cap() → LineCap() - get current cap style

**Line Join Style Configuration**

- [x] Line join enumeration → LineJoin type (`internal/basics/types.go`)
  - [x] miter_join → MiterJoin - sharp corner with miter limit
  - [x] miter_join_revert → MiterJoinRevert - fallback to bevel when limit exceeded
  - [x] round_join → RoundJoin - circular arc at corners
  - [x] bevel_join → BevelJoin - flat cut across corner
- [x] Join style methods
  - [x] line_join(join_style) → SetLineJoin() - set corner join style
  - [x] line_join() → LineJoin() - get current join style

**Inner Join Handling**

- [ ] Inner join enumeration → InnerJoin type
  - [ ] inner_bevel → InnerBevel - beveled inner joins
  - [ ] inner_miter → InnerMiter - mitered inner joins
  - [ ] inner_jag → InnerJag - jagged inner joins (fastest)
  - [ ] inner_round → InnerRound - rounded inner joins (smoothest)
- [ ] Inner join methods
  - [ ] inner_join(inner_style) → SetInnerJoin() - set inner corner style
  - [ ] inner_join() → GetInnerJoin() - get current inner style

**Stroke Width and Measurement**

- [ ] Width control
  - [ ] width(w) → SetWidth() - set stroke line width
  - [ ] width() → GetWidth() - get current stroke width
  - [ ] Width measurement in user coordinate units
  - [ ] Zero-width stroke handling (hairline rendering)

**Miter Limit Control**

- [ ] Miter limiting to prevent excessive spikes
  - [ ] miter_limit(limit) → SetMiterLimit() - set miter limit ratio
  - [ ] miter_limit() → GetMiterLimit() - get current miter limit
  - [ ] miter_limit_theta(angle) → SetMiterLimitTheta() - set limit by angle
  - [ ] inner_miter_limit(limit) → SetInnerMiterLimit() - inner corner limit

**Path Shortening**

- [ ] Path end modification
  - [ ] shorten(amount) → SetShorten() - shorten path ends by specified amount
  - [ ] shorten() → GetShorten() - get current shortening amount
  - [ ] Symmetric shortening of both path ends
  - [ ] Useful for arrow and marker integration

**Quality and Approximation Control**

- [ ] Rendering quality parameters
  - [ ] approximation_scale(scale) → SetApproximationScale() - control curve detail
  - [ ] approximation_scale() → GetApproximationScale() - get current scale
  - [ ] Affects circular arc approximation in round caps/joins
  - [ ] Balance between quality and performance

**Advanced Stroke Features**

- [ ] Complex stroke scenarios
  - [ ] Self-intersecting path handling
  - [ ] Zero-length segment handling
  - [ ] Degenerate path processing
  - [ ] Path direction preservation

**Generator Integration**

- [ ] Underlying vcgen_stroke access
  - [ ] All methods delegate to base vcgen_stroke generator
  - [ ] Type-safe parameter forwarding
  - [ ] State management through adaptor base class
  - [ ] Marker integration support

**Performance Considerations**

- [ ] Stroke generation efficiency
  - [ ] Optimized for common stroke parameters
  - [ ] Caching opportunities for static strokes
  - [ ] Memory efficient vertex generation
  - [ ] Fast path for simple stroke cases

**Integration with Path Pipeline**

- [ ] Composability with other converters
  - [ ] Can be chained with other path modifications
  - [ ] Compatible with dash, marker, and transform converters
  - [ ] Works with all vertex source types
  - [ ] Output compatible with all rendering systems

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_stroke.h → stroke vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Line style enumeration definitions

#### agg_conv_dash.h - Dash Converter (dashed line patterns) ✅ **COMPLETED**

**Dash Pattern System**

- [x] conv_dash → ConvDash struct - Add dash patterns to paths (`internal/conv/conv_dash.go`)
  - [x] VertexSource template parameter → Input path to be dashed
  - [x] Markers template parameter → Optional dash markers
  - [x] Inherits from conv_adaptor_vcgen with vcgen_dash generator
  - [x] Flexible dash pattern definition

**Dash Pattern Management**

- [x] Pattern definition methods
  - [x] remove_all_dashes() → RemoveAllDashes() - clear all dash patterns
  - [x] add_dash(dash_len, gap_len) → AddDash() - add dash/gap pair to pattern
  - [x] Multiple dash patterns create complex repeating sequences
  - [x] Pattern length calculated as sum of all dash/gap pairs

**Dash Pattern Phase Control**

- [x] Pattern positioning
  - [x] dash_start(offset) → DashStart() - set starting offset in pattern
  - [x] Phase control for pattern alignment
  - [x] Offset wraps around pattern length
  - [x] Useful for animation and pattern synchronization

**Path Length and Shortening**

- [x] Path modification
  - [x] shorten(amount) → Shorten() - shorten path before dashing
  - [x] shorten() → GetShorten() - get current shortening amount
  - [x] Applied before dash pattern calculation
  - [x] Useful for precise pattern termination

**Dash Generation Algorithm**

- [x] Pattern application process
  - [x] Path length calculation and parameterization
  - [x] Pattern repetition across path length
  - [x] Dash segment extraction from continuous path
  - [x] Gap handling (no vertex output during gaps)

**Complex Path Handling**

- [x] Multi-segment path support
  - [x] Pattern continues across path segments
  - [x] Pattern phase preservation at path connections
  - [x] Proper handling of path commands (move_to, line_to, curves)
  - [x] Closed path pattern continuity

**Integration with Stroke Converter**

- [ ] Common usage pattern: dash then stroke
  - [ ] conv_stroke<conv_dash<path_storage>> composition
  - [ ] Dash patterns applied before stroke generation
  - [ ] Stroke parameters applied to individual dash segments
  - [ ] Line caps applied to each dash segment

**Performance Optimization**

- [ ] Efficient dash generation
  - [ ] Arc-length parameterization for accurate spacing
  - [ ] Minimal vertex generation during gaps
  - [ ] Pattern calculation optimization
  - [ ] Incremental pattern processing

**Animation Support**

- [ ] Dynamic dash patterns
  - [ ] Animated dash_start offset for moving patterns
  - [ ] Real-time pattern modification
  - [ ] Smooth dash animation effects
  - [ ] Interactive dash parameter adjustment

**Edge Case Handling**

- [ ] Robust pattern processing
  - [ ] Zero-length dash or gap handling
  - [ ] Very short path segments
  - [ ] Pattern longer than path handling
  - [ ] Degenerate path processing

**Generator Integration**

- [ ] Underlying vcgen_dash access
  - [ ] Parameter delegation to base generator
  - [ ] State management through adaptor framework
  - [ ] Memory efficient dash segment storage
  - [ ] Pattern calculation caching

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_dash.h → dash vertex generator ✅
- agg_conv_adaptor_vcgen.h → base adaptor system
- Arc-length calculation utilities

#### agg_conv_contour.h - Contour Converter (path offset generation) ✅

**Contour Generation System**

- [x] conv_contour → ConvContour[VS] struct - Generate offset contours from paths
  - [x] VertexSource template parameter → Input path for contour generation
  - [x] Inherits from conv_adaptor_vcgen with vcgen_contour generator
  - [x] Parallel curve generation with configurable offset distance
  - [x] Support for both expansion and contraction

**Contour Offset Control**

- [x] Distance configuration
  - [x] width(distance) → SetWidth() - set contour offset distance
  - [x] width() → GetWidth() - get current offset distance
  - [x] Positive values expand path outward
  - [x] Negative values contract path inward
  - [x] Zero width returns original path

**Line Join Handling for Contours**

- [x] Corner processing
  - [x] line_join(join_style) → SetLineJoin() - set contour join style
  - [x] line_join() → GetLineJoin() - get current join style
  - [x] Similar to stroke joins but for offset curves
  - [x] Critical for smooth contour appearance

**Inner Join Processing**

- [x] Inner corner handling
  - [x] inner_join(inner_style) → SetInnerJoin() - set inner join style
  - [x] inner_join() → GetInnerJoin() - get current inner style
  - [x] Important for path contraction (negative offsets)
  - [x] Prevents self-intersection in concave regions

**Miter Limit for Contours**

- [x] Sharp corner control
  - [x] miter_limit(limit) → SetMiterLimit() - set miter limit for contour joins
  - [x] miter_limit() → GetMiterLimit() - get current miter limit
  - [x] Prevents excessive spikes in sharp corners
  - [x] Automatic fallback to bevel when limit exceeded

**Approximation Quality**

- [x] Curve quality control
  - [x] approximation_scale(scale) → SetApproximationScale() - control curve detail
  - [x] approximation_scale() → GetApproximationScale() - get current scale
  - [x] Affects circular arc approximation in rounded joins
  - [x] Balance between smoothness and performance

**Contour Applications**

- [ ] Text outline generation
  - [ ] Create outlined text effects
  - [ ] Multi-level text outlines (nested contours)
  - [ ] Bold text simulation through expansion
- [ ] Shape morphing effects
  - [ ] Smooth shape expansion/contraction animations
  - [ ] Organic growth/shrink effects
  - [ ] Boundary visualization and emphasis

**Advanced Contour Processing**

- [ ] Complex path handling
  - [ ] Self-intersecting contour resolution
  - [ ] Island detection and processing
  - [ ] Hole preservation in contracted paths
  - [ ] Multiple contour level generation

**Integration with Other Converters**

- [ ] Converter chaining
  - [ ] Can be combined with stroke for outlined contours
  - [ ] Works with dash patterns for dashed outlines
  - [ ] Compatible with curve and smoothing converters
  - [ ] Transform integration for scaled contours

**Performance Considerations**

- [ ] Efficient offset calculation
  - [ ] Geometric algorithms for parallel curves
  - [ ] Memory efficient vertex generation
  - [ ] Caching for repeated contour operations
  - [ ] Optimization for simple geometric shapes

**Error Handling**

- [ ] Robust contour generation
  - [ ] Degenerate path handling
  - [ ] Very small offset distances
  - [ ] Self-intersecting input paths
  - [ ] Numerical stability in edge cases

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_contour.h → contour vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Geometric calculation utilities for parallel curves

#### agg_conv_curve.h - Curve Converter (curve-to-line approximation)

**Curve Approximation System**

- [x] conv_curve → ConvCurve[VS] struct - Convert path curves to line segments
  - [x] VertexSource template parameter → Input path with curves
  - [x] Direct curve processing without vertex generator adaptor
  - [x] Automatic curve detection and approximation
  - [x] Configurable approximation quality

**Curve Detection and Processing**

- [x] Curve command handling
  - [x] curve3 command → Quadratic Bezier curve approximation
  - [x] curve4 command → Cubic Bezier curve approximation
  - [x] Automatic curve type detection
  - [x] Preserves non-curve path commands (move_to, line_to, etc.)

**Approximation Quality Control**

- [x] Quality parameters
  - [x] approximation_method(method) → SetApproximationMethod() - subdivision method
  - [x] approximation_scale(scale) → SetApproximationScale() - detail level
  - [x] angle_tolerance(angle) → SetAngleTolerance() - curvature sensitivity
  - [x] cusp_limit(limit) → SetCuspLimit() - sharp corner threshold

**Approximation Methods**

- [ ] Curve subdivision techniques
  - [ ] curve_inc → CurveInc - incremental method (faster)
  - [ ] curve_div → CurveDiv - recursive subdivision (higher quality)
  - [ ] Automatic method selection based on curve properties
  - [ ] Method switching for optimal performance/quality trade-off

**Curve-Specific Parameters**

- [ ] Curve type handling
  - [ ] Different parameter sets for quadratic vs cubic curves
  - [ ] Specialized processing for each curve type
  - [ ] Adaptive parameter adjustment based on curve complexity
  - [ ] Consistent approximation quality across curve types

**Integration with Path Pipeline**

- [ ] Pipeline position
  - [ ] Typically applied early in conversion chain
  - [ ] Converts complex curves to simple line segments
  - [ ] Enables downstream converters to work with linearized paths
  - [ ] Maintains path structure and semantics

**Performance Optimization**

- [ ] Efficient curve processing
  - [ ] Curve caching for repeated approximations
  - [ ] Adaptive subdivision based on viewport scale
  - [ ] Fast path for already-linear paths
  - [ ] Memory efficient line segment generation

**Quality vs Performance Trade-offs**

- [ ] Approximation tuning
  - [ ] Scale-dependent approximation quality
  - [ ] Interactive vs high-quality rendering modes
  - [ ] Automatic quality adjustment based on curve size
  - [ ] Performance profiling and optimization

**Advanced Curve Handling**

- [ ] Complex curve scenarios
  - [ ] Self-intersecting curves
  - [ ] Degenerate curves (zero-length control segments)
  - [ ] Curves with coincident control points
  - [ ] Very high curvature curves

**Rendering Integration**

- [ ] Renderer compatibility
  - [ ] Works with all rasterizers expecting line segments
  - [ ] Anti-aliasing preservation during approximation
  - [ ] Consistent stroke width along approximated curves
  - [ ] Smooth gradient application to curved paths

**Curve Length Approximation**

- [ ] Accurate curved path length measurement

  - [ ] Bezier curve length estimation using curve approximation
  - [ ] Integration with ConvCurve for automatic curve-to-line conversion
  - [ ] Adaptive subdivision accuracy based on approximation scale
  - [ ] PathLengthCurved() function that handles both straight and curved segments

- [ ] Length calculation algorithm
  - [ ] Detect curve commands (curve3, curve4) in vertex stream
  - [ ] Use ConvCurve to approximate curves as line segments
  - [ ] Sum line segment lengths for accurate curved path measurement
  - [ ] Maintain existing straight-line performance for non-curved paths

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_vertex_sequence.h → vertex sequence generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- agg_curves.h → curve approximation algorithms

#### agg_conv_bspline.h - B-Spline Converter (smooth curve generation)

**B-Spline Conversion System**

- [x] conv_bspline → ConvBSpline[VS] struct - Convert path to B-spline curves
  - [x] VertexSource template parameter → Input path with control points
  - [x] Inherits from conv_adaptor_vcgen with vcgen_bspline
  - [x] Smooth curve generation through control points
  - [x] Configurable spline parameters

**Spline Parameter Control**

- [x] B-spline configuration
  - [x] interpolation_step(step) → SetInterpolationStep() - distance between output points
  - [x] interpolation_step() → GetInterpolationStep() - get current step size
  - [x] Controls density of generated curve points
  - [x] Smaller steps produce smoother curves

**Curve Quality Settings**

- [ ] Smoothness control
  - [ ] Automatic tangent calculation at control points
  - [ ] Smooth curve transitions between control points
  - [ ] Configurable curve tension (if supported)
  - [ ] Endpoint behavior control (clamped vs natural)

**Control Point Processing**

- [ ] Path interpretation
  - [ ] Path vertices become B-spline control points
  - [ ] line_to commands define control point sequence
  - [ ] move_to commands start new spline curves
  - [ ] Proper handling of multi-path input

**Spline Generation Algorithm**

- [ ] B-spline mathematics
  - [ ] Cubic B-spline basis functions
  - [ ] Control point weight calculation
  - [ ] Knot vector generation (uniform spacing)
  - [ ] Curve parameter to output point mapping

**Applications**

- [ ] Smooth curve creation
  - [ ] Convert angular paths to smooth curves
  - [ ] Data visualization with smooth interpolation
  - [ ] Animation paths with smooth motion
  - [ ] Organic shape generation from rough sketches

**Integration with Other Converters**

- [ ] Converter pipeline integration
  - [ ] Can be combined with stroke for smooth stroked curves
  - [ ] Works with dash patterns for smooth dashed curves
  - [ ] Compatible with curve converter for linearization
  - [ ] Transform integration for scaled splines

**Performance Considerations**

- [ ] Spline calculation efficiency
  - [ ] Coefficient caching for static control points
  - [ ] Incremental spline evaluation
  - [ ] Memory efficient point generation
  - [ ] Adaptive quality based on curve scale

**Advanced Spline Features**

- [ ] Complex spline handling
  - [ ] Multiple spline segments in single path
  - [ ] Closed spline curve generation
  - [ ] Spline continuity control (C0, C1, C2)
  - [ ] Custom knot vector support (if needed)

**Quality Control**

- [ ] Spline approximation
  - [ ] Consistent point spacing along curve
  - [ ] Curvature-adaptive point density
  - [ ] Sharp corner detection and handling
  - [ ] Smooth transitions between spline segments

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_bspline.h → B-spline vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- agg_bspline.h → B-spline mathematical functions

#### agg_conv_smooth_poly1.h - Polygon Smoothing Converter

**Polygon Smoothing System**

- [x] conv_smooth_poly1 → ConvSmoothPoly1[VS] struct - Smooth polygon corners
  - [x] VertexSource template parameter → Input polygon path
  - [x] Inherits from conv_adaptor_vcgen with vcgen_smooth_poly1
  - [x] Corner rounding and smoothing effects
  - [x] Configurable smoothing parameters

**Smoothing Parameters**

- [x] Smoothing control
  - [x] smooth_value(value) → SetSmoothValue() - smoothing amount (0.0 to 1.0)
  - [x] smooth_value() → GetSmoothValue() - get current smoothing level
  - [x] 0.0 produces original polygon, 1.0 produces maximum smoothing
  - [x] Continuous control over smoothing intensity

**Corner Processing Algorithm**

- [ ] Corner detection and modification
  - [ ] Automatic corner angle calculation
  - [ ] Bezier curve generation for rounded corners
  - [ ] Smooth curve insertion at polygon vertices
  - [ ] Preservation of overall polygon shape

**Path Handling**

- [ ] Polygon path processing
  - [ ] Works with both open and closed polygons
  - [ ] Preserves path structure and commands
  - [ ] Maintains polygon orientation
  - [ ] Handles multi-polygon input

**Smoothing Quality**

- [ ] Curve approximation
  - [ ] High-quality Bezier curve approximation
  - [ ] Consistent smoothing across all corners
  - [ ] Adaptive curve density based on corner angle
  - [ ] Smooth transitions between straight and curved segments

**Applications**

- [ ] Shape refinement
  - [ ] Convert sharp-cornered shapes to organic forms
  - [ ] Logo and icon smoothing
  - [ ] Artistic shape modification
  - [ ] UI element softening (rounded rectangles, etc.)

**Advanced Features**

- [ ] Selective smoothing
  - [ ] Different smoothing levels for different corner angles
  - [ ] Sharp corner preservation (below threshold)
  - [ ] Adaptive smoothing based on polygon properties
  - [ ] Custom smoothing profiles

**Integration Capabilities**

- [ ] Converter pipeline
  - [ ] Can be combined with other path modifications
  - [ ] Works well before stroke converter for smooth outlines
  - [ ] Compatible with contour converter for smooth offsets
  - [ ] Integrates with transform operations

**Performance Optimization**

- [ ] Efficient smoothing
  - [ ] Corner detection optimization
  - [ ] Bezier calculation caching
  - [ ] Memory efficient curve generation
  - [ ] Fast path for low smoothing values

**Quality Control**

- [ ] Smoothing precision
  - [ ] Consistent curve quality across smoothing levels
  - [ ] Numerical stability for extreme smoothing values
  - [ ] Preservation of polygon area (approximately)
  - [ ] Smooth derivative transitions

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_smooth_poly1.h → polygon smoothing generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Bezier curve generation utilities

#### agg_conv_clip_polygon.h - Polygon Clipping Converter

**Polygon Clipping System**

- [x] conv_clip_polygon → ConvClipPolygon[VS] struct - Clip polygons to rectangular bounds
  - [x] VertexSource template parameter → Input polygon path to be clipped
  - [x] Inherits from conv_adaptor_vpgen with vpgen_clip_polygon processor
  - [x] Rectangular clipping window definition
  - [x] Liang-Barsky line clipping algorithm (as per AGG implementation)

**Clipping Window Definition**

- [x] Clipping rectangle setup
  - [x] clip_box(x1, y1, x2, y2) → ClipBox() - define rectangular clipping region
  - [x] x1(), y1(), x2(), y2() → Get current clipping bounds
  - [x] Window coordinates in user coordinate system
  - [x] Automatic coordinate ordering (ensures x1 < x2, y1 < y2)

**Polygon Clipping Algorithm**

- [x] Clipping methodology
  - [x] Liang-Barsky line clipping algorithm (per AGG vpgen_clip_polygon)
  - [x] Line-by-line clipping against window boundaries
  - [x] Proper handling of polygon vertices and edges
  - [x] Generation of new intersection vertices via ClipLiangBarsky

**Clipped Output Generation**

- [x] Result polygon creation
  - [x] Maintains polygon structure in output
  - [x] Proper path command generation for clipped polygons
  - [x] Handles multiple disjoint polygon pieces

**Implementation Files:**

- internal/vcgen/clip_polygon.go - VPGenClipPolygon vertex processor generator
- internal/conv/conv_clip_polygon.go - ConvClipPolygon converter wrapper
- Comprehensive tests with clipping scenarios
  - [ ] Preserves polygon winding direction

**Complex Polygon Handling**

- [ ] Advanced clipping scenarios
  - [ ] Self-intersecting polygon clipping
  - [ ] Concave polygon support
  - [ ] Multiple polygon processing
  - [ ] Degenerate case handling (polygon entirely outside clip region)

**Performance Optimizations**

- [ ] Efficient clipping
  - [ ] Early rejection for polygons entirely outside clip region
  - [ ] Early acceptance for polygons entirely inside clip region
  - [ ] Incremental vertex processing
  - [ ] Memory efficient intersection calculation

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vpgen_clip_polygon.h → polygon clipping processor
- agg_conv_adaptor_vpgen.h → vertex processor adaptor
- Geometric intersection utilities

#### agg_conv_clip_polyline.h - Polyline Clipping Converter

**Polyline Clipping System**

- [x] conv_clip_polyline → ConvClipPolyline[VS] struct - Clip polylines to rectangular bounds
  - [x] VertexSource template parameter → Input polyline path to be clipped
  - [x] Inherits from conv_adaptor_vpgen with vpgen_clip_polyline processor
  - [x] Line segment clipping with proper endpoint handling
  - [x] Liang-Barsky line clipping algorithm

**Clipping Window Configuration**

- [x] Clipping bounds setup
  - [x] clip_box(x1, y1, x2, y2) → ClipBox() - define rectangular clipping region
  - [x] X1(), Y1(), X2(), Y2() → get current clipping bounds
  - [x] Coordinate system alignment with polygon clipping
  - [x] Window boundary precision handling

**Line Segment Clipping**

- [x] Segment-by-segment processing
  - [x] Liang-Barsky algorithm for efficient clipping
  - [x] Line-rectangle intersection calculation
  - [x] Proper clipped segment endpoint generation
  - [x] Handling of line segments crossing multiple window boundaries

**Multi-segment Path Handling**

- [x] Complex polyline processing
  - [x] Maintains path structure across clipping operations
  - [x] Proper path command generation for clipped segments
  - [x] Handles disconnected line segments after clipping
  - [x] Path continuity management

**Clipping Edge Cases**

- [x] Special case handling
  - [x] Lines entirely outside clipping region (no output)
  - [x] Lines entirely inside clipping region (pass through)
  - [x] Lines tangent to clipping window boundaries
  - [x] Zero-length line segments

**Performance Features**

- [x] Optimized line clipping
  - [x] Fast accept/reject using clipping flags
  - [x] Minimal intersection calculations
  - [x] Efficient vertex generation
  - [x] Memory-friendly processing for long polylines

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] agg_vpgen_clip_polyline.h → polyline clipping processor (VPGenClipPolyline)
- [x] agg_conv_adaptor_vpgen.h → vertex processor adaptor (ConvAdaptorVPGen)
- [x] Line-rectangle intersection algorithms (ClipLineSegment)

#### agg_conv_gpc.h - General Polygon Clipper Converter

**GPC Integration System**

- [ ] conv_gpc → ConvGPC[VS] struct - General Polygon Clipper integration
  - [ ] VertexSource template parameter → Input polygon for boolean operations
  - [ ] Integration with GPC (General Polygon Clipper) library
  - [ ] Support for complex boolean operations on polygons
  - [ ] Industrial-strength polygon clipping capabilities

**Boolean Operations**

- [ ] Polygon boolean operations
  - [ ] Union (OR) - combine polygon areas
  - [ ] Intersection (AND) - common polygon areas only
  - [ ] Difference (A - B) - subtract second polygon from first
  - [ ] XOR (exclusive OR) - symmetric difference of polygons

**GPC Library Integration**

- [ ] External library interface
  - [ ] GPC polygon structure conversion
  - [ ] Memory management for GPC data structures
  - [ ] Error handling and status reporting
  - [ ] License compliance considerations (GPC has specific licensing)

**Complex Polygon Support**

- [ ] Advanced polygon features
  - [ ] Self-intersecting polygon handling
  - [ ] Polygon with holes support
  - [ ] Multiple contour polygons
  - [ ] Arbitrary polygon complexity

**Performance Considerations**

- [ ] GPC performance characteristics
  - [ ] Memory usage for complex polygons
  - [ ] Processing time for large polygon sets
  - [ ] Optimization strategies for common cases
  - [ ] Alternative lightweight clipping for simple cases

**Integration Challenges**

- [ ] Library compatibility
  - [ ] C library integration in Go
  - [ ] Memory management across language boundaries
  - [ ] Error propagation from C library
  - [ ] Platform-specific compilation considerations

**Dependencies**

- agg_basics.h → internal/basics package
- gpc.h and gpc.c → General Polygon Clipper library
- C library integration utilities
- Memory management for cross-language data

#### agg_conv_close_polygon.h - Polygon Closing Converter

**Polygon Closing System**

- [x] conv_close_polygon → ConvClosePolygon[VS] struct - Ensure polygons are properly closed
  - [x] VertexSource template parameter → Input path that may have unclosed polygons
  - [x] Automatic detection of unclosed polygon paths
  - [x] Addition of closing EndPoly commands with close flags where needed
  - [x] Preservation of already-closed polygons

**Polygon Closure Detection**

- [ ] Path analysis
  - [ ] Detection of polygon paths (sequences ending with end_poly)
  - [ ] Comparison of first and last vertex coordinates
  - [ ] Epsilon-based coordinate comparison for floating-point precision
  - [ ] Path command sequence analysis

**Automatic Closure Generation**

- [ ] Path modification
  - [ ] Insertion of line_to command from last to first vertex
  - [ ] Proper close_polygon command generation
  - [ ] Preservation of polygon orientation (clockwise/counter-clockwise)
  - [ ] Path command sequence integrity

**Multi-polygon Support**

- [ ] Complex path handling
  - [ ] Multiple polygon processing in single path
  - [ ] Independent closure analysis for each polygon
  - [ ] Preservation of path structure and organization
  - [ ] Proper handling of move_to commands between polygons

**Integration with Rendering**

- [ ] Renderer compatibility
  - [ ] Ensures proper polygon filling
  - [ ] Eliminates gaps in polygon outlines
  - [ ] Compatible with all fill rule algorithms
  - [ ] Works with stroke converters for complete outlines

**Performance Optimization**

- [ ] Efficient closure processing
  - [ ] Single-pass path analysis
  - [ ] Minimal vertex coordinate comparison
  - [ ] Fast path for already-closed polygons
  - [ ] Memory-efficient path modification

**Dependencies**

- agg_basics.h → internal/basics package
- Path command definitions and utilities
- Coordinate comparison utilities with epsilon handling

#### agg_conv_unclose_polygon.h - Polygon Unclosing Converter

**Polygon Unclosing System**

- [x] conv_unclose_polygon → ConvUnclosePolygon[VS] struct - Remove polygon closing commands
  - [x] VertexSource template parameter → Input path with closed polygons
  - [x] Detection and removal of polygon closing commands
  - [x] Conversion of closed polygons to open polylines
  - [x] Preservation of polygon shape without closure

**Closure Command Removal**

- [x] Path command filtering
  - [x] Detection of close_polygon commands
  - [x] Removal of explicit closing line segments
  - [x] Conversion of end_poly with close flag to end_poly without close flag
  - [x] Maintenance of path vertex sequence

**Path Structure Preservation**

- [ ] Path integrity maintenance
  - [ ] Preservation of all polygon vertices except closure
  - [ ] Maintenance of path organization and multi-polygon structure
  - [ ] Proper handling of path IDs and command sequences
  - [ ] Coordinate precision preservation

**Applications**

- [ ] Open path creation
  - [ ] Convert polygons to polylines for stroke-only rendering
  - [ ] Prepare paths for dash pattern application
  - [ ] Create open paths for text-on-path applications
  - [ ] Generate paths for marker placement

**Integration Patterns**

- [ ] Common usage scenarios
  - [ ] Used before dash converter to avoid closing dashes
  - [ ] Applied before certain marker converters
  - [ ] Useful in path editing applications
  - [ ] Compatible with all other path converters

**Performance Features**

- [ ] Efficient processing
  - [ ] Single-pass path command filtering
  - [ ] Minimal memory overhead
  - [ ] Fast identification of closing commands
  - [ ] Stream-friendly processing

**Dependencies**

- agg_basics.h → internal/basics package
- Path command definitions and filtering utilities

#### agg_conv_concat.h - Path Concatenation Converter

**Path Concatenation System**

- [x] conv_concat → ConvConcat[VS1, VS2] struct - Concatenate multiple vertex sources
  - [x] Multiple VertexSource template parameters → Input paths to concatenate
  - [x] Sequential path output from multiple sources
  - [x] Proper path command sequence management
  - [x] Support for arbitrary number of source paths

**Multi-source Management**

- [x] Source path handling
  - [x] Sequential processing of input vertex sources
  - [x] Automatic source switching at path completion
  - [x] Path ID management across multiple sources
  - [x] State management for source transitions

**Path Command Sequencing**

- [x] Command stream management
  - [x] Proper command sequence across source boundaries
  - [x] Path continuity or separation control
  - [x] end_poly command handling between sources
  - [x] move_to command insertion for path separation

**Concatenation Modes**

- [x] Different concatenation strategies
  - [ ] Continuous concatenation (paths connected)
  - [x] Separate concatenation (paths as distinct entities)
  - [x] Custom separation control
  - [x] Path orientation preservation

**Applications**

- [ ] Complex path construction
  - [ ] Combine multiple path segments into single path
  - [ ] Merge paths from different sources
  - [ ] Create composite shapes from multiple components
  - [ ] Path assembly for complex graphics

**Performance Considerations**

- [ ] Efficient concatenation
  - [ ] Minimal overhead for source switching
  - [ ] Memory-efficient vertex streaming
  - [ ] Fast source completion detection
  - [ ] Optimized for large numbers of source paths

**Dependencies**

- agg_basics.h → internal/basics package
- Multiple vertex source management utilities
- Path command sequencing logic

#### agg_conv_shorten_path.h - Path Shortening Converter

**Path Shortening System**

- [x] conv_shorten_path → ConvShortenPath[VS] struct - Shorten paths by removing segments from ends
  - [x] VertexSource template parameter → Input path to be shortened
  - [x] Configurable shortening amounts for path end (C++ only shortens from end)
  - [x] Precise length-based path modification
  - [x] Preservation of path shape and direction

**Shortening Configuration**

- [ ] Length control parameters
  - [ ] shorten(amount) → SetShorten() - symmetric shortening amount
  - [ ] shorten_start(amount) → SetShortenStart() - shortening from path start
  - [ ] shorten_end(amount) → SetShortenEnd() - shortening from path end
  - [ ] Independent control of start and end shortening

**Path Length Calculation**

- [ ] Arc-length measurement
  - [ ] Precise path length calculation
  - [ ] Cumulative distance tracking along path
  - [ ] Curve length approximation for curved paths
  - [ ] Segment-by-segment length accumulation

**Shortening Algorithm**

- [ ] Path modification process
  - [ ] Forward shortening from path start
  - [ ] Backward shortening from path end
  - [ ] Interpolation for partial segment removal
  - [ ] New vertex generation at cut points

**Complex Path Handling**

- [ ] Multi-segment path support
  - [ ] Shortening across multiple path segments
  - [ ] Proper handling of move_to commands
  - [ ] Preservation of path structure
  - [ ] Multi-path processing

**Edge Case Management**

- [ ] Boundary condition handling
  - [ ] Shortening amount exceeding path length
  - [ ] Zero-length path handling
  - [ ] Very short path segments
  - [ ] Degenerate path processing

**Applications**

- [ ] Path end modification
  - [ ] Arrow integration (shorten path for arrowhead placement)
  - [ ] Marker positioning (create space for end markers)
  - [ ] Path trimming for aesthetic purposes
  - [ ] Animation effects (growing/shrinking paths)

**Performance Optimization**

- [ ] Efficient shortening
  - [ ] Incremental length calculation
  - [ ] Early termination for excessive shortening
  - [ ] Memory-efficient vertex generation
  - [ ] Caching for repeated operations

**Dependencies**

- agg_basics.h → internal/basics package
- agg_shorten_path.h → path shortening utilities
- Arc-length calculation algorithms

#### agg_conv_segmentator.h - Segmentator Converter

**Path Segmentation System**

- [x] conv_segmentator → ConvSegmentator[VS] struct - Segment paths into equal-length pieces
  - [x] VertexSource template parameter → Input path to be segmented
  - [x] Inherits from conv_adaptor_vpgen with vpgen_segmentator processor
  - [x] Configurable segment length
  - [x] Even spacing along path curves and lines

**Segmentation Parameters**

- [ ] Segment control
  - [ ] approximation_scale(scale) → SetApproximationScale() - control segment density
  - [ ] approximation_scale() → GetApproximationScale() - get current scale
  - [ ] Uniform segment length along entire path
  - [ ] Adaptive segmentation based on path curvature

**Path Parameterization**

- [ ] Arc-length parameterization
  - [ ] Conversion of path to arc-length parameter space
  - [ ] Even spacing calculation along path
  - [ ] Interpolation between original path vertices
  - [ ] Precise positioning of segment boundaries

**Segment Generation**

- [ ] Uniform vertex output
  - [ ] Regular vertex spacing along path
  - [ ] Interpolated coordinates for segment points
  - [ ] Preservation of path direction and orientation
  - [ ] Smooth segment transitions

**Applications**

- [ ] Path sampling and analysis
  - [ ] Create evenly-spaced sample points along paths
  - [ ] Animation keyframe generation
  - [ ] Path analysis and measurement
  - [ ] Uniform marker placement preparation

**Advanced Features**

- [ ] Segmentation quality control
  - [ ] Adaptive segment length based on curvature
  - [ ] Minimum/maximum segment length constraints
  - [ ] Corner and cusp handling
  - [ ] Smooth segment transitions at path joints

**Performance Considerations**

- [ ] Efficient segmentation
  - [ ] Incremental arc-length calculation
  - [ ] Memory-efficient parameter space conversion
  - [ ] Fast interpolation algorithms
  - [ ] Optimized for long paths

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vpgen_segmentator.h → segmentation processor
- agg_conv_adaptor_vpgen.h → vertex processor adaptor
- Arc-length parameterization utilities

#### agg_conv_marker.h - Marker Converter ✅ **COMPLETED**

**Marker Placement System**

- [x] conv_marker → ConvMarker struct - Place markers along paths (`internal/conv/conv_marker.go`)
  - [x] MarkerLocator interface → Marker positioning strategy
  - [x] MarkerShapes interface → Marker geometry definition
  - [x] Automatic marker orientation along path direction
  - [x] State machine for marker placement processing with transformation support

**Marker Locator Strategies** - Basic Interface Implemented

- [x] Positioning interface
  - [x] Rewind() method for marker positioning
  - [x] Vertex() method for marker position and direction pairs
  - [x] Extensible design for custom placement patterns

**Marker Shape Integration** - Core Functionality Implemented

- [x] Shape definition and rendering
  - [x] Marker geometry from MarkerShapes interface
  - [x] Automatic marker orientation along path direction
  - [x] Scale and transformation support via TransAffine
  - [x] Marker coordinate system transformation

**Path Analysis for Marker Placement**

- [ ] Path processing
  - [ ] Arc-length calculation for distance-based placement
  - [ ] Path direction calculation for marker orientation
  - [ ] Curvature analysis for marker scaling
  - [ ] Path segment analysis for placement optimization

**Advanced Marker Features**

- [ ] Complex marker behaviors
  - [ ] Marker collision detection and avoidance
  - [ ] Variable marker size along path
  - [ ] Marker fade-in/fade-out effects
  - [ ] Custom marker transformation functions

**Performance Optimization**

- [ ] Efficient marker generation
  - [ ] Marker shape caching and reuse
  - [ ] Batch marker transformation
  - [ ] Memory-efficient marker geometry storage
  - [ ] Optimized placement calculation

**Dependencies**

- agg_basics.h → internal/basics package
- agg_conv_adaptor_vcgen.h → vertex generator adaptor
- Marker locator and shape interfaces
- Path analysis and transformation utilities

#### agg_conv_marker_adaptor.h - Marker Adaptor Converter ✅ **COMPLETED**

**Marker Adaptor System**

- [x] conv_marker_adaptor → ConvMarkerAdaptor struct - Adaptor for custom marker systems (`internal/conv/conv_marker_adaptor.go`)
  - [x] VertexSource interface → Input path for marker placement
  - [x] Markers interface → Custom marker implementation via ConvAdaptorVCGen
  - [x] Bridge between path processing and marker systems using VCGenVertexSequence
  - [x] Flexible marker integration framework with shortening support

**Custom Marker Integration**

- [x] Marker system compatibility
  - [x] Support for user-defined marker implementations via Markers interface
  - [x] Marker lifecycle management through ConvAdaptorVCGen
  - [x] State synchronization between path and markers
  - [x] Shortening functionality for marker positioning

**Marker Event Handling**

- [x] Path-to-marker communication
  - [x] Path vertex events to marker system via ConvAdaptorVCGen
  - [x] Path command interpretation for markers
  - [x] Marker preparation and cleanup phases through base adaptor
  - [x] Standard vertex source interface implementation

**Flexible Marker Framework**

- [ ] Extensible marker support
  - [ ] Template-based marker customization
  - [ ] Runtime marker behavior modification
  - [ ] Multiple marker type support
  - [ ] Marker composition and layering

**Performance Features**

- [ ] Efficient marker processing
  - [ ] Minimal overhead marker integration
  - [ ] Lazy marker evaluation
  - [ ] Memory-efficient marker state management
  - [ ] Optimized marker update cycles

**Dependencies**

- agg_basics.h → internal/basics package
- Marker interface definitions
- Vertex source integration utilities

#### agg_conv_transform.h - Transform Converter ✅ **COMPLETED**

**Transform Converter System**

- [x] conv_transform → ConvTransform[VS, Trans] struct - Apply transformations to vertex sources (`internal/conv/conv_transform.go`)
  - [x] VertexSource template parameter → Input path to be transformed
  - [x] Transformer template parameter → Transformation implementation
  - [x] Real-time coordinate transformation
  - [x] Compatible with all transformation types
  - [x] AGG-compatible API (Transformer method in addition to SetTransformer)
  - [x] Comprehensive test coverage with edge cases and integration tests

**Transformation Integration**

- [ ] Transform compatibility
  - [ ] Works with trans_affine for 2D affine transformations
  - [ ] Supports trans_perspective for perspective projection
  - [ ] Compatible with trans_bilinear for quadrilateral mapping
  - [ ] Integrates with custom transformation implementations

**Real-time Transformation**

- [ ] Streaming coordinate transformation
  - [ ] Transform coordinates as vertices are requested
  - [ ] No intermediate storage required
  - [ ] Memory-efficient transformation
  - [ ] Preserves path structure and commands

**Transform Parameter Control**

- [ ] Transformer access
  - [ ] transformer() → GetTransformer() - access underlying transformer
  - [ ] const transformer() → GetTransformer() const - read-only access
  - [ ] Direct parameter modification through transformer interface
  - [ ] Real-time transformation parameter updates

**Path Command Preservation**

- [ ] Command stream integrity
  - [ ] Preserves all path commands (move_to, line_to, etc.)
  - [ ] Maintains path structure and organization
  - [ ] Proper path ID propagation
  - [ ] Command sequence preservation

**Performance Optimization**

- [ ] Efficient transformation
  - [ ] Single-pass coordinate transformation
  - [ ] Minimal computational overhead
  - [ ] Memory-friendly streaming operation
  - [ ] Optimized for interactive transformation

**Advanced Transformation Features**

- [ ] Complex transformation scenarios
  - [ ] Nested transformation support
  - [ ] Transform composition through chaining
  - [ ] Interactive transformation updates
  - [ ] Animation-friendly transformation

**Integration with Rendering Pipeline**

- [ ] Renderer compatibility
  - [ ] Compatible with all AGG renderers
  - [ ] Works with rasterizers and scanline renderers
  - [ ] Integrates with anti-aliasing systems
  - [ ] Supports all pixel formats

**Dependencies**

- agg_basics.h → internal/basics package
- Transformation interface definitions (trans_affine, trans_perspective, etc.)
- Coordinate transformation utilities

---

### Vertex Generators

Vertex generators transform input path vertices into specialized output sequences for rendering operations like stroking, dashing, and contouring. They sit between path definition and rasterization in the AGG pipeline.

**Overview**

Vertex generators implement the "Vertex Generator Interface" with `remove_all()`, `add_vertex()` methods and "Vertex Source Interface" with `rewind()`, `vertex()` methods. They process input vertices to create modified vertex streams for specific rendering effects.

**Key Components**

- [x] **agg_vcgen_bspline.h** - B-spline curve generator (`internal/vcgen/bspline.go`)

  - Converts control points to smooth B-spline curves
  - Configurable interpolation step for curve resolution
  - Status: ✅ **COMPLETED** - Basic implementation with tests

- [x] **agg_vcgen_smooth_poly1.h** - Polygon smoothing generator (`internal/vcgen/smooth_poly1.go`)

  - Smooths polygon vertices using curve interpolation
  - Creates rounded corners and smooth transitions
  - Status: ✅ **COMPLETED** - Basic implementation with tests

- [x] **agg_vcgen_vertex_sequence.h** - Vertex sequence manager (`internal/vcgen/vertex_sequence.go`)

  - Manages vertex sequences with distance calculations
  - Supports path shortening and vertex filtering
  - Status: ✅ **COMPLETED** - Basic implementation with tests

- [ ] **agg_vcgen_stroke.h** - Stroke generator (`internal/vcgen/stroke.go`)

  - Converts single-line paths into stroked outlines
  - Implements line caps (butt, round, square) and joins (miter, round, bevel)
  - Handles stroke width, miter limits, and inner joins
  - Status: ❌ **PENDING** - Critical for basic drawing operations

- [x] **agg_vcgen_dash.h** - Dash pattern generator (`internal/vcgen/dash.go`)

  - Creates dashed line patterns from solid paths
  - Configurable dash lengths and gap patterns
  - Maintains dash phase across path segments
  - Status: ✅ **COMPLETED** - Important for styled line drawing

- [ ] **agg_vcgen_contour.h** - Contour generator (`internal/vcgen/contour.go`)

  - Generates parallel contours (outlines) from paths
  - Creates offset curves at specified distances
  - Similar to stroke but for filled shapes
  - Status: ❌ **PENDING** - Needed for advanced text and shape effects

- [x] **agg_vcgen_markers_term.h** - Terminal markers generator (`internal/vcgen/markers_term.go`)
  - Generates terminal markers (arrowheads, tails) for paths
  - Places markers at path start/end points
  - Calculates marker orientation based on path direction
  - Status: ✅ **COMPLETED** - Needed for arrows and decorative elements

**Implementation Requirements**

1. **Core Interface Implementation**

   - Vertex Generator Interface: `RemoveAll()`, `AddVertex(x, y float64, cmd uint)`
   - Vertex Source Interface: `Rewind(pathID uint)`, `Vertex() (x, y float64, cmd uint)`
   - State management with status enums (initial, ready, processing, stop)

2. **Stroke Generator (Priority: HIGH)**

   - Line cap types: butt, round, square
   - Line join types: miter, round, bevel, miter_revert
   - Inner join handling for acute angles
   - Stroke width and miter limit parameters
   - Integration with math_stroke calculations

3. **Dash Generator (Priority: HIGH)**

   - Dash pattern array support (up to 32 patterns)
   - Dash start offset and scaling
   - State tracking across path segments
   - Pattern cycling and phase management

4. **Contour Generator (Priority: MEDIUM)**

   - Positive/negative width for inward/outward contours
   - Integration with stroke math for offset calculations
   - Closed polygon handling
   - Self-intersection resolution

5. **Markers Generator (Priority: MEDIUM)**
   - Coordinate storage for marker positions
   - Orientation calculation from path tangents
   - Multiple marker support along paths
   - Marker transformation interface

**Missing Dependencies**

- [ ] **agg_math_stroke.h** → `internal/basics/math_stroke.go` (526 lines)

  - Line cap/join calculations
  - Stroke mathematics for width application
  - Miter limit and angle calculations
  - **Required for**: vcgen_stroke, vcgen_contour

- [ ] **agg_shorten_path.h** → `internal/basics/shorten_path.go` (66 lines)

  - Path shortening utilities
  - Used by vcgen_vertex_sequence for path end adjustments
  - **Required for**: Enhanced vertex_sequence functionality

- [ ] **vertex_dist types** → Enhance `internal/basics/types.go`
  - `VertexDist` struct (x, y, dist fields)
  - `VertexDistCmd` struct (extends VertexDist with cmd field)
  - Distance calculation methods
  - **Required for**: All vertex generators using distance-based filtering

**Testing Requirements**

1. **Unit Tests**

   - Stroke generation with different cap/join combinations
   - Dash pattern generation and phase management
   - Contour generation with positive/negative widths
   - Marker placement and orientation accuracy

2. **Integration Tests**

   - Vertex generator chaining (dash → stroke)
   - Integration with rasterizer pipeline
   - Performance testing with complex paths
   - Memory usage validation

3. **Visual Tests**
   - Stroke rendering with various parameters
   - Dash pattern visual verification
   - Contour accuracy against reference
   - Marker placement and scaling

**Integration Points**

- **Input**: Path vertices from `types.Path` or transformation pipeline
- **Output**: Modified vertex streams to rasterizer or scanline renderer
- **Conversion Chain**: Path → [Transform] → [VCGen] → [Conv] → Rasterizer
- **Renderer Compatibility**: Must work with all AGG renderers and pixel formats

**Performance Considerations**

- Vertex generators process large vertex streams
- Memory pooling for intermediate vertex storage
- Efficient state machine implementation
- SIMD optimization opportunities for math operations

**Dependencies**

- `internal/basics` → Basic types, constants, and math utilities
- `internal/array` → Vertex sequence and storage containers
- Future: `internal/basics/math_stroke.go` → Stroke mathematics
- Future: `internal/basics/shorten_path.go` → Path manipulation utilities

---

### Vertex Processors

- [x] agg_vpgen_clip_polygon.h - Polygon clipping vertex processor
- [x] agg_vpgen_clip_polyline.h - Polyline clipping vertex processor
- [x] agg_vpgen_segmentator.h - Segmentator vertex processor

---

### Spans and Gradients

#### agg_span_allocator.h - Memory allocation for color spans ✅ **COMPLETED**

**Core Span Allocator Structure**

- [x] span_allocator → SpanAllocator[C] struct - Memory allocator for color spans (`internal/spans/allocator.go`)
  - [x] Generic over color type C (RGBA8, Gray8, etc.)
  - [x] Dynamic memory allocation with size optimization
  - [x] 256-element alignment for reduced reallocations
  - [x] Reusable span buffer management

**Memory Management Methods**

- [x] Allocation interface
  - [x] allocate(span_len) → Allocate(spanLen) - allocate span buffer of specified length
  - [x] span() → Span() - get pointer to current span buffer
  - [x] max_span_len() → MaxSpanLen() - get maximum allocated span length
- [x] Memory optimization
  - [x] Size alignment to 256-color boundaries
  - [x] Reuse existing buffer when size permits
  - [x] Minimal reallocation strategy

**Integration with Rendering Pipeline**

- [x] Span generator compatibility
  - [x] Works with all span generator types
  - [x] Provides temporary storage for scanline rendering
  - [x] Thread-safe allocation patterns
- [x] Memory efficiency
  - [x] Single allocation per scanline
  - [x] Persistent buffer reuse across multiple scanlines
  - [x] Automatic cleanup and garbage collection

**Dependencies**

- [x] agg_array.h → internal/array package (PodArray implementation)
- [x] Color type definitions from basics package

#### agg_span_converter.h - Span conversion pipeline

**Span Converter System**

- [x] span_converter → SpanConverter[SG, SC] struct - Pipeline for span processing
  - [x] SpanGenerator template parameter → Input span generator type
  - [x] SpanConverter template parameter → Conversion function type
  - [x] Two-stage processing: generation then conversion
  - [x] Composable conversion pipeline architecture

**Generator and Converter Management**

- [x] Component attachment
  - [x] attach_generator(span_gen) → AttachGenerator() - connect span generator
  - [x] attach_converter(span_cnv) → AttachConverter() - connect span converter
  - [x] Dynamic component swapping during rendering
  - [x] Type-safe template composition

**Conversion Pipeline Methods**

- [x] Processing interface
  - [x] prepare() → Prepare() - initialize both generator and converter
  - [x] generate(span, x, y, len) → Generate() - two-stage span processing
  - [x] Coordinated preparation of pipeline components
  - [x] Sequential processing with intermediate span buffer

**Pipeline Composition Patterns**

- [x] Multi-stage conversion
  - [x] Chain multiple converters together
  - [x] Color space conversion (RGB → CMYK, etc.)
  - [x] Gamma correction and color management
  - [x] Alpha blending and compositing operations
- [x] Performance optimization
  - [x] Single-pass processing where possible
  - [x] Minimal intermediate buffer allocation
  - [x] Inline conversion for simple operations

**Integration Points**

- [x] Compatible generators
  - [x] Gradient generators (linear, radial, conic)
  - [x] Pattern generators (image, texture)
  - [x] Solid color generators
  - [x] Gouraud shading generators
- [x] Compatible converters
  - [x] Color space converters
  - [x] Alpha manipulation converters
  - [x] Dithering and quantization converters
  - [x] Custom effect converters

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] Color type system compatibility
- [x] Span generator interface definitions

#### agg_span_solid.h - Solid color span generation ✅ **COMPLETED**

**Note**: This component is marked as completed. Implementation should be verified in `internal/spans/solid.go` to ensure all functionality matches AGG specification.

**Verification Checklist**

- [x] span_solid → SpanSolid[C] struct - generates uniform color spans
- [x] Single color fill across entire span length
- [x] Efficient constant-time generation
- [x] Integration with span allocator system

#### agg_span_gradient.h - Gradient span generation ✅ COMPLETED

**Core Gradient Span System**

- [x] span_gradient → SpanGradient[C, I, GF, CF] struct - Multi-parameter gradient generator
  - [x] ColorT template parameter → Output color type (RGBA8, etc.)
  - [x] Interpolator template parameter → Coordinate transformation
  - [x] GradientF template parameter → Gradient shape function (linear, radial, etc.)
  - [x] ColorF template parameter → Color lookup function (LUT, procedural)

**Gradient Subpixel Precision**

- [x] Subpixel coordinate system
  - [x] gradient_subpixel_shift → GradientSubpixelShift constant (4 bits)
  - [x] gradient_subpixel_scale → GradientSubpixelScale constant (16x precision)
  - [x] gradient_subpixel_mask → GradientSubpixelMask for coordinate masking
  - [x] High-precision gradient calculations

**Constructor and Configuration**

- [x] Initialization methods
  - [x] Default constructor → NewSpanGradient() - uninitialized state
  - [x] Parameter constructor with interpolator, gradient function, color function
  - [x] Distance range configuration (d1, d2) for gradient mapping
  - [x] Component attachment and detachment

**Component Access and Management**

- [x] Component accessors
  - [x] interpolator() → Interpolator() - get/set coordinate interpolator
  - [x] gradient_function() → GradientFunction() - get/set gradient shape function
  - [x] color_function() → ColorFunction() - get/set color lookup function
  - [x] d1(), d2() → D1(), D2() - get/set distance range parameters

**Gradient Generation Pipeline**

- [x] Span generation method
  - [x] generate(span, x, y, len) → Generate() - produce gradient-filled span
  - [x] Coordinate interpolation for each pixel in span
  - [x] Gradient distance calculation using shape function
  - [x] Color lookup and mapping to output span
- [x] Distance mapping
  - [x] Linear mapping from gradient distance to color index
  - [x] Range clamping to prevent color table overflow
  - [x] Subpixel precision preservation

**Gradient Shape Functions Integration**

- [x] Shape function interface
  - [x] calculate(x, y, d2) → Calculate() - compute gradient distance
  - [x] Support for linear, radial, conic, and custom gradients
  - [x] Coordinate system compatibility
  - [x] Performance optimization for common shapes

**Color Function Integration**

- [x] Color lookup interface
  - [x] size() → Size() - get color table size
  - [x] Color indexing and interpolation
  - [x] Support for gradient LUT (lookup table)
  - [x] Procedural color generation

**Performance Optimizations**

- [x] Efficient span processing
  - [x] Single interpolator begin() call per span
  - [x] Incremental coordinate calculation
  - [x] Minimal per-pixel overhead
  - [x] Branch reduction in inner loops
- [x] Subpixel precision management
  - [x] Bit shifting for coordinate downscaling
  - [x] Integer arithmetic where possible
  - [x] Avoiding floating-point in inner loops

**Dependencies**

- agg_basics.h → internal/basics package
- agg_math.h → internal/math package
- agg_array.h → internal/array package
- Interpolator interface definitions
- Gradient function implementations
- Color function implementations

**Implementation Files**

- `internal/span/span_gradient.go` - Main gradient span generator and shape functions
- `internal/span/interpolator_linear.go` - Linear span interpolator with affine transform support
- `internal/span/span_gradient_test.go` - Comprehensive test coverage
- `internal/span/interpolator_linear_test.go` - Interpolator test coverage

**Implemented Gradient Shape Functions**

- GradientLinearX, GradientLinearY - Linear gradients
- GradientRadial, GradientRadialDouble - Circular gradients
- GradientRadialFocus - Radial gradient with focal point
- GradientDiamond, GradientXY, GradientSqrtXY - Custom shapes
- GradientConic - Angular/conic gradients
- GradientRepeatAdaptor, GradientReflectAdaptor - Wrapping modes

#### agg_span_gradient_alpha.h - Alpha-only gradient generation ✅ COMPLETED

**Alpha Gradient Span System**

- [x] span_gradient_alpha → SpanGradientAlpha[I, GF, AF] struct - Alpha-only gradient generator
  - [x] Interpolator template parameter → Coordinate transformation system
  - [x] GradientF template parameter → Gradient shape function
  - [x] AlphaF template parameter → Alpha lookup function or LUT
  - [x] 8-bit alpha value output (0-255 range)

**Alpha Generation Methods**

- [x] Core alpha generation
  - [x] generate(alpha_span, x, y, len) → Generate() - produce alpha-only span
  - [x] Single-channel alpha output instead of full color
  - [x] Distance calculation using gradient function
  - [x] Alpha lookup using alpha function

**Alpha Function Integration**

- [x] Alpha lookup interface
  - [x] Support for alpha LUT (lookup tables)
  - [x] Linear alpha interpolation
  - [x] Custom alpha mapping functions
  - [x] Procedural alpha generation
- [x] Alpha precision
  - [x] 8-bit alpha values (0-255)
  - [x] Smooth alpha transitions
  - [x] Anti-aliased alpha boundaries

**Specialized Use Cases**

- [x] Alpha masking applications
  - [x] Soft-edged masks using gradient shapes
  - [x] Feathered selection boundaries
  - [x] Transparency effects and fading
  - [x] Complex mask composition
- [x] Performance benefits
  - [x] Single channel processing (faster than full color)
  - [x] Reduced memory bandwidth
  - [x] Optimized for masking operations

**Integration with Rendering Pipeline**

- [x] Alpha application methods
  - [x] Compatible with alpha blending systems
  - [x] Mask overlay operations
  - [x] Alpha channel compositing
  - [x] Transparency rendering support

**Dependencies**

- Gradient span base functionality
- Alpha function interface definitions
- Interpolator system
- Gradient shape functions

**Implementation Files**

- `internal/span/span_gradient_alpha.go` - Alpha gradient span generator and alpha functions
- `internal/span/span_gradient_alpha_test.go` - Comprehensive test coverage with benchmarks

**Implemented Alpha Functions**

- GradientAlphaLinear - Linear alpha interpolation between two values
- GradientAlphaX - Identity alpha function (pass-through)
- GradientAlphaOneMinusX - Inverse alpha function (255-x)
- GradientAlphaLUT - Custom lookup table for alpha values

**Alpha Wrapper Types**

- RGBA8AlphaWrapper[CS] - Wraps RGBA8 colors for alpha manipulation
- Gray8AlphaWrapper[CS] - Wraps Gray8 colors for alpha manipulation
- Helper functions for creating gradient configurations

#### agg_span_gradient_contour.h - Contour-based gradient generation ✅ **COMPLETED**

**Contour Gradient System**

- [x] gradient_contour → GradientContour struct - Core distance field gradient generator (`internal/span/span_gradient_contour.go`)
  - [x] Distance transform algorithm (Pedro Felzenszwalb) implementation
  - [x] Path-based contour rasterization and distance field generation
  - [x] Buffer management for grayscale distance maps
  - [x] Calculate method for gradient distance lookup with subpixel precision
  - [ ] span_gradient_contour → SpanGradientContour[I, CF] struct - Template wrapper (future enhancement)

**Distance Field Calculation**

- [x] Core distance methods
  - [x] Distance calculation from path boundaries via rasterization
  - [x] 2D distance transform using separable 1D transforms
  - [x] Path vertex adapter for interface compatibility
  - [ ] Multi-contour distance field composition
  - [ ] Anti-aliased contour boundaries
- [ ] Distance field preprocessing
  - [ ] Contour rasterization to distance field
  - [ ] Distance field caching and optimization
  - [ ] Real-time vs. precomputed distance fields

**Contour Definition Interface**

- [ ] Contour input methods
  - [ ] Vector path contour definition
  - [ ] Bitmap contour extraction
  - [ ] Bezier curve contour support
  - [ ] Complex shape decomposition
- [ ] Multi-contour support
  - [ ] Multiple gradient sources
  - [ ] Contour priority and blending
  - [ ] Hierarchical contour systems

**Advanced Distance Field Features**

- [ ] Distance field operations
  - [ ] Union, intersection, and subtraction
  - [ ] Distance field morphology (expand, contract)
  - [ ] Smooth blending between contours
  - [ ] Distance field animation support
- [ ] Optimization techniques
  - [ ] Distance field approximation methods
  - [ ] Hierarchical distance field evaluation
  - [ ] Spatial acceleration structures

**Gradient Mapping**

- [ ] Distance-to-color mapping
  - [ ] Linear distance mapping to color function
  - [ ] Non-linear distance curves
  - [ ] Multiple gradient zones
  - [ ] Smooth color transitions
- [ ] Edge handling
  - [ ] Soft edge gradients
  - [ ] Sharp contour boundaries
  - [ ] Edge thickness control

**Dependencies**

- Gradient span base system
- Path and contour definitions
- Distance field calculation utilities
- Color function interface

#### agg_span_gradient_image.h - Image-based gradient generation

**Image Gradient System**

- [x] span_gradient_image → GradientImageRGBA8 struct - Image-derived gradient generator
  - [x] RGBA8 color buffer support → Internal image buffer management
  - [x] Image sampling for gradient generation → Calculate method with coordinate wrapping
  - [x] Texture-based gradient effects → Coordinate wrapping for tiling behavior
  - [x] Real-time image gradient computation → Efficient pixel sampling with subpixel precision

**Image Sampling Methods**

- [ ] Pixel sampling interface
  - [ ] Nearest neighbor sampling
  - [ ] Bilinear interpolation sampling
  - [ ] Custom sampling filters
  - [ ] Edge handling (clamp, wrap, mirror)
- [ ] Image coordinate mapping
  - [ ] UV coordinate generation
  - [ ] Image bounds checking
  - [ ] Coordinate transformation chain

**Image Data Management**

- [ ] Image source interface
  - [ ] Compatible with various image formats
  - [ ] Dynamic image loading and caching
  - [ ] Image preprocessing for gradient use
  - [ ] Memory-efficient image access
- [ ] Image transformation
  - [ ] Rotation, scaling, and translation
  - [ ] Perspective transformation
  - [ ] Image warping and distortion

**Gradient Extraction from Images**

- [ ] Image-to-gradient conversion
  - [ ] Luminance-based gradient extraction
  - [ ] Color channel selection for gradient
  - [ ] Custom image analysis functions
  - [ ] Multi-channel gradient generation
- [ ] Image processing pipeline
  - [ ] Image filtering before gradient extraction
  - [ ] Edge detection for gradient boundaries
  - [ ] Image enhancement for better gradients

**Performance Optimizations**

- [ ] Caching strategies
  - [ ] Image tile caching
  - [ ] Gradient computation caching
  - [ ] Incremental image updates
- [ ] Memory management
  - [ ] Efficient image memory access
  - [ ] Streaming image processing
  - [ ] Reduced memory footprint

**Dependencies**

- Image loading and processing utilities
- Interpolator system
- Sampling and filtering algorithms
- Memory management systems

#### agg_span_gouraud.h - Base Gouraud shading implementation ✅ **COMPLETED**

**Gouraud Shading Foundation**

- [x] span_gouraud → SpanGouraud[C] struct - Base Gouraud shading system
  - [x] ColorT template parameter → Output color type
  - [x] Triangle-based color interpolation
  - [x] Smooth color transitions across surfaces
  - [x] 3D-style shading in 2D rendering

**Triangle Vertex System**

- [x] Vertex definition
  - [x] 3-point triangle specification
  - [x] Per-vertex color assignment
  - [x] Coordinate and color validation
  - [x] Triangle geometry calculations
- [x] Vertex color interpolation
  - [x] Barycentric coordinate calculation
  - [x] Smooth color blending across triangle
  - [x] Edge color interpolation
  - [x] Color space preservation

**Coordinate System**

- [x] Coordinate calculations
  - [x] Triangle area calculation
  - [x] Barycentric coordinate computation
  - [x] Edge coefficient calculation
  - [x] Subpixel precision handling
- [x] Transformation integration
  - [x] Compatible with affine transformations
  - [x] Coordinate system mapping
  - [x] Perspective correction support

**Color Interpolation Mathematics**

- [x] Interpolation algorithms
  - [x] Linear interpolation across triangle surface
  - [x] Perspective-correct interpolation
  - [x] Color gradient calculation
  - [x] Edge case handling (degenerate triangles)
- [x] Precision management
  - [x] Fixed-point vs. floating-point arithmetic
  - [x] Numerical stability in interpolation
  - [x] Color quantization handling

**Triangle Setup and Management**

- [x] Triangle configuration
  - [x] Vertex ordering (clockwise/counterclockwise)
  - [x] Triangle validity checking
  - [x] Degenerate triangle handling
  - [x] Multiple triangle support
- [x] Performance optimization
  - [x] Triangle setup caching
  - [x] Incremental interpolation
  - [x] SIMD optimization opportunities

**Dependencies**

- Color type system
- Mathematical utilities
- Coordinate transformation system
- Interpolation algorithms

#### agg_span_gouraud_gray.h - Grayscale Gouraud shading ✅ **COMPLETED**

**Grayscale Gouraud System**

- [x] span_gouraud_gray → SpanGouraudGray struct - Optimized grayscale Gouraud shading
  - [x] Single-channel (grayscale) output
  - [x] Optimized interpolation for grayscale values
  - [x] Memory and performance benefits over full-color version
  - [x] 8-bit grayscale output (0-255 range)

**Grayscale-Specific Optimizations**

- [ ] Single channel processing
  - [ ] Reduced memory bandwidth (1/3 of RGB)
  - [ ] Faster interpolation calculations
  - [ ] Simplified color blending
  - [ ] Cache-friendly memory access patterns
- [ ] Grayscale interpolation
  - [ ] Linear grayscale value interpolation
  - [ ] Gamma-correct grayscale blending
  - [ ] Perceptual grayscale handling

**Triangle Vertex Configuration**

- [ ] Grayscale vertex setup
  - [ ] Single grayscale value per vertex
  - [ ] Coordinate system identical to color version
  - [ ] Triangle geometry calculations reused
  - [ ] Simplified vertex structure

**Specialized Use Cases**

- [ ] Monochrome rendering
  - [ ] Black and white artistic effects
  - [ ] Lighting and shadow simulation
  - [ ] Height field visualization
  - [ ] Scientific data visualization
- [ ] Performance-critical applications
  - [ ] Real-time grayscale shading
  - [ ] Low-memory device rendering
  - [ ] High-throughput image processing

**Integration Points**

- [ ] Grayscale rendering pipeline
  - [ ] Compatible with grayscale pixel formats
  - [ ] Grayscale compositing operations
  - [ ] Conversion to color when needed
  - [ ] Anti-aliasing support

**Dependencies**

- Base Gouraud shading system
- Grayscale color type definitions
- Optimized interpolation algorithms

#### agg_span_gouraud_rgba.h - RGBA Gouraud shading ✅ **COMPLETED**

**RGBA Gouraud System**

- [x] span_gouraud_rgba → SpanGouraudRGBA struct - Full-color RGBA Gouraud shading
  - [x] 4-channel RGBA color interpolation
  - [x] Independent alpha channel handling
  - [x] Full-color triangle rendering
  - [x] Premium quality color blending

**RGBA Color Interpolation**

- [ ] Multi-channel interpolation
  - [ ] Independent RGB channel interpolation
  - [ ] Alpha channel interpolation
  - [ ] Premultiplied alpha support
  - [ ] Color space preservation across interpolation
- [ ] Advanced color handling
  - [ ] Gamma-correct color interpolation
  - [ ] Linear vs. sRGB color space handling
  - [ ] HDR color support considerations
  - [ ] Color precision management

**Alpha Channel Features**

- [ ] Alpha interpolation
  - [ ] Smooth alpha transitions across triangle
  - [ ] Alpha premultiplication handling
  - [ ] Alpha blending preparation
  - [ ] Transparency gradient effects
- [ ] Alpha compositing
  - [ ] Compatible with alpha blending systems
  - [ ] Premultiplied alpha optimizations
  - [ ] Alpha test support

**Performance Characteristics**

- [ ] Multi-channel optimization
  - [ ] SIMD optimization opportunities (4 channels)
  - [ ] Memory access pattern optimization
  - [ ] Cache-friendly channel processing
  - [ ] Parallel channel interpolation
- [ ] Quality vs. performance trade-offs
  - [ ] High-precision vs. fast interpolation modes
  - [ ] Quality level configuration
  - [ ] Adaptive precision based on triangle size

**Advanced Shading Features**

- [ ] Complex shading effects
  - [ ] Multi-light simulation
  - [ ] Material property interpolation
  - [ ] Texture coordinate interpolation preparation
  - [ ] Advanced lighting model support
- [ ] Integration capabilities
  - [ ] Compatible with texture mapping systems
  - [ ] Normal map support preparation
  - [ ] Multi-pass rendering support

**Dependencies**

- Base Gouraud shading system
- RGBA color type definitions
- Advanced interpolation algorithms
- Alpha handling utilities

---

### Image Processing

#### [x] agg_image_accessors.h - Image pixel data access with boundary handling

**Core Image Accessor System**

- [x] image_accessor_clip → ImageAccessorClip[PixFmt] struct - Bounds-checked pixel access with background color
  - [x] PixFmt template parameter → Generic pixel format support (RGBA8, RGB565, etc.)
  - [x] Boundary clipping with configurable background color
  - [x] Safe access for coordinates outside image bounds
  - [x] Efficient span-based pixel reading with bounds checking
        **Pixel Access Methods**
- [x] Construction and attachment
  - [x] Default constructor → NewImageAccessorClip() - empty accessor
  - [x] Constructor with background → NewImageAccessorClipWithBackground() - set background color
  - [x] attach(pixfmt) → Attach() - connect to pixel format buffer
  - [x] background_color(color) → SetBackgroundColor() - update background fill color
- [x] Pixel reading interface
  - [x] span(x, y, len) → Span() - read horizontal pixel span with bounds checking
  - [x] next_x() → NextX() - advance to next pixel in current span
  - [x] next_y() → NextY() - advance to next scanline, reset X position
  - [x] Automatic background pixel return for out-of-bounds access
        **No-Clip Accessor (Performance Optimized)**
- [x] image_accessor_no_clip → ImageAccessorNoClip[PixFmt] struct - Fast unchecked pixel access
  - [x] No boundary checking for maximum performance
  - [x] Direct pixel buffer access without safety overhead
  - [x] span(x, y, len) → Span() - direct span access (caller must ensure bounds)
  - [x] next_x() → NextX() - fast pixel advancement
  - [x] next_y() → NextY() - fast scanline advancement
        **Clone Edge Accessor**
- [x] image_accessor_clone → ImageAccessorClone[PixFmt] struct - Edge pixel replication
  - [x] Out-of-bounds coordinates clamp to nearest edge pixel
  - [x] No background color needed - always returns valid image data
  - [x] Seamless texture extension behavior
  - [x] span/next_x/next_y interface with edge clamping logic
        **Wrap Mode Accessors**
- [x] image_accessor_wrap → ImageAccessorWrap[PixFmt, WrapX, WrapY] struct - Tiling/wrapping modes
  - [x] WrapX/WrapY template parameters → Coordinate wrapping strategies
  - [x] Automatic tiling for seamless pattern repetition
  - [x] Multiple wrap mode support (repeat, reflect, etc.)
  - [x] High-performance coordinate transformation
        **Coordinate Wrapping Modes**
- [x] wrap_mode_repeat → WrapModeRepeat struct - Standard tiling repetition
  - [x] Modulo-based coordinate wrapping
  - [x] operator()(int v) → Apply() - wrap coordinate to valid range
  - [x] operator++() → Next() - increment with automatic wrapping
  - [x] Efficient handling of large coordinate values
- [x] wrap_mode_repeat_pow2 → WrapModeRepeatPow2 struct - Power-of-2 optimized repetition
  - [x] Bit-mask based wrapping for power-of-2 image dimensions
  - [x] Single bitwise AND operation for coordinate wrapping
  - [x] Significant performance improvement for 2^N sized images
- [x] wrap_mode_repeat_auto_pow2 → WrapModeRepeatAutoPow2 struct - Adaptive repetition
  - [x] Automatic selection between mask and modulo based on image size
  - [x] Optimal performance regardless of image dimensions
  - [x] Transparent fallback between optimization strategies
- [x] wrap_mode_reflect → WrapModeReflect struct - Mirror repetition
  - [x] Ping-pong coordinate reflection at boundaries
  - [x] Seamless mirrored tiling without edge artifacts
  - [x] Symmetric pattern generation
- [x] wrap_mode_reflect_pow2 → WrapModeReflectPow2 struct - Optimized mirror for power-of-2
  - [x] Bit-mask based reflection for 2^N dimensions
  - [x] Fast reflection using bitwise operations
- [x] wrap_mode_reflect_auto_pow2 → WrapModeReflectAutoPow2 struct - Adaptive mirror
  - [x] Automatic optimization selection for reflection
  - [x] Performance optimization regardless of image size
        **Dependencies**
- [x] agg_basics.h → internal/basics package (coordinate and color types)
- [x] Pixel format system compatibility (internal/pixfmt/)
- [x] Color type system for background colors
- [x] Efficient span-based pixel access patterns

#### agg_image_filters.h - Image filtering kernel functions and lookup tables

**Filter Scale Constants**

- [ ] image_filter_scale_e → ImageFilterScale enumeration - Filter precision constants
  - [ ] image_filter_shift → ImageFilterShift constant (14 bits) - Fixed-point precision
  - [ ] image_filter_scale → ImageFilterScale constant (16384) - Unity filter value
  - [ ] image_filter_mask → ImageFilterMask constant (16383) - Value masking
- [ ] image_subpixel_scale_e → ImageSubpixelScale enumeration - Subpixel precision
  - [ ] image_subpixel_shift → ImageSubpixelShift constant (8 bits) - Subpixel resolution
  - [ ] image_subpixel_scale → ImageSubpixelScale constant (256) - Subpixel unity
  - [ ] image_subpixel_mask → ImageSubpixelMask constant (255) - Subpixel masking
        **Filter Lookup Table System**
- [ ] image_filter_lut → ImageFilterLUT struct - Pre-computed filter weight lookup table
  - [ ] Template-based filter calculation → Calculate[FilterF]() - generate LUT from filter function
  - [ ] Normalization support for weight sum conservation
  - [ ] radius() → Radius() - filter kernel radius
  - [ ] diameter() → Diameter() - full filter width
  - [ ] start() → Start() - starting offset for filter weights
  - [ ] weight_array() → WeightArray() - access to pre-computed weights
- [ ] Filter weight computation
  - [ ] Subpixel precision weight calculation
  - [ ] Symmetric weight distribution around center
  - [ ] Automatic normalization to preserve image brightness
  - [ ] Memory-efficient weight storage
        **Generic Filter Template**
- [ ] image_filter → ImageFilter[FilterF] struct - Template wrapper for filter functions
  - [ ] FilterF template parameter → Any filter function with radius() and calc_weight() methods
  - [ ] Automatic LUT generation from filter function
  - [ ] Single-line filter instantiation
        **Standard Filter Implementations**
- [ ] image_filter_bilinear → ImageFilterBilinear struct - Linear interpolation filter
  - [ ] radius() → 1.0 - single pixel radius
  - [ ] calc_weight(x) → CalcWeight() - linear weight calculation (1.0 - x)
  - [ ] Fastest quality filter with minimal blur
- [ ] image_filter_hanning → ImageFilterHanning struct - Hanning window filter
  - [ ] radius() → 1.0 - single pixel radius
  - [ ] calc_weight(x) → CalcWeight() - cosine-based weight (0.5 + 0.5*cos(π*x))
  - [ ] Smooth transition with reduced ringing
- [ ] image_filter_hamming → ImageFilterHamming struct - Hamming window filter
  - [ ] radius() → 1.0 - single pixel radius
  - [ ] calc_weight(x) → CalcWeight() - modified cosine (0.54 + 0.46*cos(π*x))
  - [ ] Better frequency response than Hanning
- [ ] image_filter_hermite → ImageFilterHermite struct - Hermite cubic filter
  - [ ] radius() → 1.0 - single pixel radius
  - [ ] calc_weight(x) → CalcWeight() - cubic polynomial ((2*x-3)*x\*x + 1)
  - [ ] Smooth cubic interpolation
        **Higher-Order Filters**
- [ ] image_filter_quadric → ImageFilterQuadric struct - Quadratic B-spline
  - [ ] radius() → 1.5 - extended support
  - [ ] calc_weight(x) → CalcWeight() - piecewise quadratic function
  - [ ] Good balance of quality and performance
- [ ] image_filter_bicubic → ImageFilterBicubic struct - Bicubic interpolation
  - [ ] radius() → 2.0 - two-pixel support
  - [ ] calc_weight(x) → CalcWeight() - cubic B-spline weights
  - [ ] High-quality smooth interpolation
- [ ] image_filter_catrom → ImageFilterCatrom struct - Catmull-Rom cubic
  - [ ] radius() → 2.0 - two-pixel support
  - [ ] calc_weight(x) → CalcWeight() - Catmull-Rom polynomial
  - [ ] Sharp, contrast-preserving interpolation
- [ ] image_filter_mitchell → ImageFilterMitchell struct - Mitchell-Netravali filter
  - [ ] radius() → 2.0 - two-pixel support
  - [ ] Configurable B and C parameters (default 1/3, 1/3)
  - [ ] calc_weight(x) → CalcWeight() - parameterized cubic
  - [ ] Tunable balance between sharpness and ringing
        **Advanced Mathematical Filters**
- [ ] image_filter_spline16 → ImageFilterSpline16 struct - 16-sample spline
  - [ ] radius() → 2.0 - two-pixel support
  - [ ] calc_weight(x) → CalcWeight() - optimized spline function
  - [ ] High quality with minimal computational cost
- [ ] image_filter_spline36 → ImageFilterSpline36 struct - 36-sample spline
  - [ ] radius() → 3.0 - three-pixel support
  - [ ] calc_weight(x) → CalcWeight() - extended spline function
  - [ ] Maximum quality spline interpolation
- [ ] image_filter_gaussian → ImageFilterGaussian struct - Gaussian blur filter
  - [ ] radius() → 2.0 - two-pixel support
  - [ ] calc_weight(x) → CalcWeight() - exp(-2*x²)*sqrt(2/π)
  - [ ] Natural blur with no ringing artifacts
- [ ] image_filter_kaiser → ImageFilterKaiser struct - Kaiser window filter
  - [ ] Configurable β parameter (default 6.33)
  - [ ] radius() → 1.0 - single pixel radius
  - [ ] calc_weight(x) → CalcWeight() - Bessel function based
  - [ ] Optimal frequency domain characteristics
- [ ] image_filter_bessel → ImageFilterBessel struct - Bessel function filter
  - [ ] radius() → 3.2383 - optimal support radius
  - [ ] calc_weight(x) → CalcWeight() - first-order Bessel function
  - [ ] Minimal ringing with sharp transitions
        **Windowed Sinc Filters**
- [ ] image_filter_sinc → ImageFilterSinc struct - Windowed sinc filter
  - [ ] Configurable radius (minimum 2.0)
  - [ ] calc_weight(x) → CalcWeight() - sin(πx)/(πx)
  - [ ] Theoretical ideal reconstruction filter
- [ ] image_filter_lanczos → ImageFilterLanczos struct - Lanczos filter
  - [ ] Configurable radius (minimum 2.0)
  - [ ] calc_weight(x) → CalcWeight() - windowed sinc with sinc window
  - [ ] Industry-standard high-quality resampling
        **Dependencies**
- [ ] agg_array.h → internal/array package (pod_array for weight storage)
- [ ] agg_math.h → internal/math package (mathematical functions, π constant)
- [ ] Mathematical utilities (trigonometric functions, Bessel functions)
- [ ] Fixed-point arithmetic support

#### agg_span_image_filter.h - Base classes for image filtering span generators

**Base Image Filter Span**

- [ ] span_image_filter → SpanImageFilter[Source, Interpolator] struct - Foundation for filtered image spans
  - [ ] Source template parameter → Image source type (pixel accessor)
  - [ ] Interpolator template parameter → Coordinate transformation
  - [ ] image_filter_lut integration for filter weight lookup
  - [ ] Subpixel positioning with configurable offset
        **Filter Configuration**
- [ ] Construction and setup
  - [ ] Constructor with source, interpolator, and filter LUT
  - [ ] attach(source) → Attach() - connect to image source
  - [ ] filter(image_filter_lut) → SetFilter() - assign filter weights
  - [ ] interpolator(interpolator) → SetInterpolator() - assign coordinate transform
- [ ] Filter offset control
  - [ ] filter_offset(dx, dy) → SetFilterOffset() - subpixel positioning
  - [ ] filter_offset(d) → SetFilterOffset() - uniform X/Y offset
  - [ ] Dual precision: floating-point and integer offset storage
  - [ ] Automatic subpixel scale conversion
        **Filter Property Access**
- [ ] source() → Source() - access underlying image source
- [ ] filter() → Filter() - access filter weight lookup table
- [ ] interpolator() → Interpolator() - access coordinate transformation
- [ ] filter_dx_int/filter_dy_int() → FilterDxInt/FilterDyInt() - integer offsets
- [ ] filter_dx_dbl/filter_dy_dbl() → FilterDxDbl/FilterDyDbl() - floating offsets
- [ ] prepare() → Prepare() - setup for span generation (base no-op)
      **Affine Resampling Span**
- [ ] span_image_resample_affine → SpanImageResampleAffine[Source] struct - Optimized affine transformation resampling
  - [ ] Built-in span_interpolator_linear[trans_affine] interpolator
  - [ ] Automatic scaling analysis and blur control
  - [ ] Scale limitation to prevent excessive memory usage
  - [ ] Separate X/Y blur factor support
        **Affine Scaling Control**
- [ ] Scale analysis and limits
  - [ ] scale_limit() → ScaleLimit() - maximum allowed scaling factor
  - [ ] Automatic scale detection from affine transformation
  - [ ] Scale limiting to prevent performance degradation
  - [ ] Combined X/Y scale factor management
- [ ] Blur control
  - [ ] blur_x()/blur_y() → BlurX/BlurY() - independent axis blur factors
  - [ ] blur(v) → SetBlur() - uniform blur setting
  - [ ] Blur integration with computed scaling factors
  - [ ] Quality vs. performance trade-off control
        **Advanced Scaling Computation**
- [ ] prepare() → Prepare() - pre-compute scaling parameters
  - [ ] transformer.scaling_abs() analysis for X/Y scale factors
  - [ ] Scale limiting with preservation of aspect ratios
  - [ ] Blur factor integration into final scaling
  - [ ] Subpixel precision scale and inverse scale computation
- [ ] Internal scaling parameters
  - [ ] m_rx/m_ry → RX/RY - computed scale factors in subpixel units
  - [ ] m_rx_inv/m_ry_inv → RXInv/RYInv - inverse scales for efficient computation
  - [ ] Automatic subpixel precision conversion
        **Generic Resampling Span**
- [ ] span_image_resample → SpanImageResample[Source, Interpolator] struct - General resampling with any interpolator
  - [ ] Generic interpolator support (perspective, bilinear, etc.)
  - [ ] Integer-based scale limiting (different from affine version)
  - [ ] Subpixel blur factors for fine quality control
  - [ ] Compatible with all interpolator types
        **Dependencies**
- [ ] agg_basics.h → internal/basics package
- [ ] agg_image_filters.h → filter weight lookup tables (ImageFilterLUT)
- [ ] agg_span_interpolator_linear.h → linear interpolation support
- [ ] Interpolator system compatibility
- [ ] Image source/accessor system integration

#### agg_span_image_filter_gray.h - Specialized grayscale image filtering spans

**Nearest Neighbor Grayscale**

- [ ] span_image_filter_gray_nn → SpanImageFilterGrayNN[Source, Interpolator] struct - Fast grayscale nearest neighbor
  - [ ] No filtering - direct pixel selection
  - [ ] color_type → Grayscale color type (Gray8, Gray16, etc.)
  - [ ] Subpixel coordinate truncation to integer pixels
  - [ ] Alpha channel set to full opacity automatically
        **NN Generation Method**
- [ ] generate(span, x, y, len) → Generate() - fill span with nearest neighbor pixels
  - [ ] interpolator.begin() → coordinate sequence initialization
  - [ ] interpolator.coordinates() → subpixel coordinate retrieval
  - [ ] image_subpixel_shift → coordinate conversion to integer pixels
  - [ ] Direct pixel value assignment with full alpha
        **Bilinear Grayscale Filtering**
- [ ] span_image_filter_gray_bilinear → SpanImageFilterGrayBilinear[Source, Interpolator] struct - Bilinear grayscale interpolation
  - [ ] 2x2 pixel sampling for smooth interpolation
  - [ ] Fixed-point arithmetic for performance
  - [ ] calc_type → calculation type for intermediate values
  - [ ] long_type → extended precision for accumulation
        **Bilinear Generation Algorithm**
- [ ] generate(span, x, y, len) → Generate() - bilinear filtered span generation
  - [ ] Four-pixel sampling (top-left, top-right, bottom-left, bottom-right)
  - [ ] Subpixel weight calculation for X and Y directions
  - [ ] Weighted average computation in fixed-point arithmetic
  - [ ] Integer to color value type conversion
        **Advanced Grayscale Filtering**
- [ ] span_image_filter_gray → SpanImageFilterGray[Source, Interpolator] struct - Full kernel grayscale filtering
  - [ ] Complete filter LUT support (any filter type)
  - [ ] Variable kernel size based on filter radius
  - [ ] Support for all standard filters (bicubic, Lanczos, etc.)
  - [ ] Optimized grayscale-specific processing
        **Advanced Generation Process**
- [ ] generate(span, x, y, len) → Generate() - full kernel filtering
  - [ ] Filter radius determination from LUT
  - [ ] Multi-pixel sampling based on kernel size
  - [ ] Weight lookup from pre-computed LUT
  - [ ] Weighted sum accumulation with proper normalization
  - [ ] High-precision intermediate calculations
        **Grayscale-Specific Optimizations**
- [ ] Single-channel processing optimization
  - [ ] No color component separation needed
  - [ ] Simplified weight application (single multiply per pixel)
  - [ ] Reduced memory bandwidth (1 byte vs 3-4 bytes per pixel)
  - [ ] Cache-friendly access patterns
- [ ] Precision handling
  - [ ] value_type → input pixel precision (8-bit, 16-bit)
  - [ ] calc_type → calculation precision (typically wider)
  - [ ] Proper range clamping and overflow prevention
  - [ ] Efficient conversion between precision levels
        **Dependencies**
- [ ] agg_basics.h → internal/basics package
- [ ] agg_color_gray.h → grayscale color types (Gray8, Gray16)
- [ ] agg_span_image_filter.h → base image filter span functionality
- [ ] Fixed-point arithmetic utilities
- [ ] Subpixel coordinate system constants

#### agg_span_image_filter_rgb.h - RGB color image filtering spans

**Nearest Neighbor RGB**

- [ ] span_image_filter_rgb_nn → SpanImageFilterRGBNN[Source, Interpolator] struct - Fast RGB nearest neighbor
  - [ ] No filtering - direct RGB pixel selection
  - [ ] Three-channel color processing (R, G, B)
  - [ ] Subpixel coordinate truncation to integer pixels
  - [ ] Alpha channel handling (if present in source)
        **RGB NN Generation**
- [ ] generate(span, x, y, len) → Generate() - RGB nearest neighbor span generation
  - [ ] coordinate retrieval and integer conversion
  - [ ] Direct RGB pixel value assignment
  - [ ] Component-wise value copying (r, g, b)
  - [ ] Optional alpha channel preservation
        **Bilinear RGB Filtering**
- [ ] span_image_filter_rgb_bilinear → SpanImageFilterRGBBilinear[Source, Interpolator] struct - Bilinear RGB interpolation
  - [ ] 2x2 pixel sampling per color channel
  - [ ] Independent R, G, B channel processing
  - [ ] Fixed-point arithmetic for all three channels
  - [ ] Synchronized channel interpolation
        **RGB Bilinear Algorithm**
- [ ] generate(span, x, y, len) → Generate() - RGB bilinear span generation
  - [ ] Four-pixel RGB sampling (2x2 neighborhood)
  - [ ] Per-channel weight calculation and application
  - [ ] Independent R, G, B weighted averages
  - [ ] Parallel channel processing for efficiency
  - [ ] Color component clamping and conversion
        **Advanced RGB Filtering**
- [ ] span_image_filter_rgb → SpanImageFilterRGB[Source, Interpolator] struct - Full kernel RGB filtering
  - [ ] Complete filter LUT support for RGB images
  - [ ] Variable kernel size RGB processing
  - [ ] All standard filters (bicubic, Lanczos, etc.) for RGB
  - [ ] Channel-synchronized filtering
        **Advanced RGB Generation**
- [ ] generate(span, x, y, len) → Generate() - full kernel RGB filtering
  - [ ] Multi-pixel RGB sampling based on filter radius
  - [ ] Per-channel weight lookup and application
  - [ ] Independent R, G, B weighted sum accumulation
  - [ ] Synchronized normalization across channels
  - [ ] High-precision RGB intermediate calculations
        **RGB-Specific Optimizations**
- [ ] Three-channel processing
  - [ ] Vectorized RGB operations where possible
  - [ ] Cache-friendly RGB pixel layout access
  - [ ] Efficient RGB-to-RGB color space operations
  - [ ] Minimal branching in RGB processing loops
- [ ] Color precision handling
  - [ ] RGB value_type → input pixel precision (RGB555, RGB565, RGB888)
  - [ ] RGB calc_type → calculation precision for each channel
  - [ ] Independent channel range clamping
  - [ ] Efficient RGB format conversions
        **Memory and Performance**
- [ ] RGB memory access patterns
  - [ ] Sequential RGB component access
  - [ ] Aligned RGB pixel reads where possible
  - [ ] Prefetching for RGB neighborhoods
  - [ ] Cache-aware RGB span processing
        **Dependencies**
- [ ] agg_basics.h → internal/basics package
- [ ] agg_color_rgba.h → RGB color types (RGB555, RGB565, RGB888)
- [ ] agg_span_image_filter.h → base image filter span functionality
- [ ] Three-channel fixed-point arithmetic
- [ ] RGB pixel format system integration

#### agg_span_image_filter_rgba.h - RGBA image filtering with alpha channel processing

**Nearest Neighbor RGBA**

- [ ] span_image_filter_rgba_nn → SpanImageFilterRGBANN[Source, Interpolator] struct - Fast RGBA nearest neighbor
  - [ ] Four-channel processing (R, G, B, A)
  - [ ] Direct RGBA pixel selection without filtering
  - [ ] Full alpha channel preservation
  - [ ] Subpixel coordinate truncation
        **RGBA NN Generation**
- [ ] generate(span, x, y, len) → Generate() - RGBA nearest neighbor span generation
  - [ ] Complete RGBA pixel value copying
  - [ ] Alpha channel direct transfer
  - [ ] Four-component assignment (r, g, b, a)
  - [ ] Efficient RGBA pixel access
        **Bilinear RGBA Filtering**
- [ ] span_image_filter_rgba_bilinear → SpanImageFilterRGBABilinear[Source, Interpolator] struct - Bilinear RGBA interpolation
  - [ ] 2x2 pixel sampling for all four channels
  - [ ] Independent RGBA channel processing
  - [ ] Alpha-aware interpolation
  - [ ] Synchronized four-channel filtering
        **RGBA Bilinear Algorithm**
- [ ] generate(span, x, y, len) → Generate() - RGBA bilinear span generation
  - [ ] Four-pixel RGBA sampling (2x2 neighborhood)
  - [ ] Individual R, G, B, A weight application
  - [ ] Independent four-channel weighted averages
  - [ ] Alpha channel proper interpolation
  - [ ] Premultiplied alpha handling (if applicable)
        **Advanced RGBA Filtering**
- [ ] span_image_filter_rgba → SpanImageFilterRGBA[Source, Interpolator] struct - Full kernel RGBA filtering
  - [ ] Complete filter LUT support for RGBA images
  - [ ] Variable kernel size four-channel processing
  - [ ] All standard filters with alpha support
  - [ ] Alpha-aware filter weight application
        **Advanced RGBA Generation**
- [ ] generate(span, x, y, len) → Generate() - full kernel RGBA filtering
  - [ ] Multi-pixel RGBA sampling based on filter radius
  - [ ] Four-channel weight lookup and application
  - [ ] Independent R, G, B, A weighted sum accumulation
  - [ ] Alpha-aware normalization and processing
  - [ ] High-precision RGBA intermediate calculations
        **Alpha Channel Handling**
- [ ] Alpha processing modes
  - [ ] Straight alpha processing (independent alpha channel)
  - [ ] Premultiplied alpha support (color values pre-multiplied by alpha)
  - [ ] Alpha blending preparation
  - [ ] Alpha channel precision preservation
- [ ] Alpha-aware filtering
  - [ ] Alpha weight consideration in filtering
  - [ ] Alpha channel edge handling
  - [ ] Transparent pixel boundary processing
  - [ ] Alpha premultiplication/demultiplication as needed
        **RGBA-Specific Optimizations**
- [ ] Four-channel processing
  - [ ] SIMD-friendly RGBA operations where possible
  - [ ] Aligned RGBA pixel access (32-bit/64-bit alignment)
  - [ ] Vectorized RGBA arithmetic
  - [ ] Efficient RGBA span traversal
- [ ] Alpha optimization
  - [ ] Alpha channel early termination for fully transparent pixels
  - [ ] Alpha-aware sampling reduction
  - [ ] Optimized alpha blending preparation
  - [ ] Alpha channel-specific precision handling
        **Memory Layout Considerations**
- [ ] RGBA pixel formats
  - [ ] RGBA8888 (32-bit) optimized access
  - [ ] RGBA16161616 (64-bit) high precision
  - [ ] Component order handling (RGBA vs ABGR, etc.)
  - [ ] Packed pixel format support
- [ ] Memory access patterns
  - [ ] Sequential RGBA component access
  - [ ] Cache-friendly RGBA neighborhood reads
  - [ ] Prefetching for RGBA filtering kernels
  - [ ] Memory bandwidth optimization
        **Dependencies**
- [ ] agg_basics.h → internal/basics package
- [ ] agg_color_rgba.h → RGBA color types (RGBA8, RGBA16)
- [ ] agg_span_image_filter.h → base image filter span functionality
- [ ] Four-channel fixed-point arithmetic
- [ ] Alpha channel processing utilities
- [ ] RGBA pixel format system integration

---

### Pattern Processing

#### agg_pattern_filters_rgba.h - RGBA pattern filters (`internal/span/pattern_filters.go`)

**Pattern Filter System** - Core texture sampling filters for patterns and images

- [ ] pattern_filter_nn → PatternFilterNN[ColorT] struct - Nearest neighbor sampling
  - [ ] ColorT template parameter → Generic color type support (RGBA8, RGBA16, etc.)
  - [ ] dilation() → Dilation() method - returns filter kernel size (0 for NN)
  - [ ] pixel_low_res() → PixelLowRes() - direct pixel access for low-resolution sampling
  - [ ] pixel_high_res() → PixelHighRes() - subpixel-aware sampling with bit shifting
  - [ ] Fast pixel-perfect sampling with no interpolation

**Bilinear Pattern Filter**

- [ ] pattern_filter_bilinear_rgba → PatternFilterBilinearRGBA[ColorT] struct - Smooth bilinear interpolation
  - [ ] ColorT template parameter → Supports RGBA8, RGBA16 color types
  - [ ] value_type and calc_type → Component and calculation precision types
  - [ ] dilation() → Dilation() method - returns 1 (requires 2x2 pixel neighborhood)
  - [ ] pixel_low_res() → PixelLowRes() - fallback to direct sampling for low resolution
  - [ ] pixel_high_res() → PixelHighRes() - full bilinear interpolation with subpixel precision

**Bilinear Interpolation Algorithm**

- [ ] Subpixel coordinate extraction
  - [ ] line_subpixel_shift constant for bit operations
  - [ ] line_subpixel_mask for fractional coordinate extraction
  - [ ] line_subpixel_scale for normalized weights
- [ ] 2x2 neighborhood sampling
  - [ ] Four corner pixel reads from source buffer
  - [ ] Bilinear weight calculation based on fractional coordinates
  - [ ] Separate RGBA component interpolation with proper precision
- [ ] High-precision color arithmetic
  - [ ] calc_type precision for intermediate calculations
  - [ ] Component-wise weighted accumulation
  - [ ] Final normalization and clamping to value_type range

**Filter Integration**

- [ ] Compatible with pattern span generators
  - [ ] Template parameter for span_pattern_rgba and similar
  - [ ] Consistent interface for filter switching
  - [ ] Performance-optimized filter selection
- [ ] Subpixel-aware rendering pipeline
  - [ ] Integration with anti-aliased rendering
  - [ ] Coordinate system compatibility with rasterizer
  - [ ] Memory-efficient buffer access patterns

#### agg_span_pattern_gray.h - Grayscale pattern span generator (`internal/span/span_pattern_gray.go`)

**Grayscale Pattern Span System**

- [ ] span_pattern_gray → SpanPatternGray[Source] struct - Grayscale pattern rendering
  - [ ] Source template parameter → Image source with grayscale pixel access
  - [ ] source_type and color_type type definitions
  - [ ] offset_x and offset_y for pattern positioning
  - [ ] Single-channel grayscale value handling

**Pattern Positioning Control**

- [ ] Pattern offset management
  - [ ] offset_x() → OffsetX() - horizontal pattern offset accessor/mutator
  - [ ] offset_y() → OffsetY() - vertical pattern offset accessor/mutator
  - [ ] Pattern tiling through coordinate transformation
  - [ ] Seamless pattern repetition handling

**Span Generation Interface**

- [ ] SpanGenerator interface implementation
  - [ ] prepare() → Prepare() - pre-rendering setup
  - [ ] generate() → Generate() - fill span with grayscale pattern pixels
  - [ ] alpha() → Alpha() - alpha channel handling (not used for grayscale)
  - [ ] Integration with scanline rendering pipeline

**Grayscale Pattern Rendering**

- [ ] Efficient grayscale pixel processing
  - [ ] Single-component value extraction from source
  - [ ] Memory-efficient grayscale span generation
  - [ ] Compatible with grayscale pixel format system
- [ ] Source coordinate calculation
  - [ ] Pattern coordinate wrapping and bounds handling
  - [ ] next_x() iteration through source pixels
  - [ ] Horizontal span traversal optimization

#### agg_span_pattern_rgb.h - RGB pattern span generator (`internal/span/span_pattern_rgb.go`)

**RGB Pattern Span System**

- [ ] span_pattern_rgb → SpanPatternRGB[Source] struct - RGB pattern rendering
  - [ ] Source template parameter → Image source with RGB pixel access
  - [ ] order_type for RGB component arrangement (RGB vs BGR)
  - [ ] Three-channel RGB color handling

**RGB Component Processing**

- [ ] Component order abstraction
  - [ ] order_type::R, order_type::G, order_type::B constants
  - [ ] Flexible RGB vs BGR pixel format support
  - [ ] Component-wise span filling from source pattern
- [ ] Efficient RGB pixel extraction
  - [ ] Three-component value reads from source buffer
  - [ ] Memory-aligned RGB access where possible
  - [ ] Compatible with RGB pixel format system

**Pattern Rendering Pipeline**

- [ ] RGB span generation process
  - [ ] Coordinate offset application (pattern positioning)
  - [ ] Source span() method integration for horizontal strips
  - [ ] next_x() advancement through RGB pattern data
  - [ ] Length-based span filling with RGB values

#### agg_span_pattern_rgba.h - RGBA pattern span generator (`internal/span/span_pattern_rgba.go`)

**RGBA Pattern Span System**

- [ ] span_pattern_rgba → SpanPatternRGBA[Source] struct - Full-color pattern rendering
  - [ ] Source template parameter → Image source with RGBA pixel access
  - [ ] order_type for RGBA component arrangement (RGBA, BGRA, ARGB, etc.)
  - [ ] Four-channel RGBA color with alpha support

**RGBA Component Processing**

- [ ] Component order abstraction
  - [ ] order_type::R, order_type::G, order_type::B, order_type::A constants
  - [ ] Support for multiple RGBA pixel arrangements
  - [ ] Full alpha channel preservation and handling
- [ ] High-precision color handling
  - [ ] value_type and calc_type precision support
  - [ ] Component-wise extraction with proper precision
  - [ ] Alpha channel integration with blending system

**Advanced Pattern Features**

- [ ] Pattern transformation support
  - [ ] Compatible with pattern filters (bilinear, nearest neighbor)
  - [ ] Subpixel-accurate pattern sampling
  - [ ] High-quality pattern scaling and rotation
- [ ] Memory optimization
  - [ ] Efficient RGBA buffer traversal
  - [ ] Cache-friendly pattern access patterns
  - [ ] Minimal memory allocation during span generation

**Pattern Span Integration**

- [ ] Rendering pipeline compatibility
  - [ ] Works with existing scanline renderer
  - [ ] Compatible with anti-aliasing system
  - [ ] Integration with alpha blending operations
- [ ] Multi-format source support
  - [ ] Compatible with different RGBA pixel formats
  - [ ] Handles various bit depths (8-bit, 16-bit)
  - [ ] Flexible source buffer arrangements

**Error Handling and Edge Cases**

- [ ] Robust pattern processing
  - [ ] Boundary condition handling for pattern edges
  - [ ] Zero-length span handling
  - [ ] Invalid source pattern handling
  - [ ] Graceful degradation for malformed patterns

**Performance Optimization**

- [ ] Efficient span generation
  - [ ] Vectorized color component processing
  - [ ] Memory prefetching for pattern data
  - [ ] Reduced function call overhead in inner loops
  - [ ] Cache-optimized pattern traversal

**Dependencies**

- [x] Color system (`internal/basics/` - color types and precision) ✅ **COMPLETED**
- [x] Span interpolation system (`internal/span/interpolator_linear.go`) ✅ **COMPLETED**
- [ ] Image accessor system (`internal/image/image_accessors.go`) - **REQUIRED**
  - [ ] Pixel-level access with boundary handling
  - [ ] Multiple pixel format support
  - [ ] Memory-safe buffer access
- [ ] Pixel format system (`internal/pixfmt/`) - **REQUIRED**
  - [ ] Component order definitions (RGB, BGR, RGBA, BGRA, etc.)
  - [ ] Value type and calculation type definitions
  - [ ] Pixel format compatibility layer
- [ ] Scanline rendering system integration
  - [x] Scanline storage and management ✅ **COMPLETED**
  - [x] Span-based rendering pipeline ✅ **COMPLETED**
- [ ] Pattern filter integration
  - [ ] Filter template parameter support
  - [ ] Nearest neighbor and bilinear filter compatibility
  - [ ] High-quality texture sampling

---

### Interpolators

#### agg_span_interpolator_linear.h - Linear Span Interpolator ⚠️ **PARTIAL**

**Linear Interpolation System**

- [x] span_interpolator_linear → SpanInterpolatorLinear[T] struct - Linear span interpolation with affine transformation
  - [x] Transformer template parameter → Generic transformer interface support
  - [x] Configurable subpixel precision (default 8-bit shift = 256x)
  - [x] DDA-based linear interpolation for efficient span generation
  - [x] Integration with affine transformation matrix

**Core Interpolation Interface**

- [x] Construction and configuration
  - [x] Default constructor → NewSpanInterpolatorLinear() - empty interpolator
  - [x] Constructor with transformer → NewSpanInterpolatorLinearWithTransformer() - attach transformer
  - [x] transformer(trans) → SetTransformer() - attach transformation matrix
  - [x] Transformer() → GetTransformer() - access current transformer

**Span Generation**

- [x] Linear interpolation methods
  - [x] begin(x, y, len) → Begin() - start interpolation for horizontal span
  - [x] resynchronize(xe, ye, len) → Resynchronize() - adjust end point for accuracy correction
  - [x] coordinates(x, y) → Coordinates() - get current transformed coordinates
  - [x] operator++() → Next() - advance to next pixel position

**Subpixel Precision**

- [x] High-precision coordinate handling
  - [x] Configurable subpixel shift (4-16 bits typical)
  - [x] Fixed-point arithmetic for performance
  - [x] Subpixel-accurate coordinate interpolation
  - [x] Minimal rounding error accumulation

**Performance Optimization**

- [x] Efficient span processing
  - [x] DDA2 line interpolator for linear coordinate advancement
  - [x] Minimal floating-point operations during span generation
  - [x] Cache-friendly sequential access patterns
  - [x] Optimized for texture mapping and gradient generation

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] agg_dda_line.h → DDA line interpolation algorithms
- [x] agg_trans_affine.h → internal/transform/affine.go
- [x] Transformer interface definitions

#### agg_span_interpolator_persp.h - Perspective Span Interpolator ✅ **COMPLETED**

**Perspective Interpolation System**

- [x] span_interpolator_persp_exact → SpanInterpolatorPerspectiveExact struct - Exact perspective interpolation (`internal/span/interpolator_persp.go`)

  - [x] Uses trans_perspective for precise quadrilateral transformations
  - [x] Handles arbitrary quadrangle-to-quadrangle mapping
  - [x] High-precision perspective-correct interpolation
  - [x] Subdivision-based accuracy for long spans

- [x] span_interpolator_persp_lerp → SpanInterpolatorPerspectiveLerp struct - Linear approximation perspective interpolation (`internal/span/interpolator_persp.go`)
  - [x] Fast perspective interpolation using linear approximation
  - [x] Configurable subdivision for balance between speed and accuracy
  - [x] Suitable for moderate perspective distortion

**Quadrilateral Transformations**

- [ ] Flexible mapping support
  - [ ] quad_to_quad(src, dst) → QuadToQuad() - arbitrary quadrangle mapping
  - [ ] rect_to_quad(x1, y1, x2, y2, quad) → RectToQuad() - rectangle to quadrangle
  - [ ] quad_to_rect(quad, x1, y1, x2, y2) → QuadToRect() - quadrangle to rectangle
  - [ ] Direct and inverse transformation support

**Perspective Coordinate Generation**

- [ ] High-accuracy interpolation
  - [ ] begin(x, y, len) → Begin() - initialize perspective span
  - [ ] coordinates(x, y) → Coordinates() - get perspective-corrected coordinates
  - [ ] operator++() → Next() - advance with perspective correction
  - [ ] Automatic subdivision for accuracy maintenance

**Subdivision Control**

- [ ] Adaptive accuracy management
  - [ ] Configurable subdivision threshold for accuracy vs. performance
  - [ ] Automatic subdivision when perspective distortion is high
  - [ ] Linear interpolation within subdivided segments
  - [ ] Quality control for texture mapping applications

**Advanced Perspective Features**

- [ ] Complex projection support
  - [ ] Perspective projection with vanishing points
  - [ ] 3D-to-2D projection coordinate generation
  - [ ] Lens distortion correction preparation
  - [ ] Camera transformation integration

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] agg_trans_perspective.h → internal/transform/perspective.go
- [x] agg_dda_line.h → DDA line interpolation algorithms
- [ ] Subdivision algorithms for accuracy control

#### agg_span_interpolator_trans.h - Transform Span Interpolator

**Generic Transform Interpolation System**

- [ ] span_interpolator_trans → SpanInterpolatorTransform[T] struct - Generic transformer-based interpolation
  - [ ] Transformer template parameter → Any transformation class support
  - [ ] Per-pixel transformation for maximum flexibility
  - [ ] Configurable subpixel precision
  - [ ] Direct transformation call per coordinate

**Universal Transformer Support**

- [ ] Flexible transformation interface
  - [ ] Works with any transformer implementing Transform() method
  - [ ] Supports non-linear transformations (fish-eye, barrel distortion, etc.)
  - [ ] Real-time coordinate transformation during span generation
  - [ ] No linear interpolation assumptions

**Per-Pixel Accuracy**

- [ ] Exact transformation
  - [ ] begin(x, y, len) → Begin() - initialize with base coordinates
  - [ ] coordinates(x, y) → Coordinates() - get precisely transformed coordinates
  - [ ] operator++() → Next() - advance source coordinates, transform each pixel
  - [ ] No interpolation error accumulation

**Performance Considerations**

- [ ] Transformation overhead management
  - [ ] Direct transform() call per pixel (can be expensive)
  - [ ] Suitable for complex non-linear transformations
  - [ ] May be slower than linear interpolation for simple affine transforms
  - [ ] Best for high-quality arbitrary transformations

**Advanced Transform Applications**

- [ ] Non-linear transformation support
  - [ ] Fish-eye lens correction
  - [ ] Barrel and pincushion distortion
  - [ ] Custom mathematical transformations
  - [ ] Real-time procedural distortions

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] Generic transformer interface (any class with Transform() method)
- [x] Subpixel coordinate utilities

#### agg_span_interpolator_adaptor.h - Interpolator Adaptor

**Distortion Adaptor System**

- [ ] span_interpolator_adaptor → SpanInterpolatorAdaptor[Interpolator, Distortion] struct - Distortion effect wrapper
  - [ ] Interpolator template parameter → Base interpolator (linear, perspective, etc.)
  - [ ] Distortion template parameter → Distortion effect class
  - [ ] Combines coordinate interpolation with distortion effects
  - [ ] Decorator pattern for interpolator enhancement

**Base Interpolator Wrapping**

- [ ] Interpolator inheritance
  - [ ] Inherits from base interpolator class
  - [ ] Preserves all base interpolator functionality
  - [ ] Adds distortion calculation to coordinate output
  - [ ] Transparent interface extension

**Distortion Integration**

- [ ] Distortion effect application
  - [ ] distortion(dist) → SetDistortion() - attach distortion effect
  - [ ] distortion() → GetDistortion() - access current distortion
  - [ ] coordinates(x, y) → Coordinates() - get distorted coordinates
  - [ ] Distortion applied after base interpolation

**Distortion Types**

- [ ] Flexible distortion support
  - [ ] Any class implementing calculate(x, y) method
  - [ ] Real-time coordinate modification
  - [ ] Supports complex mathematical distortions
  - [ ] Chainable distortion effects

**Applications**

- [ ] Advanced visual effects
  - [ ] Lens distortion correction/application
  - [ ] Fish-eye projection effects
  - [ ] Ripple and wave distortions
  - [ ] Custom procedural coordinate modifications

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [ ] Base interpolator implementations (linear, perspective, transform)
- [ ] Distortion effect interface and implementations
- [ ] Mathematical distortion algorithms

#### agg_span_subdiv_adaptor.h - Subdivision Adaptor

**Adaptive Subdivision System**

- [ ] span_subdiv_adaptor → SpanSubdivisionAdaptor[Interpolator] struct - Adaptive subdivision wrapper
  - [ ] Interpolator template parameter → Base interpolator (linear, perspective, etc.)
  - [ ] Configurable subdivision size (power of 2, typically 4-8 bit)
  - [ ] Automatic subdivision for accuracy improvement
  - [ ] Linear interpolation within subdivisions

**Subdivision Control**

- [ ] Adaptive accuracy management
  - [ ] subdiv_shift() → SubdivisionShift() - get current subdivision size (as power of 2)
  - [ ] subdiv_shift(shift) → SetSubdivisionShift() - configure subdivision size
  - [ ] Default 16-pixel subdivision (4-bit shift)
  - [ ] Balance between accuracy and performance

**Accuracy Enhancement**

- [ ] Error reduction strategy
  - [ ] Subdivides long spans into smaller segments
  - [ ] Re-calculates accurate coordinates at subdivision points
  - [ ] Linear interpolation within each subdivision
  - [ ] Prevents error accumulation in long spans

**Performance Optimization**

- [ ] Efficient subdivision processing
  - [ ] begin(x, y, len) → Begin() - initialize with subdivision planning
  - [ ] coordinates(x, y) → Coordinates() - get subdivision-corrected coordinates
  - [ ] operator++() → Next() - advance with subdivision awareness
  - [ ] Minimal overhead for subdivision management

**Advanced Applications**

- [ ] High-quality transformation
  - [ ] Improves accuracy for perspective interpolation
  - [ ] Reduces distortion in long spans
  - [ ] Essential for high-quality texture mapping
  - [ ] Automatic quality vs. performance balancing

**Integration Patterns**

- [ ] Wrapper compatibility
  - [ ] Works with any base interpolator
  - [ ] Transparent interface preservation
  - [ ] Stackable with other adaptors
  - [ ] Configurable subdivision policies

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [ ] Base interpolator implementations
- [ ] Linear interpolation utilities for subdivisions
- [ ] Subdivision coordinate calculation algorithms

---

### Utility and Math

#### agg_alpha_mask_u8.h - 8-bit alpha mask (`internal/pixfmt/alpha_mask.go`)

**Alpha Mask System** - Provides alpha channel masking functionality with configurable mask sources

- [ ] one_component_mask_u8 → OneComponentMaskU8 struct - Single channel alpha extraction
  - [ ] calculate() → Calculate() method - extracts alpha from single byte
  - [ ] Direct 8-bit alpha value reading
  - [ ] Simple mask function for grayscale alpha sources
- [ ] rgb_to_gray_mask_u8 → RGBToGrayMaskU8[R, G, B] struct - RGB to grayscale alpha conversion
  - [ ] R, G, B template parameters → RGB channel offset constants
  - [ ] calculate() → Calculate() method - weighted RGB to grayscale conversion
  - [ ] Standard luminance weights (77, 150, 29) for accurate grayscale conversion
  - [ ] Efficient bit-shift based calculation

**Alpha Mask Class**

- [ ] alpha_mask_u8 → AlphaMaskU8[Step, Offset, MaskF] struct - Configurable alpha mask renderer
  - [ ] Step template parameter → Pixel stride for different formats
  - [ ] Offset template parameter → Alpha channel offset within pixel
  - [ ] MaskF template parameter → Mask function type (OneComponent, RGBToGray, etc.)
  - [ ] cover_type → CoverType type alias for alpha coverage values

**Alpha Mask Operations**

- [ ] Pixel access methods
  - [ ] pixel() → Pixel() method - get alpha value at coordinates with bounds checking
  - [ ] Returns 0 for out-of-bounds coordinates
  - [ ] Applies mask function to source pixel data
  - [ ] Efficient rendering buffer integration
- [ ] Mask attachment and management
  - [ ] attach() → Attach() method - attach to rendering buffer
  - [ ] mask_function() → MaskFunction() accessor for mask function configuration
  - [ ] Cover scale constants (cover_shift=8, cover_none=0, cover_full=255)

**Performance Optimization**

- [ ] Efficient alpha extraction
  - [ ] Template-based compile-time optimization
  - [ ] Minimal memory access for alpha values
  - [ ] Bounds checking with branch-predictable patterns
- [ ] Multi-format alpha support
  - [ ] Configurable pixel stride and offset
  - [ ] Support for RGBA, BGRA, grayscale formats
  - [ ] Custom mask function extensibility

#### agg_bitset_iterator.h - Bitset iterator (`internal/basics/bitset_iterator.go`)

**Bitset Iteration System** - Efficient iteration over set bits in bitmask data structures

- [ ] bitset_iterator → BitsetIterator struct - Iterator for traversing set bits
  - [ ] Bit position tracking and advancement
  - [ ] Efficient bit scanning algorithms
  - [ ] Support for different word sizes (32-bit, 64-bit)
  - [ ] Forward iteration with optimal bit detection

**Bitset Operations**

- [ ] Iterator interface methods
  - [ ] begin() → Begin() method - initialize iterator to first set bit
  - [ ] next() → Next() method - advance to next set bit
  - [ ] current() → Current() method - get current bit position
  - [ ] done() → Done() method - check if iteration is complete
- [ ] Bit scanning optimization
  - [ ] Hardware bit scan instructions where available
  - [ ] Lookup table fallback for portable implementation
  - [ ] Word-level skipping for sparse bitsets
  - [ ] Cache-friendly iteration patterns

**Use Cases**

- [ ] Scanline rendering optimization
  - [ ] Sparse pixel coverage iteration
  - [ ] Active scanline detection
  - [ ] Region-of-interest processing
- [ ] Memory-efficient data structure support
  - [ ] Sparse array implementations
  - [ ] Set membership testing
  - [ ] Efficient bit manipulation utilities

#### agg_blur.h - Blur effects (`internal/effects/blur.go`)

**Blur Algorithm System** - High-quality image blurring with various kernel types

- [ ] Gaussian blur implementation
  - [ ] Separable Gaussian kernel generation
  - [ ] Configurable blur radius and sigma parameters
  - [ ] Two-pass horizontal/vertical blur for efficiency
  - [ ] Edge handling strategies (clamp, mirror, wrap)
- [ ] Box blur implementation
  - [ ] Fast uniform kernel blurring
  - [ ] Multiple-pass box blur for Gaussian approximation
  - [ ] Integer arithmetic optimization
  - [ ] Sliding window technique for performance

**Blur Kernel Management**

- [ ] Kernel generation and caching
  - [ ] Dynamic kernel size calculation
  - [ ] Normalization for proper brightness preservation
  - [ ] Precomputed kernel tables for common radii
  - [ ] Memory-efficient kernel storage
- [ ] Multi-channel blur support
  - [ ] RGB and RGBA channel processing
  - [ ] Alpha-aware blurring for transparent images
  - [ ] Separate channel blur parameters
  - [ ] Premultiplied alpha handling

**Performance Optimization**

- [ ] SIMD-optimized blur kernels
  - [ ] Vectorized convolution operations
  - [ ] Cache-friendly memory access patterns
  - [ ] Parallel processing support for large images
- [ ] Memory management
  - [ ] In-place blurring where possible
  - [ ] Temporary buffer reuse
  - [ ] Minimal memory allocation during processing

#### agg_bounding_rect.h - Bounding rectangle calculation (`internal/basics/bounding_rect.go`)

**Bounding Rectangle System** - Efficient bounding box calculations for geometric objects

- [ ] bounding_rect → BoundingRect function - Calculate axis-aligned bounding rectangle
  - [ ] Generic vertex source input support
  - [ ] Efficient min/max coordinate tracking
  - [ ] Handles empty and single-point cases
  - [ ] Returns proper rectangle structure

**Geometric Bounding Calculations**

- [ ] Path bounding rectangle
  - [ ] Vertex sequence traversal for bounds calculation
  - [ ] Handles different vertex command types (move_to, line_to, curves)
  - [ ] Ignores control points for accurate bounds
  - [ ] Coordinate precision handling
- [ ] Transformed bounding rectangle
  - [ ] Apply transformation matrix to bounds calculation
  - [ ] Handles rotation, scaling, and translation
  - [ ] Efficient transformed vertex processing
  - [ ] Maintains precision during transformation

**Edge Case Handling**

- [ ] Robust bounds calculation
  - [ ] Empty path handling (invalid/empty bounds)
  - [ ] Single vertex paths
  - [ ] Infinite or very large coordinates
  - [ ] Numerical precision considerations
- [ ] Performance optimization
  - [ ] Early termination for simple cases
  - [ ] Minimal floating-point operations
  - [ ] Cache-friendly vertex traversal

#### ✅ agg_clip_liang_barsky.h - Liang-Barsky clipping algorithm (`internal/basics/clip_liang_barsky.go`) **COMPLETED**

**Line Clipping System** - Efficient parametric line clipping against rectangular regions

- [x] Clipping flag constants for vertex classification
  - [x] ClippingFlagsX1Clipped, ClippingFlagsX2Clipped for horizontal bounds
  - [x] ClippingFlagsY1Clipped, ClippingFlagsY2Clipped for vertical bounds
  - [x] Combined flags for unified clipping state
- [x] ClippingFlags() → Cyrus-Beck vertex classification
  - [x] 9-region subdivision for efficient clipping
  - [x] Bitwise flag operations for fast region testing
  - [x] Optimized coordinate comparisons

**Liang-Barsky Algorithm Implementation**

- [x] Parametric line clipping
  - [x] Efficient parameter calculation for line segments
  - [x] Precise intersection point computation
  - [x] Handles all edge cases (parallel lines, degenerate cases)
  - [x] Returns clipped segment coordinates
- [x] Integration with rendering pipeline
  - [x] Compatible with anti-aliased line rendering
  - [x] Subpixel-accurate clipping results
  - [x] Minimal coordinate precision loss

#### ✅ agg_dda_line.h - DDA line algorithm (`internal/span/dda_line.go`) **COMPLETED**

**DDA Line Interpolation System** - Digital Differential Analyzer for line rasterization

- [x] GouraudDDAInterpolator struct - DDA line interpolator for Gouraud shading
  - [x] Configurable fraction precision with fractionShift parameter
  - [x] Integer arithmetic for performance and precision
  - [x] Incremental y-value calculation with sub-pixel accuracy
  - [x] Forward and backward stepping support

**DDA Algorithm Implementation**

- [x] Interpolation methods
  - [x] Inc() → increment interpolator position
  - [x] Dec() → decrement interpolator position
  - [x] Add()/Sub() → bulk step operations
  - [x] Y() → current interpolated value with precision
- [x] Fractional arithmetic system
  - [x] Fixed-point arithmetic for sub-pixel precision
  - [x] Configurable precision shift for different use cases
  - [x] Efficient integer-only calculations

#### agg_gamma_functions.h - Gamma correction functions (`internal/pixfmt/gamma_functions.go`)

**Gamma Correction System** - Mathematical gamma correction and color space conversion functions

- [ ] Linear gamma functions
  - [ ] gamma_none → GammaNone struct - identity gamma (no correction)
  - [ ] Direct value pass-through for linear color spaces
  - [ ] Minimal computational overhead
  - [ ] Used for linear color workflows
- [ ] Power gamma functions
  - [ ] gamma_power → GammaPower struct - configurable power curve gamma
  - [ ] Customizable gamma exponent (typically 2.2 for sRGB)
  - [ ] Efficient power function implementation
  - [ ] Forward and inverse gamma transformations

**Gamma Curve Implementation**

- [ ] Standard gamma curves
  - [ ] sRGB gamma curve with linear segment
  - [ ] Adobe RGB gamma curve implementation
  - [ ] Custom gamma curve support with configurable parameters
- [ ] Lookup table generation
  - [ ] Precomputed gamma tables for performance
  - [ ] Configurable table resolution (8-bit, 16-bit)
  - [ ] Memory-efficient table storage
  - [ ] Runtime gamma curve switching

**Color Space Integration**

- [ ] Pixel format gamma support
  - [ ] Integration with RGB and RGBA pixel formats
  - [ ] Per-channel gamma correction
  - [ ] Alpha channel handling (linear vs. gamma-corrected)
- [ ] Rendering pipeline integration
  - [ ] Gamma-aware blending operations
  - [ ] Linear interpolation in gamma-corrected space
  - [ ] Color accuracy for photo-realistic rendering

#### agg_gamma_lut.h - Gamma lookup table (`internal/pixfmt/gamma_lut.go`)

**Gamma Lookup Table System** - High-performance gamma correction using precomputed tables

- [ ] gamma_lut → GammaLUT[GammaF] struct - Gamma lookup table with configurable function
  - [ ] GammaF template parameter → Gamma function type (power, sRGB, etc.)
  - [ ] Precomputed lookup tables for fast gamma correction
  - [ ] Configurable table size (256, 1024, 4096 entries)
  - [ ] Forward and inverse gamma tables

**Lookup Table Management**

- [ ] Table generation and initialization
  - [ ] Dynamic table generation from gamma function
  - [ ] gamma() → Gamma() accessor for gamma function configuration
  - [ ] Automatic table rebuilding when gamma parameters change
  - [ ] Memory-efficient table storage
- [ ] Fast gamma correction methods
  - [ ] operator() → Direct lookup for 8-bit values
  - [ ] Interpolated lookup for higher precision inputs
  - [ ] Batch gamma correction for spans of pixels
  - [ ] SIMD-optimized table lookups where possible

**Performance Optimization**

- [ ] Cache-efficient lookups
  - [ ] Compact table layout for cache friendliness
  - [ ] Prefetching for predictable access patterns
  - [ ] Branch-free table index calculation
- [ ] Multi-precision support
  - [ ] 8-bit and 16-bit input/output support
  - [ ] Configurable interpolation for smooth gradients
  - [ ] Precision vs. speed trade-offs

#### agg_gradient_lut.h - Gradient lookup table (`internal/span/gradient_lut.go`)

**Gradient Lookup Table System** - Efficient gradient color interpolation using precomputed tables

- [ ] gradient_lut → GradientLUT struct - Gradient color lookup table
  - [ ] Configurable gradient size (256, 512, 1024 color stops)
  - [ ] Linear and smooth gradient interpolation
  - [ ] Multiple color format support (RGB, RGBA, grayscale)
  - [ ] Efficient color array storage

**Gradient Generation and Management**

- [ ] Color stop management
  - [ ] build_lut() → BuildLUT() method - build lookup table from color stops
  - [ ] Multiple color stops with position and color
  - [ ] Automatic interpolation between color stops
  - [ ] Support for alpha gradients and transparency
- [ ] Gradient access methods
  - [ ] operator[] → Color lookup by gradient position
  - [ ] Fast array-based color access
  - [ ] Wrap and clamp modes for gradient edges
  - [ ] Smooth color interpolation

**Advanced Gradient Features**

- [ ] Gradient transformation
  - [ ] Linear gradient along arbitrary vectors
  - [ ] Radial gradient with configurable center and radius
  - [ ] Conical/angular gradients
  - [ ] Complex gradient transformations
- [ ] Performance optimization
  - [ ] Precomputed color interpolation
  - [ ] SIMD-optimized gradient evaluation
  - [ ] Cache-friendly gradient table layout
  - [ ] Minimal per-pixel gradient calculations

#### ✅ agg_line_aa_basics.h - Anti-aliased line basics (`internal/primitives/line_aa_basics.go`) **COMPLETED**

**Anti-Aliased Line System** - Foundation constants and utilities for high-quality line rendering

- [x] Subpixel precision constants
  - [x] LineSubpixelShift = 8 for 256 subpixel divisions
  - [x] LineSubpixelScale = 256 for coordinate scaling
  - [x] LineSubpixelMask = 255 for coordinate masking
  - [x] LineMaxCoord and LineMaxLength for coordinate limits
- [x] Medium resolution constants
  - [x] LineMRSubpixelShift = 4 for 16 subpixel divisions
  - [x] LineMRSubpixelScale = 16 for reduced precision operations
  - [x] LineMRSubpixelMask = 15 for coordinate masking

**Coordinate Conversion Functions**

- [x] Resolution conversion utilities
  - [x] LineMR() → reduce to medium resolution
  - [x] LineHR() → increase to high resolution
  - [x] LineDblHR() → double high resolution
  - [x] Efficient bit-shift based conversions
- [x] Integration with rendering pipeline
  - [x] Subpixel coordinate system for anti-aliasing
  - [x] Compatible with rasterizer and scanline systems
  - [x] Precision-aware line rendering

#### ✅ agg_math_stroke.h - Stroke mathematics (`internal/basics/math_stroke.go`) **COMPLETED**

**Stroke Mathematics System** - Comprehensive stroke geometry calculations for line rendering

- [x] Line cap and join style definitions
  - [x] LineCap enum (ButtCap, SquareCap, RoundCap)
  - [x] LineJoin enum (MiterJoin, RoundJoin, BevelJoin, etc.)
  - [x] InnerJoin enum (InnerBevel, InnerMiter, InnerJag, InnerRound)
- [x] MathStroke struct for stroke calculations
  - [x] Configurable stroke width and miter limits
  - [x] Width sign handling for positive/negative stroke widths
  - [x] Precision control with width epsilon

**Stroke Geometry Algorithms**

- [x] Line join calculations
  - [x] Miter join with configurable miter limit
  - [x] Round join with circle arc generation
  - [x] Bevel join with straight line connections
  - [x] Inner join handling for stroke intersections
- [x] Line cap calculations
  - [x] Butt cap (no extension)
  - [x] Square cap with width extension
  - [x] Round cap with semicircle generation
- [x] VertexConsumer interface for stroke output
  - [x] Add() method for vertex generation
  - [x] RemoveAll() method for stroke reset
  - [x] Integration with path storage systems

#### agg_shorten_path.h - Path shortening (`internal/path/shorten_path.go`)

**Path Shortening System** - Utility for trimming path length while preserving shape

- [ ] shorten_path → ShortenPath function - Remove length from path end
  - [ ] Generic vertex sequence input support
  - [ ] Configurable shortening distance
  - [ ] Closed path handling option
  - [ ] Preserves path shape during shortening

**Path Shortening Algorithm**

- [ ] Length-based path trimming
  - [ ] Backwards traversal from path end
  - [ ] Cumulative distance calculation
  - [ ] Vertex removal when distance exceeded
  - [ ] Partial vertex interpolation for precise length
- [ ] Edge case handling
  - [ ] Empty path handling
  - [ ] Single vertex paths
  - [ ] Shortening distance exceeding path length
  - [ ] Closed path continuation

**Integration with Path System**

- [ ] Vertex sequence compatibility
  - [ ] Works with path_storage and vertex arrays
  - [ ] Maintains vertex command types
  - [ ] Preserves path closure state
- [ ] Use in stroke processing
  - [ ] Dash pattern implementation
  - [ ] Arrow head positioning
  - [ ] Path trimming for decorative elements

#### agg_simul_eq.h - Simultaneous equations solver (`internal/math/simul_eq.go`)

**Linear Equation System Solver** - Mathematical solver for simultaneous linear equations

- [ ] SimulEq struct - Simultaneous equation system solver
  - [ ] Gaussian elimination with partial pivoting
  - [ ] Support for square matrices (NxN systems)
  - [ ] Numerical stability improvements
  - [ ] Error handling for singular matrices

**Equation Solving Methods**

- [ ] Linear system solution
  - [ ] Matrix setup and coefficient storage
  - [ ] Forward elimination phase
  - [ ] Back substitution for solution
  - [ ] Solution vector extraction
- [ ] Numerical robustness
  - [ ] Pivot selection for numerical stability
  - [ ] Scaling for improved precision
  - [ ] Condition number estimation
  - [ ] Singular matrix detection

**Mathematical Applications**

- [ ] Geometric transformations
  - [ ] Affine transformation matrix calculation
  - [ ] Curve fitting applications
  - [ ] Intersection point calculation
- [ ] Graphics pipeline integration
  - [ ] Coordinate system transformations
  - [ ] Projection matrix calculations
  - [ ] Geometric constraint solving

#### ✅ agg_vertex_sequence.h - Vertex sequence (`internal/array/vertex_sequence.go`, `internal/vcgen/vertex_sequence.go`) **COMPLETED**

**Vertex Sequence System** - Efficient storage and manipulation of vertex data

- [x] VertexSequence struct - Dynamic vertex array with distance tracking
  - [x] Generic vertex type support
  - [x] Automatic distance calculation between vertices
  - [x] Efficient append and remove operations
  - [x] Memory management with capacity growth
- [x] Vertex operations
  - [x] Add() method for appending vertices
  - [x] RemoveLast() and RemoveAll() for vertex removal
  - [x] Close() method for path closure
  - [x] Distance-based vertex access

**Vertex Storage Optimization**

- [x] Memory-efficient vertex arrays
  - [x] Contiguous memory layout for cache efficiency
  - [x] Minimal memory overhead per vertex
  - [x] Automatic capacity management
- [x] Distance calculation and caching
  - [x] Euclidean distance between consecutive vertices
  - [x] Cumulative distance tracking for path parameterization
  - [x] Distance-based vertex lookup and manipulation

**Dependencies**

- [x] Basic coordinate and math systems (`internal/basics/`) ✅ **COMPLETED**
- [x] Path storage integration (`internal/path/`) ✅ **COMPLETED**
- [x] Rendering buffer system (`internal/renderer/`) ✅ **COMPLETED**
- [ ] Pixel format system (`internal/pixfmt/`) - **REQUIRED** for alpha mask and gamma systems
  - [ ] RGBA and grayscale pixel format support
  - [ ] Multi-channel pixel access
  - [ ] Pixel format compatibility layer
- [ ] Effects processing system (`internal/effects/`) - **NEW** for blur effects
  - [ ] Image filtering infrastructure
  - [ ] Multi-pass processing support
  - [ ] Kernel-based image operations

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

- [x] agg_platform_support.h - Platform support interface ✅ **COMPLETED**
  - [x] PlatformSupport struct - Main platform abstraction (`internal/platform/platform_support.go`)
  - [x] WindowFlags enum - Window configuration flags
  - [x] PixelFormat enum - Comprehensive pixel format support
  - [x] InputFlags and KeyCode enums - Event handling types (`internal/platform/events.go`)
  - [x] RenderingContext - Enhanced rendering capabilities (`internal/platform/rendering_context.go`)
  - [x] Event handling system - Mouse, keyboard, and window events
  - [x] Multiple image buffer management
  - [x] Timer functionality for performance measurement
  - [x] Coordinate transformation support with resize handling
  - [x] Basic drawing primitives (lines, rectangles, circles)
  - [x] Pixel manipulation with alpha blending
  - [x] Comprehensive test coverage
  - [x] Example demonstration (`examples/platform/basic_demo/main.go`)

---

### Utilities (util/)

- [x] agg_color_conv.h - Color conversion utilities
- [x] agg_color_conv_rgb16.h - 16-bit RGB color conversion
- [x] agg_color_conv_rgb8.h - 8-bit RGB color conversion

## Core Implementation Files (src/)

### Basic Implementations

- [x] agg_arc.cpp - Arc generation implementation
- [x] agg_arrowhead.cpp - Arrowhead implementation
- [x] agg_ellipse.cpp - Ellipse generation implementation
- [x] agg_bezier_arc.cpp - Bezier arc implementation
- [x] agg_bspline.cpp - B-spline implementation
- [ ] agg_color_rgba.cpp - RGBA color implementation
- [ ] agg_curves.cpp - Curve approximation implementation
- [ ] agg_embedded_raster_fonts.cpp - Embedded fonts implementation
- [ ] agg_gsv_text.cpp - GSV text implementation
- [ ] agg_image_filters.cpp - Image filters implementation
- [ ] agg_line_aa_basics.cpp - Anti-aliased line basics
- [ ] agg_line_profile_aa.cpp - Anti-aliased line profile
- [x] agg_rounded_rect.cpp - Rounded rectangle implementation
- [ ] agg_sqrt_tables.cpp - Square root tables
- [x] agg_trans_affine.cpp - Affine transformation implementation
- [x] agg_trans_double_path.cpp - Double path transformation
- [x] agg_trans_single_path.cpp - Single path transformation
- [x] agg_trans_warp_magnifier.cpp - Warp magnifier implementation

### Vertex Generators

- [x] agg_vcgen_bspline.cpp - B-spline vertex generator
- [ ] agg_vcgen_contour.cpp - Contour vertex generator
- [x] agg_vcgen_dash.cpp - Dash vertex generator
- [x] agg_vcgen_markers_term.cpp - Terminal markers implementation
- [x] agg_vcgen_smooth_poly1.cpp - Polygon smoothing implementation
- [ ] agg_vcgen_stroke.cpp - Stroke vertex generator

### Vertex Processors

- [x] agg_vpgen_clip_polygon.cpp - Polygon clipping implementation
- [x] agg_vpgen_clip_polyline.cpp - Polyline clipping implementation
- [x] agg_vpgen_segmentator.cpp - Segmentator implementation

### Controls Implementation (src/ctrl/) - Interactive UI Controls System

The AGG controls system provides interactive UI widgets for graphics applications. All controls implement a common interface for mouse/keyboard interaction, coordinate transformation, and vector-based rendering.

#### Base Control Infrastructure (agg_ctrl.h)

**Core Control Interface**

- [ ] ctrl → Ctrl interface - Base control abstraction (`internal/ctrl/ctrl.go`)
  - [ ] Coordinate system management (x1, y1, x2, y2 bounding rectangle)
  - [ ] Y-axis flip support for different coordinate systems
  - [ ] Transformation matrix integration for scaling/rotation
  - [ ] Virtual destructor and lifecycle management

**Mouse and Keyboard Event System**

- [ ] Event handler interface methods
  - [ ] in_rect(x, y) → InRect(x, y) - Hit testing for mouse coordinates
  - [ ] on_mouse_button_down(x, y) → OnMouseButtonDown(x, y) - Mouse press handling
  - [ ] on_mouse_button_up(x, y) → OnMouseButtonUp(x, y) - Mouse release handling
  - [ ] on_mouse_move(x, y, flag) → OnMouseMove(x, y, pressed) - Mouse drag/hover
  - [ ] on_arrow_keys(l, r, d, u) → OnArrowKeys(left, right, down, up) - Keyboard navigation

**Coordinate Transformation System**

- [ ] Transformation support
  - [ ] transform(mtx) → SetTransform(mtx) - Apply transformation matrix
  - [ ] no_transform() → ClearTransform() - Remove transformation
  - [ ] transform_xy(x, y) → TransformXY(x, y) - Transform coordinates for rendering
  - [ ] inverse_transform_xy(x, y) → InverseTransformXY(x, y) - Transform mouse coordinates
  - [ ] scale() → Scale() - Get current transformation scale factor

**Control Rendering Framework**

- [ ] render_ctrl() → RenderCtrl() template function - Generic control rendering
  - [ ] Multi-path rendering support (controls can have multiple visual paths)
  - [ ] Color assignment per path
  - [ ] Rasterizer and scanline integration
  - [ ] Anti-aliased solid color rendering

#### agg_slider_ctrl.h/cpp - Horizontal/Vertical Slider Control

**Core Slider Structure**

- [ ] slider_ctrl_impl → SliderCtrlImpl struct - Slider implementation
  - [ ] Horizontal/vertical orientation support
  - [ ] Configurable border width and extra spacing
  - [ ] Text thickness for labels and value display
  - [ ] Range definition (minimum and maximum values)
  - [ ] Discrete step support for quantized values

**Value Management**

- [ ] Range and value operations
  - [ ] range(min, max) → SetRange(min, max) - Set slider value range
  - [ ] value() → Value() - Get current normalized value
  - [ ] value(v) → SetValue(v) - Set slider value with clamping
  - [ ] num_steps(n) → SetNumSteps(n) - Set discrete step count
  - [ ] normalize_value() - Internal value normalization

**Visual Customization**

- [ ] Appearance configuration
  - [ ] border_width(t, extra) → SetBorderWidth(thickness, extra) - Visual styling
  - [ ] label(fmt) → SetLabel(format) - Text label with printf-style formatting
  - [ ] text_thickness(t) → SetTextThickness(t) - Label text stroke width
  - [ ] descending() → IsDescending() - Reverse value direction

**Mouse Interaction**

- [ ] Interactive behavior
  - [ ] Click-to-position functionality
  - [ ] Drag handling for continuous adjustment
  - [ ] Preview value during drag operations
  - [ ] Coordinate mapping from screen to slider value
  - [ ] Hit testing for slider handle and track

**Keyboard Control**

- [ ] Keyboard navigation
  - [ ] Arrow key support for value adjustment
  - [ ] Step-wise increment/decrement
  - [ ] Large step support (Shift+Arrow)
  - [ ] Focus handling and visual feedback

**Vector Graphics Rendering**

- [ ] Multi-path rendering (6 paths total)
  - [ ] Path 0: Background track/groove
  - [ ] Path 1: Slider handle/thumb
  - [ ] Path 2: Value indicator/fill
  - [ ] Path 3: Border outline
  - [ ] Path 4: Text label rendering
  - [ ] Path 5: Focus indicator
- [ ] Vertex source interface implementation
  - [ ] rewind(path_id) → Rewind(pathID) - Reset path iteration
  - [ ] vertex(x, y) → Vertex() - Generate path vertices

#### agg_cbox_ctrl.h/cpp - Checkbox Control with Label

**Checkbox Structure**

- [ ] cbox_ctrl_impl → CboxCtrlImpl struct - Checkbox implementation
  - [ ] Boolean state management (checked/unchecked)
  - [ ] Text label support with configurable size
  - [ ] Compact square checkbox with text positioning
  - [ ] Text rendering integration (GSV text system)

**Text and Label Management**

- [ ] Label configuration
  - [ ] label() → Label() - Get current label text
  - [ ] label(text) → SetLabel(text) - Set label text with copy
  - [ ] text_size(h, w) → SetTextSize(height, width) - Configure text dimensions
  - [ ] text_thickness(t) → SetTextThickness(t) - Text stroke thickness
  - [ ] Dynamic text layout and positioning

**State Management**

- [ ] Boolean state operations
  - [ ] status() → IsChecked() - Get checkbox state
  - [ ] status(state) → SetChecked(state) - Set checkbox state
  - [ ] Toggle functionality on click
  - [ ] State change event handling

**Mouse Interaction**

- [ ] Click handling
  - [ ] Hit testing for checkbox and label area
  - [ ] Toggle on mouse click
  - [ ] Visual feedback during interaction
  - [ ] Hover state indication

**Vector Graphics Rendering**

- [ ] Multi-path rendering (3 paths total)
  - [ ] Path 0: Checkbox square outline
  - [ ] Path 1: Checkmark symbol (when checked)
  - [ ] Path 2: Text label rendering
- [ ] GSV text system integration
  - [ ] gsv_text → GSVText for text rendering
  - [ ] conv_stroke → ConvStroke for text outline
  - [ ] Text positioning relative to checkbox

#### agg_bezier_ctrl.h/cpp - Interactive Cubic Bezier Curve Editor

**Bezier Curve Structure**

- [ ] bezier_ctrl_impl → BezierCtrlImpl struct - Bezier control implementation
  - [ ] 4-point cubic bezier curve definition (P0, P1, P2, P3)
  - [ ] Interactive control point manipulation
  - [ ] curve4 integration for curve mathematics
  - [ ] Real-time curve preview during editing

**Control Point Management**

- [ ] Point coordinate access
  - [ ] x1(), y1(), x2(), y2(), x3(), y3(), x4(), y4() → P1(), P2(), P3(), P4() accessors
  - [ ] curve(x1,y1,x2,y2,x3,y3,x4,y4) → SetCurve() - Set all control points
  - [ ] curve() → Curve() - Get curve4 reference for detailed operations
  - [ ] Point constraint and validation

**Interactive Editing**

- [ ] Point manipulation
  - [ ] Individual control point selection and dragging
  - [ ] Visual feedback for selected points
  - [ ] Snap-to-grid functionality (optional)
  - [ ] Point coordinate display/editing

**Mouse Interaction**

- [ ] Interactive behavior
  - [ ] Control point hit testing and selection
  - [ ] Drag operations for curve reshaping
  - [ ] Multi-point selection support
  - [ ] Real-time curve updates during manipulation

**Vector Graphics Rendering**

- [ ] Curve visualization
  - [ ] Smooth curve rendering using conv_curve
  - [ ] Control point visualization (handles)
  - [ ] Control polygon display (construction lines)
  - [ ] Selection highlighting

**Integration with Curve Mathematics**

- [ ] curve4 → Curve4 integration (`internal/curves/` dependency)
  - [ ] Bezier curve calculation and approximation
  - [ ] Curve subdivision and flattening
  - [ ] Arc length parameterization
  - [ ] Derivative calculation for tangent vectors

#### agg_gamma_ctrl.h/cpp - Interactive Gamma Correction Curve Editor

**Gamma Curve Structure**

- [ ] gamma_ctrl_impl → GammaCtrlImpl struct - Gamma control implementation
  - [ ] Gamma correction curve visualization and editing
  - [ ] Multi-point spline-based curve definition
  - [ ] Real-time gamma preview functionality
  - [ ] Integration with color management pipeline

**Curve Point Management**

- [ ] Spline control points
  - [ ] Variable number of control points along gamma curve
  - [ ] Point insertion and deletion functionality
  - [ ] Automatic curve smoothing between points
  - [ ] Curve interpolation and extrapolation

**Interactive Editing**

- [ ] Curve manipulation
  - [ ] Click-to-add control points
  - [ ] Drag existing points for curve adjustment
  - [ ] Delete points with keyboard/right-click
  - [ ] Real-time curve preview during editing

**Gamma Correction Integration**

- [ ] Color correction features
  - [ ] Gamma value calculation and display
  - [ ] Curve-to-lookup-table conversion
  - [ ] Integration with pixel format gamma correction
  - [ ] Preview rendering with applied gamma

**Vector Graphics Rendering**

- [ ] Curve visualization
  - [ ] Smooth gamma curve rendering
  - [ ] Grid background for reference
  - [ ] Control point visualization
  - [ ] Curve interpolation display

#### agg_gamma_spline.h/cpp - Gamma Spline Mathematics

**Spline Mathematics**

- [ ] gamma_spline → GammaSpline struct - Spline-based gamma curve
  - [ ] Cubic spline interpolation for smooth gamma curves
  - [ ] Automatic control point generation
  - [ ] Curve smoothing and regularization
  - [ ] Efficient evaluation for pixel processing

**Spline Operations**

- [ ] Mathematical operations
  - [ ] Spline coefficient calculation
  - [ ] Control point optimization
  - [ ] Curve derivative calculation
  - [ ] Fast lookup table generation

**Integration with Gamma Control**

- [ ] Backend mathematics
  - [ ] Support for interactive gamma editing
  - [ ] Real-time curve updates
  - [ ] Numerical stability for extreme gamma values
  - [ ] Memory-efficient curve representation

#### agg_polygon_ctrl.h/cpp - Interactive Polygon Editor

**Polygon Structure**

- [ ] polygon_ctrl_impl → PolygonCtrlImpl struct - Polygon control implementation
  - [ ] Variable vertex count support
  - [ ] Interactive vertex manipulation
  - [ ] Polygon closing and opening functionality
  - [ ] Self-intersection detection and handling

**Vertex Management**

- [ ] Point operations
  - [ ] Dynamic vertex addition and removal
  - [ ] Vertex coordinate access and modification
  - [ ] Polygon centroid calculation
  - [ ] Vertex ordering and winding direction

**Interactive Editing**

- [ ] Polygon manipulation
  - [ ] Vertex selection and dragging
  - [ ] Edge manipulation (move entire edges)
  - [ ] Vertex insertion on edge click
  - [ ] Vertex deletion with keyboard

**Mouse Interaction**

- [ ] Interactive behavior
  - [ ] Vertex hit testing and selection
  - [ ] Edge proximity detection
  - [ ] Polygon interior hit testing
  - [ ] Multi-vertex selection support

**Vector Graphics Rendering**

- [ ] Polygon visualization
  - [ ] Filled polygon rendering
  - [ ] Outline stroke rendering
  - [ ] Vertex handle visualization
  - [ ] Selection highlighting

#### agg_rbox_ctrl.h/cpp - Radio Button Group Control

**Radio Button Structure**

- [ ] rbox_ctrl_impl → RboxCtrlImpl struct - Radio button group implementation
  - [ ] Multiple radio button options in group
  - [ ] Mutual exclusion logic (only one selected)
  - [ ] Text labels for each option
  - [ ] Vertical or horizontal layout

**Option Management**

- [ ] Radio button options
  - [ ] Dynamic option addition and removal
  - [ ] Option text label configuration
  - [ ] Selected option tracking
  - [ ] Option enabling/disabling

**Selection Logic**

- [ ] Group behavior
  - [ ] Mutual exclusion enforcement
  - [ ] Selection change event handling
  - [ ] Default selection support
  - [ ] Programmatic selection control

**Mouse Interaction**

- [ ] Interactive behavior
  - [ ] Option click handling
  - [ ] Selection change on click
  - [ ] Visual feedback for hover state
  - [ ] Keyboard navigation between options

**Vector Graphics Rendering**

- [ ] Multi-option visualization
  - [ ] Radio button circles for each option
  - [ ] Selection indicator (filled circle)
  - [ ] Text labels for each option
  - [ ] Group layout and spacing

#### agg_scale_ctrl.h/cpp - Interactive Scale/Zoom Control

**Scale Control Structure**

- [ ] scale_ctrl_impl → ScaleCtrlImpl struct - Scale control implementation
  - [ ] Scale factor adjustment widget
  - [ ] Zoom in/out functionality
  - [ ] Scale value display and editing
  - [ ] Logarithmic and linear scale modes

**Scale Management**

- [ ] Scale operations
  - [ ] Scale factor setting and retrieval
  - [ ] Minimum and maximum scale limits
  - [ ] Scale step increments
  - [ ] Scale reset to default functionality

**Interactive Control**

- [ ] User interaction
  - [ ] Click-and-drag scaling
  - [ ] Discrete step scaling with buttons
  - [ ] Keyboard scaling shortcuts
  - [ ] Mouse wheel scaling support

**Visual Feedback**

- [ ] Scale visualization
  - [ ] Current scale value display
  - [ ] Scale range indicator
  - [ ] Zoom level visualization
  - [ ] Scale increment markers

#### agg_spline_ctrl.h/cpp - Interactive Spline Curve Editor

**Spline Structure**

- [ ] spline_ctrl_impl → SplineCtrlImpl struct - Spline control implementation
  - [ ] Variable control point spline curves
  - [ ] Smooth curve interpolation between points
  - [ ] Curve tension and continuity control
  - [ ] Real-time spline preview

**Control Point Management**

- [ ] Point operations
  - [ ] Dynamic control point addition and removal
  - [ ] Point coordinate manipulation
  - [ ] Automatic curve smoothing
  - [ ] Point constraint handling

**Spline Mathematics**

- [ ] Curve computation
  - [ ] Cubic spline interpolation
  - [ ] Curve subdivision and approximation
  - [ ] Tangent vector calculation
  - [ ] Arc length parameterization

**Interactive Editing**

- [ ] Curve manipulation
  - [ ] Control point selection and dragging
  - [ ] Curve tension adjustment
  - [ ] Point insertion on curve click
  - [ ] Point deletion functionality

**Vector Graphics Rendering**

- [ ] Spline visualization
  - [ ] Smooth curve rendering
  - [ ] Control point visualization
  - [ ] Control polygon display
  - [ ] Curve tangent indicators

**Dependencies**

- [ ] Required components for full controls implementation
  - [ ] agg_basics.h → internal/basics package
  - [ ] agg_math.h → internal/math package
  - [ ] agg_trans_affine.h → internal/transform package ✅ **COMPLETED**
  - [ ] agg_color_rgba.h → internal/color package (partially complete)
  - [ ] agg_conv_stroke.h → internal/conv package (stroke converter)
  - [ ] agg_conv_curve.h → internal/conv package (curve converter)
  - [ ] agg_gsv_text.h → internal/gsv package (vector text rendering)
  - [ ] agg_path_storage.h → internal/path package
  - [ ] agg_ellipse.h → internal/shapes package
  - [ ] Rasterizer and scanline systems for rendering
  - [ ] Vertex source interface compatibility

**Integration Notes**

- [ ] Platform integration considerations

  - [ ] Event system integration (mouse, keyboard)
  - [ ] Window system coordinate mapping
  - [ ] Device pixel ratio handling for high-DPI displays
  - [ ] Clipboard integration for cut/copy/paste operations

- [ ] Usage patterns
  - [ ] Control instantiation and lifecycle management
  - [ ] Event loop integration
  - [ ] Control state serialization/deserialization
  - [ ] Theme and styling support

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

- [x] agg_platform_support.cpp (generic interface)

### Platform-Specific (Optional - for examples)

- [ ] src/platform/X11/agg_platform_support.cpp - X11 support (implemented, requires X11 dev headers)
- [ ] src/platform/win32/agg_platform_support.cpp - Win32 support
- [ ] src/platform/win32/agg_win32_bmp.cpp - Win32 bitmap support
- [ ] src/platform/mac/agg_platform_support.cpp - macOS support
- [ ] src/platform/mac/agg_mac_pmap.cpp - macOS pixmap support
- [ ] src/platform/sdl/agg_platform_support.cpp - SDL support
- [x] src/platform/sdl2/agg_platform_support.cpp - SDL2 support (COMPLETE - working with go-sdl2 dependency)

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

---

## Test Failures and Implementation Issues

For a comprehensive catalog of all failing tests, implementation deviations from C++ AGG, detailed analysis, and fix priorities, see **TEST_TASKS.md**.

This includes:

- Algorithm failures (premultiplied alpha blending, rasterizer clipping, etc.)
- Implementation deviations from C++ AGG behavior
- Build failures and dependency issues
- Testing strategies and fix recommendations
- Priority ordering for addressing issues

Always consult TEST_TASKS.md when working on test failures or investigating behavioral differences from the original AGG library.
