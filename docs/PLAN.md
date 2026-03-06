# SIMD Acceleration Plan for AGG Go Port

## Overview

This plan outlines opportunities for SIMD acceleration of hot rendering paths
in the AGG Go port. The approach follows the same architecture proven in
[algo-fft](../algo-fft): Plan 9 assembly with runtime CPU detection and
graceful fallback to scalar Go.

## Architecture

### CPU Detection & Dispatch (Phase 0 — Infrastructure)

Reuse the pattern from `algo-fft/internal/cpu/`:

```
internal/simd/
├── cpu.go               # Feature detection (sync.Once cached)
├── detect_amd64.go      # golang.org/x/sys/cpu → HasSSE2, HasAVX2
├── detect_arm64.go      # HasNEON (baseline on ARMv8)
├── detect_generic.go    # All features false
├── blend_amd64.go       # //go:build amd64 && !purego — dispatch
├── blend_arm64.go       # //go:build arm64 && !purego — dispatch
├── blend_generic.go     # Scalar fallback wrappers
├── asm_amd64/
│   ├── decl.go          # //go:noescape function declarations
│   ├── blend_sse2.s     # SSE2 kernels
│   ├── blend_avx2.s     # AVX2 kernels
│   └── clear_avx2.s     # Buffer fill kernels
└── asm_arm64/
    ├── decl.go
    ├── blend_neon.s      # NEON kernels
    └── clear_neon.s
```

Build tags: `//go:build {arch} && !purego`

The `purego` tag disables all assembly (for testing, WASM, unusual platforms).

Runtime dispatch selects the best available implementation at init time, not
per-call. Function pointers are stored in a package-level `Impl` struct.

**Effort**: ~2 days
**Dependencies**: None

---

## Phase 1 — Bulk Pixel Operations

These operate on contiguous `[]uint8` row buffers and are the simplest to
vectorize because the loop bodies are uniform and data-parallel.

### 1a. CopyHline / Clear (memset-style fill)

**Profile**: 14.6% of lion CPU (CopyHline alone for ClearAll)

Fills N consecutive RGBA pixels with a single 4-byte value. Currently a
scalar loop writing 4 bytes at a time.

| ISA  | Strategy                 | Pixels/cycle | Register |
| ---- | ------------------------ | ------------ | -------- |
| SSE2 | MOVDQU 16B stores (4px)  | 4            | XMM      |
| AVX2 | VMOVDQU 32B stores (8px) | 8            | YMM      |
| NEON | VST1 16B stores (4px)    | 4            | Q-reg    |

Assembly signature:

```go
//go:noescape
func fillRGBA(dst []uint8, r, g, b, a uint8, count int)
```

The RGBA value is broadcast to fill a full vector register, then stored in an
unrolled loop. Tail pixels (count % vector width) handled scalar.

**Impact**: High — every `ClearAll` and solid `CopyBar`
**Effort**: ~1 day per ISA
**Complexity**: Low (no reads, just writes)
**Rating**: ★★★★★

### 1b. BlendSolidHspan (coverage-varied solid color blend)

**Profile**: 13.7% flat + 43.4% cumulative in lion (the single hottest path)

Blends a single solid RGBA color into N destination pixels, each with a
different coverage value from a `[]uint8` coverage array. This is the core of
anti-aliased rendering.

Per-pixel scalar logic:

```
alpha = MultCover(srcA, covers[i])     // uint8 × uint8 → uint8
dst[R] = Lerp(dst[R], srcR, alpha)     // 3× for RGB
dst[G] = Lerp(dst[G], srcG, alpha)
dst[B] = Lerp(dst[B], srcB, alpha)
dst[A] = Prelerp(dst[A], alpha, alpha) // 1× for A
```

`Lerp(p, q, a)` is: `p + (((q-p)*a + 128 - (p>q)) >> 8 + (same)) >> 8`

| ISA  | Strategy                                 | Pixels/iteration |
| ---- | ---------------------------------------- | ---------------- |
| SSE2 | Unpack bytes → 16-bit, PMULLW, pack back | 4                |
| AVX2 | VPUNPCKLBW/VPMULLW on 32-byte lanes      | 8                |
| NEON | VMULL.U8 / VRSHRN                        | 4–8              |

The key insight: all four channels can be processed in a single 32-bit or
64-bit SIMD lane. Load 4 dst bytes + 1 cover byte, broadcast src color,
compute blend, store 4 bytes.

**Impact**: Very high — dominates every AA-rendered shape
**Effort**: ~2–3 days per ISA
**Complexity**: Medium (signed intermediate for Lerp, conditional `p>q` bit)
**Rating**: ★★★★★

### 1c. BlendHline (uniform coverage blend)

**Profile**: Called from RenderScanlineAASolid for solid interior spans

Same as 1b but with a single constant coverage for all pixels. This removes
the per-pixel coverage load, making it even simpler. The blend computation is
identical but the alpha is broadcast once.

**Impact**: Medium — only solid interior spans
**Effort**: ~0.5 day per ISA (variant of 1b)
**Complexity**: Low
**Rating**: ★★★★☆

