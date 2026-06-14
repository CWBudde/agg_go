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

| Capability      | `port` | `cpp` stub (`agogo`) | `cpp` real (`agogo aggreal`) | Notes                                                                                                                                                         |
| --------------- | ------ | -------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `solid_style`   | yes    | unavailable          | yes                          | Fill and stroke colors work in both available engines.                                                                                                        |
| `path`          | yes    | unavailable          | yes                          | `MoveTo`, `LineTo`, `QuadTo`, `CubicTo`, `ClosePath`, fill, stroke, rectangle, and circle helpers are available.                                              |
| `transforms`    | yes    | unavailable          | yes                          | Translation, rotation, scale, and reset are implemented.                                                                                                      |
| `clip_box`      | yes    | unavailable          | yes                          | The current C++ backend clips fill, stroke, and image operations.                                                                                             |
| `compositing`   | yes    | unavailable          | partial                      | The current C++ backend supports `BlendAlpha`, `BlendClear`, `BlendSrc`, `BlendDst`, and `BlendSrcOver`. Other blend modes fail with typed capability errors. |
| `image_draw`    | yes    | unavailable          | partial                      | Copy, region draw, scaling, and quad mapping are implemented. `DrawImageRegion` with an active transform is still rejected as unsupported.                    |
| `image_export`  | yes    | unavailable          | yes                          | PNG, JPEG, `ToGoImage`, and `ToStandardImage` work through the facade.                                                                                        |
| `image_interop` | yes    | unavailable          | no                           | The C++ backend still rejects `Premultiply` and `Demultiply`.                                                                                                 |
| `gradients`     | yes    | unavailable          | yes                          | The current C++ subset applies fill and stroke gradients during actual rendering.                                                                             |
| `text`          | yes    | unavailable          | partial                      | Font loading, hinting, draw, measure, and bounds exist in the real-native path. Advanced text features are still out of scope.                                |
| `dashed_stroke` | yes    | unavailable          | yes                          | `AddDash`/`RemoveAllDashes`/`DashStart`/`GetDashStart` drive AGG `conv_dash` on both backends.                                                                |

## Intentional Gaps

- The `engine` facade is deliberately narrower than the root `agg` API.
- The current C++ backend is still partial even in `agogo aggreal`.
- Some C++ image/compositing paths still rely on local CPU helper logic instead
  of AGG-native implementations.
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

| Scene class                   | Tolerance | MaxDifferentRatio | Rationale                                                              |
| ----------------------------- | --------- | ----------------- | ---------------------------------------------------------------------- |
| solid / dashed / path / clip  | 2         | 0.025             | Edge-AA disagreement on ~1.5% of pixels; bulk identical within 2 LSB.  |
| linear / radial gradient      | 3         | 0.020             | Independent gradient interpolation rounding.                           |
| scaled image (`image_scaled`) | 4         | 0.080             | Independent samplers disagree along upscaled hard edges (~6%).         |
| text (`text_basic`)           | 8         | 0.100             | Native AGG vs Go-port FreeType AA/hinting; observed ~0.5% in practice. |

### Known Cross-Backend Divergences

These scenes render on both engines but differ beyond an AA-noise envelope
because the C++ backend's implementation is still partial (PLAN.md §5.5). The
conformance suite renders, measures, and **logs** them as tracked baselines but
does not fail; `cmd/engine-compare` emits their diffs for triage. Promote each
to a strict scene once the corresponding C++ parity work lands.

| Scene                 | Divergence                                                                                  |
| --------------------- | ------------------------------------------------------------------------------------------- |
| `compositing_srcover` | CPP blends translucent fills darker than the port (8-bit blend/premultiply rounding).       |
| `compositing_src`     | CPP applies the `Src` operator over the whole buffer, wiping untouched background to clear. |
| `compositing_clear`   | CPP `Clear` operator coverage differs from the port across the filled region.               |

### Capability-Gap Skips

Scenes that hit a typed `engine.ErrUnsupportedCapability` on a backend are
**skipped**, not failed — this is the correct response to a documented gap:

- `image_affine` (scaled image draw under an active transform) skips on `cpp`,
  which still rejects `DrawImageRegion` with an active transform.

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
