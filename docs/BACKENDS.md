# Backend Matrix

This repository now has one canonical codebase and two engine modes exposed
through the opt-in `engine` package:

- `port`: the default pure-Go renderer in the root `agg` package
- `cpp`: an optional in-repo C++ AGG-backed engine selected through
  `engine.Config{Kind: engine.CPP}`

The pure-Go root API remains the default path for existing users. Nothing in
this file changes the compatibility story for callers that stay on `agg`.

## Build Modes

| Engine | Build tags | Availability | Notes |
| --- | --- | --- | --- |
| `port` | none | available | Default pure-Go path. No cgo required. |
| `cpp` | none | unavailable | Returns a typed `engine.ErrUnavailable` because the `agogo` build tag is missing. |
| `cpp` | `agogo` with `CGO_ENABLED=0` | unavailable | Returns a typed `engine.ErrUnavailable` because cgo is disabled. |
| `cpp` | `agogo` | unavailable today | The current native layer compiles, but this build still rejects the backend when it is only the stub implementation. |
| `cpp` | `agogo aggreal` | available for current subset | Temporary real-native development path while dependency detection is being moved into this repo. |

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

| Capability | `port` | `cpp` stub (`agogo`) | `cpp` real (`agogo aggreal`) | Notes |
| --- | --- | --- | --- | --- |
| `solid_style` | yes | unavailable | yes | Fill and stroke colors work in both available engines. |
| `path` | yes | unavailable | yes | `MoveTo`, `LineTo`, `QuadTo`, `CubicTo`, `ClosePath`, fill, stroke, rectangle, and circle helpers are available. |
| `transforms` | yes | unavailable | yes | Translation, rotation, scale, and reset are implemented. |
| `clip_box` | yes | unavailable | yes | The current C++ backend clips fill, stroke, and image operations. |
| `compositing` | yes | unavailable | partial | The current C++ backend supports `BlendAlpha`, `BlendClear`, `BlendSrc`, `BlendDst`, and `BlendSrcOver`. Other blend modes fail with typed capability errors. |
| `image_draw` | yes | unavailable | partial | Copy, region draw, scaling, and quad mapping are implemented. `DrawImageRegion` with an active transform is still rejected as unsupported. |
| `image_export` | yes | unavailable | yes | PNG, JPEG, `ToGoImage`, and `ToStandardImage` work through the facade. |
| `image_interop` | yes | unavailable | no | The C++ backend still rejects `Premultiply` and `Demultiply`. |
| `gradients` | yes | unavailable | yes | The current C++ subset applies fill and stroke gradients during actual rendering. |
| `text` | yes | unavailable | partial | Font loading, hinting, draw, measure, and bounds exist in the real-native path. Advanced text features are still out of scope. |
| `dashed_stroke` | no | unavailable | no | Not part of the current shared facade subset yet. |

## Intentional Gaps

- The `engine` facade is deliberately narrower than the root `agg` API.
- Dashed strokes are not exposed through the shared facade yet.
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
