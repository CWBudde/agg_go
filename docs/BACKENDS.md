# Backend Matrix

This repository now has one canonical codebase and two engine modes exposed
through the opt-in `engine` package:

- `port`: the default pure-Go renderer in the root `agg` package
- `cpp`: an optional in-repo C++ AGG-backed engine selected through
  `engine.Config{Kind: engine.CPP}`

The pure-Go root API remains the default path for existing users. Nothing in
this file changes the compatibility story for callers that stay on `agg`.

## Build Modes

| Engine | Build tags                   | Availability                 | Notes                                                                                                                |
| ------ | ---------------------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `port` | none                         | available                    | Default pure-Go path. No cgo required.                                                                               |
| `cpp`  | none                         | unavailable                  | Returns a typed `engine.ErrUnavailable` because the `agogo` build tag is missing.                                    |
| `cpp`  | `agogo` with `CGO_ENABLED=0` | unavailable                  | Returns a typed `engine.ErrUnavailable` because cgo is disabled.                                                     |
| `cpp`  | `agogo`                      | unavailable today            | The current native layer compiles, but this build still rejects the backend when it is only the stub implementation. |
| `cpp`  | `agogo aggreal`              | available for current subset | Temporary real-native development path while dependency detection is being moved into this repo.                     |

## Native Dependencies

The temporary real-native `cpp` path currently assumes:

- AGG 2.6 headers, currently found via `/usr/include/agg2`
- `libagg`
- `libaggfontfreetype`
- `pkg-config` support for `freetype2`
- cgo-enabled Go builds
- a usable font file at runtime for text rendering

The `aggreal` split is temporary. The plan is to collapse that back into the
plain `agogo` build once dependency probing and unavailable-path reporting are
handled cleanly inside this repository.

## Capability Matrix

This matrix reflects the `engine.Capabilities(...)` contract and the currently
implemented high-level facade subset.

| Capability      | `port` | `cpp` stub (`agogo`) | `cpp` real (`agogo aggreal`) | Notes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --------------- | ------ | -------------------- | ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `solid_style`   | yes    | unavailable          | yes                          | Fill and stroke colors work in both available engines, with symmetric `GetFillColor`/`GetStrokeColor`/`GetLineWidth`/`GetLineCap`/`GetLineJoin` readback.                                                                                                                                                                                                                                                                                                                                                                                                        |
| `path`          | yes    | unavailable          | yes                          | `MoveTo`, `LineTo`, `QuadTo`, `CubicTo`, `ClosePath`, fill, stroke, rectangle, and circle helpers are available.                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `transforms`    | yes    | unavailable          | yes                          | Translation, rotation, scale, and reset are implemented, plus `GetTransform` readback of the cumulative affine matrix (`agg.Transformations`, AGG order).                                                                                                                                                                                                                                                                                                                                                                                                        |
| `clip_box`      | yes    | unavailable          | yes                          | The current C++ backend clips fill, stroke, and image operations.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `compositing`   | yes    | unavailable          | partial                      | Vector fill/stroke, gradient fills/strokes, **and** image/text draw are faithful for the **full** `agg.BlendMode` set (Porter-Duff plus separable blend modes): solids render directly through a comp-op pixfmt (straight-alpha adaptor); gradients and the text coverage layer composite the recoloured/glyph layer through the same operator using the shape/glyph AA coverage as cover; image blits composite each covered pixel through `comp_op_adaptor_rgba_plain`. (`partial` reflects the remaining dashed-stroke-under-transform gap, not blend modes.) |
| `image_draw`    | yes    | unavailable          | partial                      | Copy, region draw, scaling, and quad mapping are implemented, including draw under an active transform (the destination rectangle is mapped through the CTM and blitted via the quad path) and under the full blend-mode set (comp-op per covered pixel). Sampling is nearest-neighbour (CPU), so it diverges from the port's bilinear filter within the documented image envelope.                                                                                                                                                                              |
| `image_export`  | yes    | unavailable          | yes                          | PNG, JPEG, `ToGoImage`, and `ToStandardImage` work through the facade.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `image_interop` | yes    | unavailable          | no                           | The C++ backend still rejects `Premultiply` and `Demultiply`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `gradients`     | yes    | unavailable          | yes                          | The current C++ subset applies fill and stroke gradients during actual rendering.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `text`          | yes    | unavailable          | partial                      | Font loading, hinting, draw, measure, and bounds exist in the real-native path. Advanced text features are still out of scope.                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `dashed_stroke` | yes    | unavailable          | yes                          | `AddDash`/`RemoveAllDashes`/`DashStart`/`GetDashStart` drive AGG `conv_dash` on both backends.                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |

## Intentional Gaps

- The `engine` facade is deliberately narrower than the root `agg` API.
- The current C++ backend is still partial even in `agogo aggreal`.
- All drawing paths — solid/gradient fills/strokes and image/text draw —
  composite through AGG's comp-op operator under the full blend-mode set. Image
  sampling itself is still nearest-neighbour (CPU), so blended image draw diverges
  from the port's bilinear filter within the documented image envelope.
- The current real-native build uses the temporary `aggreal` tag and
  compile-time system-library assumptions.

## Error Contract

The `engine` package is expected to fail explicitly rather than silently
switching behavior:

- selecting `cpp` in an unavailable build returns `engine.ErrUnavailable`
- selecting an unsupported feature returns `engine.ErrUnsupportedCapability`
- mixing resources from different engines returns `engine.ErrEngineMismatch`

The C++ backend must never silently fall back to the Go port, and stub builds
must never be advertised as valid `cpp` support.

## Cross-Backend Conformance

The backend-neutral scene corpus (`engine/scene`) renders the same high-level
operations through every available engine. Three artifacts consume it:

- `tests/conformance/conformance_test.go` — renders each scene through `port`
  and `cpp` and compares the two with documented tolerance envelopes.
- `cmd/engine-compare` — writes `<scene>_port.png`, `<scene>_cpp.png`, and an
  amplified `<scene>_diff.png` per scene for manual inspection.
- `tests/conformance/benchmark_test.go` (`BenchmarkCorpusRender`) — runs the
  corpus through each available engine and reports ns/op and allocs/op.

All three are untagged: they compile and run in the default build (where only
`port` is available and the cross-backend test skips), and exercise `cpp` when
built with `-tags "agogo aggreal"` (add `freetype` for the text scene).

### Tolerance Envelopes

Both engines are 8-bit RGBA, so most operations match everywhere except
anti-aliased edges, where two independent rasterizers disagree on a small,
bounded fraction of pixels. Strict scenes must stay within these envelopes
(`Tolerance` = max per-channel LSB delta before a pixel counts as "different";
`MaxDifferentRatio` = bound on the fraction of such pixels):

| Scene class                              | Tolerance | MaxDifferentRatio | Rationale                                                                                              |
| ---------------------------------------- | --------- | ----------------- | ------------------------------------------------------------------------------------------------------ |
| solid / dashed / path / clip             | 2         | 0.025             | Edge-AA disagreement on ~1.5% of pixels; bulk identical within 2 LSB.                                  |
| compositing (src/srcover/clear/multiply) | 2         | 0.005             | Byte-exact apart from 1-LSB premul/demul rounding on AA edges; includes the separable `multiply` mode. |
| linear / radial gradient                 | 3         | 0.020             | Independent gradient interpolation rounding.                                                           |
| compositing gradient                     | 3         | 0.020             | Gradient interp plus a thin src-replaced AA rim (~0.8% of pixels).                                     |
| scaled image (`image_scaled`)            | 4         | 0.080             | Independent samplers disagree along upscaled hard edges (~6%).                                         |
| blended image (`image_blend`)            | 4         | 0.080             | Image over a colour field under `multiply`; sampler noise carried through the operator (~5.3%).        |
| transformed image (`image_affine`)       | 4         | 0.100             | As scaled image, plus rotation lengthening every hard edge; CPP nearest vs port bilinear (~8.8%).      |
| text (`text_basic`)                      | 8         | 0.100             | Native AGG vs Go-port FreeType AA/hinting; observed ~0.5% in practice.                                 |