### 1d. BlendColorHspan (per-pixel color + coverage blend)

**Profile**: Used by gradient and image rendering

Blends N source colors (from a `[]RGBA8` array) into N destination pixels with
per-pixel coverage. More complex than 1b because both source color and coverage
vary per pixel.

**Impact**: High for gradient-heavy and image-transform demos
**Effort**: ~1 day per ISA (extension of 1b)
**Complexity**: Medium
**Rating**: ★★★★☆

---

## Phase 2 — Premultiply / Demultiply

### 2a. Premultiply (whole-buffer)

Multiplies R, G, B by A for every pixel in the buffer. Used when converting
to/from premultiplied alpha. Pure data-parallel: read 4 bytes, multiply 3
channels by the 4th, write back.

| ISA  | Strategy                                         |
| ---- | ------------------------------------------------ |
| SSE2 | PMULLW on unpacked bytes, 4 pixels per iteration |
| AVX2 | 8 pixels per iteration with VPMULLW              |
| NEON | VMULL.U8 + VRSHRN, 8 pixels per iteration        |

**Impact**: Medium — used in image compositing workflows
**Effort**: ~1 day per ISA
**Complexity**: Low
**Rating**: ★★★☆☆

### 2b. Demultiply (whole-buffer)

Divides R, G, B by A. Requires integer division or reciprocal lookup.

| ISA  | Strategy                                     |
| ---- | -------------------------------------------- |
| SSE2 | Convert to float, DIVPS, convert back (4 px) |
| AVX2 | VDIVPS or VRCPPS + Newton-Raphson (8 px)     |
| NEON | FRECPE + FRECPS refinement (4–8 px)          |

**Impact**: Medium
**Effort**: ~2 days per ISA (division is tricky)
**Complexity**: High (division, zero-alpha guard)
**Rating**: ★★★☆☆

---

## Phase 3 — Composite Blend Modes

### 3a. Porter-Duff / SVG Composite Operations

**Profile**: Dominates compositing demos (compositing, compositing2, blend_modes)

The `CompositeBlender.BlendPix` normalizes to `float64`, dispatches to one of
39 blend mode functions, then converts back to `uint8`. Each mode is a simple
formula on normalized [0,1] values.

Vectorization approach: batch 4 or 8 pixels through the same blend mode.
Convert uint8 → float32 (not float64 — sufficient precision for 8-bit output),
apply blend formula, convert back.

Common modes and their SIMD-friendliness:

| Mode       | Formula                     | SIMD fit           |
| ---------- | --------------------------- | ------------------ |
| SrcOver    | Cs + Cd × (1-As)            | Excellent          |
| Multiply   | Cs × Cd                     | Excellent          |
| Screen     | Cs + Cd - Cs×Cd             | Excellent          |
| Overlay    | conditional Multiply/Screen | Good (mask)        |
| Darken     | min(Cs, Cd)                 | Excellent (PMINUB) |
| Lighten    | max(Cs, Cd)                 | Excellent (PMAXUB) |
| Difference | abs(Cs - Cd)                | Excellent          |

**Impact**: Very high for compositing-heavy demos
**Effort**: ~3–5 days per ISA (many modes, but formulaic)
**Complexity**: Medium (float conversion, mode dispatch)
**Rating**: ★★★★☆

---

## Phase 4 — Span Generation

### 4a. Gradient Span Generation

Gradient rendering generates N pixels per span by:

1. Stepping an interpolator (add dx, dy per pixel)
2. Computing gradient distance (e.g., `x` for linear, `sqrt(x²+y²)` for radial)
3. Looking up color from a 256-entry LUT

The interpolator step and LUT lookup are hard to vectorize (data-dependent
index). But the distance calculation for linear gradients is trivially
vectorizable: it's just an incrementing integer.

**Impact**: High for gradient demos (gradients, alpha_gradient)
**Effort**: ~2 days per ISA
**Complexity**: Medium (gather loads for LUT)
**Rating**: ★★★☆☆

### 4b. Image Span Filtering (Bilinear, Bicubic)

Convolution kernels applied per-pixel for image resampling. The inner loop
multiplies source pixels by int16 filter weights and accumulates. Classic
SIMD territory.

| ISA  | Strategy                                       |
| ---- | ---------------------------------------------- |
| SSE2 | PMADDWD for 16-bit weight × pixel accumulation |
| AVX2 | VPMADDWD on wider vectors                      |
| NEON | VMLA.S16 accumulate                            |

**Impact**: High for image transformation demos
**Effort**: ~3 days per ISA
**Complexity**: High (kernel size varies, edge handling)
**Rating**: ★★★☆☆

---

## Phase 5 — Alpha Mask Operations

### 5a. Alpha Mask FillHspan

Reads mask values from a grayscale buffer and multiplies with coverage values.
Simple byte×byte multiply on arrays.

**Impact**: Medium — alpha mask demos only
**Effort**: ~1 day per ISA
**Complexity**: Low
**Rating**: ★★☆☆☆

### 5b. RGB-to-Gray Mask Conversion

Weighted sum: `(77×R + 150×G + 29×B) >> 8`. Classic SIMD reduction.