### Compositing parity

The C++ backend renders solid fills and strokes **directly** through a comp-op
pixfmt (`pixfmt_custom_blend_rgba`), so the operator is applied per span with
anti-aliased coverage — exactly as AGG's `Agg2D` does. Because the destination
buffer is straight-alpha, a custom adaptor premultiplies the destination on
read, evaluates the operator in premultiplied space, then demultiplies the
result on write, matching the port's `CompositeBlenderPlain`. The previous
approach rendered each shape into a transparent layer and composited the whole
layer over the destination, which applied the operator across the entire clip
rectangle — so `src`/`clear` wiped the untouched background. The
`compositing_src`, `compositing_srcover`, and `compositing_clear` scenes are now
strict (byte-exact within 1-LSB rounding) rather than logged divergences.

Gradient fills and strokes are faithful under every supported blend mode too.
The shape is rasterised into a transparent layer to capture its anti-aliased
coverage; the layer is then recoloured with the straight gradient colour (RGB
plus the gradient's own alpha, not premultiplied by coverage) and composited
through the same comp-op operator using the captured coverage as the per-pixel
rasterizer cover (`agg_go_cpp_image_composite_cover` →
`comp_op_adaptor_rgba_plain::blend_pix`). Pixels outside the shape keep cover 0,
so `src`/`clear` no longer wipe the background. This mirrors AGG's
`renderer_scanline_aa` + `span_gradient` + comp-op pixfmt path; the gradient
colour itself is still the backend's Go-side computation, which already matches
the port within the gradient tolerance. The `compositing_gradient` scene
(gradient circle under `src` over an opaque block) covers this.

The vector fill/stroke and gradient paths support the **full AGG operator set**,
not just the original five. `map_comp_op` translates every `agg.BlendMode` 1:1
onto AGG 2.6's `comp_op_e` (the Porter-Duff operators plus the separable blend
modes — `multiply`, `screen`, `overlay`, `darken`, `lighten`, `color-dodge`,
`color-burn`, `hard-light`, `soft-light`, `difference`, `exclusion`; `add` maps
to `comp_op_plus`), and `g_comp_op_func[op]` evaluates them in premultiplied
space exactly as the port's `CompositeBlender` does. `compositing_multiply` is
byte-exact cross-backend; `TestCPPExtendedBlendModesRenderWithAggReal` exercises
a representative spread.

Operators that produce a **translucent result over an opaque destination** (e.g.
`xor`, `dst-out`) are intentionally not in the strict corpus. The C++ backend is
AGG-faithful there — the straight-alpha adaptor demultiplies the translucent
result back to straight on write, locked by
`TestCPPXorBlendIsAGGFaithfulWithAggReal` — but the Go _port_ currently leaves
premultiplied data in its straight buffer for such results (a separate port-side
bug tracked in PLAN.md §5.5), so a cross-backend corpus scene would fail on the
port, not the C++ engine. `compositing_src`/`srcover`/`clear` stayed byte-exact
only because src-over-opaque yields an opaque result, where premultiplied equals
straight.

**Image and text** draw now honour the full operator set too. The scaled/quad
image blits composite each covered pixel through `comp_op_adaptor_rgba_plain`
(the same straight-alpha bridge, via the `blend_image_pixel` helper) with full
cover — their loops only visit pixels the image covers, so no untouched
background is disturbed. The text coverage layer routes clear/src and every
separable mode through `agg_go_cpp_image_composite_cover` with the glyph layer's
alpha as the per-pixel cover, confining the operator to the glyphs (a whole-layer
composite would let clear/src wipe the background; src-over/alpha/dst keep the
proven whole-layer path since they only touch covered pixels anyway). This
mirrors the port, whose `currentImageRenderer` composites image spans through the
comp-op base renderer (`renBaseCompPre`) for every non-`BlendAlpha` mode. The Go
gate is the single `requireBlendMode` (full enum) and the native gate is
`supported_comp_op_mode`. The strict `image_blend` scene (image over a colour
field under `multiply`) verifies cross-backend parity, and
`TestCPPImageDrawUnderExtendedBlendModeIsFaithfulWithAggReal`,
`TestCPPTextUnderExtendedBlendModePreservesBackgroundWithAggReal`, and
`TestCPPBackendExtendedBlendModeOnDrawImageQuad` lock the C++ side.

### Known Cross-Backend Divergences

None currently. The conformance suite retains a `knownDivergence` table (empty)
so a future partial feature can be logged as a tracked baseline rather than
gating CI; `cmd/engine-compare` emits per-scene diffs for triage either way.

### Capability-Gap Skips

Scenes that hit a typed `engine.ErrUnsupportedCapability` on a backend are
**skipped**, not failed — this is the correct response to a documented gap.
There are currently no image/transform scenes in this state: `image_affine`
(scaled image draw under an active transform) is now strict on `cpp` — see
"Transformed image draw" below.

### Dashed strokes

`AddDash(dashLen, gapLen)`, `RemoveAllDashes()`, `DashStart(offset)`, and
`GetDashStart()` are now on `engine.Context`. The port delegates to the root
`Agg2D` dash API; the real C++ backend feeds `agg::conv_dash` into
`agg::conv_stroke`. The `dashed_stroke` corpus scene compares within the
solid/path envelope (observed ~1.3% edge-AA disagreement). Dash lengths are
measured in user space by the port and in device space by the C++ backend (it
dashes the pre-transformed path), so the corpus scene uses no active transform;
dashed strokes under a non-identity transform are not yet a guaranteed parity
case (tracked in PLAN.md §5.5).

### Transformed image draw

`DrawImage`, `DrawImageScaled`, and `DrawImageRegion` honour the active
transform on the C++ backend: the destination rectangle's corners are mapped
through the current matrix and the image is blitted into the resulting
parallelogram via the quad path, mirroring the port's `renderImage` (which
composes the CTM into the source→destination parallelogram matrix). Sampling on
the C++ side is nearest-neighbour (the CPU quad blit, compositing each covered
pixel through `blend_image_pixel`) while the port resamples with a bilinear image
filter, so the two diverge along rotated
edges within the documented `image_affine` envelope — not a geometry mismatch.

Getting this right required fixing the native matrix's composition order: its
`Translate`/`Rotate`/`Scale` had composed in the reverse order from
`agg::trans_affine` (the most recent call ended up innermost rather than
outermost), which silently mis-placed every transformed draw relative to the
faithful port. The native ops now pre-multiply the primitive in output space,
matching `trans_affine` exactly; the order is guarded cross-backend by
`TestCPPTransformComposeOrderMatchesPortWithAggReal`. This fix also corrects
transformed **vector** rendering, which no corpus scene had been exercising.

### State and style readback

The facade v1 getter contract is symmetric with its setters. In addition to the
earlier subset (`GetBlendMode`, `GetFillEvenOdd`, `GetFillGradientType`,
`GetStrokeGradientType`, `GetClipBox`, `GetTextHints`, `MeasureText`,
`GetTextBounds`), it now exposes:

- `GetFillColor()` / `GetStrokeColor()` — current solid fill/stroke colors.
- `GetLineWidth()` / `GetLineCap()` / `GetLineJoin()` — current stroke style.
- `GetTransform()` — the cumulative affine transform as an
  `agg.Transformations` value in AGG order (`sx, shy, shx, sy, tx, ty`),
  mirroring `GetClipBox`'s value-return style.

Both backends return identical results: the port delegates to the root
`agg.Context`/`Agg2D` getters, and the C++ backend returns its stored Go-side
style state and reads the transform back from the native `agg`-order matrix via
a dedicated `store` bridge call (so it reflects the actual native CTM rather
than a Go-side mirror). These getters were the open §5.4 v1 decisions: both the
transform-matrix readback and the style/state getters are deliberately **in
scope** for v1 because they are trivially backend-neutral and round-trip exactly
on both engines.