| ISA  | Strategy                               |
| ---- | -------------------------------------- |
| SSE2 | PMADDUBSW with weight vector, 8 pixels |
| AVX2 | 16 pixels per iteration                |
| NEON | VMULL.U8 + VPADD, 8 pixels             |

**Impact**: Low — only used during mask setup
**Effort**: ~0.5 day per ISA
**Complexity**: Low
**Rating**: ★★☆☆☆

---

## Phase 6 — Gamma / LUT Application

### 6a. ApplyGammaDir (LUT-based gamma)

Applies a 256-entry byte LUT to R, G, B channels of every pixel. The LUT
lookup itself is not vectorizable (random byte access), but SSSE3's PSHUFB
can serve as a 16-entry LUT, and splitting into high/low nibble lookups
enables full 256-entry vectorized LUT application.

Alternatively, AVX2's VPGATHERDD can gather 8 values at once (but with high
latency).

**Impact**: Low — gamma applied once at setup, not per-frame
**Effort**: ~2 days per ISA
**Complexity**: High (LUT vectorization is non-trivial)
**Rating**: ★☆☆☆☆

---

## Summary & Recommended Order

| Phase | Target          | CPU % (lion) | Demos most affected       | Rating | Effort   |
| ----- | --------------- | ------------ | ------------------------- | ------ | -------- |
| 0     | Infrastructure  | —            | All                       | Prereq | 2d       |
| 1a    | CopyHline/Clear | 14.6%        | All                       | ★★★★★  | 1d/ISA   |
| 1b    | BlendSolidHspan | 43.4% cum    | All AA rendering          | ★★★★★  | 2–3d/ISA |
| 1c    | BlendHline      | part of 1b   | All AA rendering          | ★★★★☆  | 0.5d/ISA |
| 1d    | BlendColorHspan | —            | Gradients, images         | ★★★★☆  | 1d/ISA   |
| 2a    | Premultiply     | —            | Compositing               | ★★★☆☆  | 1d/ISA   |
| 2b    | Demultiply      | —            | Compositing               | ★★★☆☆  | 2d/ISA   |
| 3a    | Composite modes | dom.         | compositing, blend_modes  | ★★★★☆  | 3–5d/ISA |
| 4a    | Gradient spans  | —            | gradients, alpha_gradient | ★★★☆☆  | 2d/ISA   |
| 4b    | Image filtering | —            | image_transforms, image1  | ★★★☆☆  | 3d/ISA   |
| 5a    | Alpha mask fill | —            | alpha_mask demos          | ★★☆☆☆  | 1d/ISA   |
| 5b    | RGB→Gray mask   | —            | alpha_mask demos          | ★★☆☆☆  | 0.5d/ISA |
| 6a    | Gamma LUT       | —            | gamma_correction          | ★☆☆☆☆  | 2d/ISA   |

### Recommended ISA priority

1. **AVX2** (amd64) — 8 pixels/iteration, widest adoption on modern x86
2. **SSE2** (amd64/386) — 4 pixels/iteration, baseline for all x86-64
3. **NEON** (arm64) — 4–8 pixels/iteration, baseline on ARMv8 (Apple Silicon, RPi4+)

### Quick wins (Phase 0 + 1a + 1b)

Just the infrastructure plus `CopyHline` and `BlendSolidHspan` would cover
**~58% of lion rendering CPU time** and benefit every single demo. This is
where to start.

### Demos that benefit most from each phase

| Demo                      | Phase 1 | Phase 2 | Phase 3 | Phase 4 | Phase 5 |
| ------------------------- | ------- | ------- | ------- | ------- | ------- |
| lion, lion_outline        | ★★★★★   |         |         |         |         |
| shapes, circles, lines    | ★★★★★   |         |         |         |         |
| gradients, alpha_gradient | ★★★★☆   |         |         | ★★★★★   |         |
| compositing, blend_modes  | ★★★☆☆   | ★★★★★   | ★★★★★   |         |         |
| image_transforms, image1  | ★★★☆☆   | ★★★★☆   |         | ★★★★★   |         |
| alpha_mask, alpha_mask2/3 | ★★★★☆   |         |         |         | ★★★★★   |
| blur, simple_blur         | ★★★☆☆   |         |         | ★★★★☆   |         |
| conv_stroke, conv_contour | ★★★★★   |         |         |         |         |
| gouraud, gouraud_mesh     | ★★★☆☆   |         |         | ★★★★☆   |         |
| flash_rasterizer          | ★★★★★   |         | ★★★☆☆   |         |         |

---

## Notes

- The `purego` build tag must always be respected for WASM and test builds.
- All SIMD kernels must produce bit-identical output to the scalar Go code.
  Use the existing visual test infrastructure to verify.
- Start with AVX2 for development (easiest to debug, widest registers), then
  port down to SSE2 (subset of AVX2 instructions), then NEON (different ISA
  but similar concepts).
- The `algo-fft` project has 166 assembly files — AGG will need far fewer
  because pixel operations are simpler than FFT butterflies. Estimate ~15–25
  assembly files total across all ISAs for Phase 1–3.
