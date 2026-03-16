This directory stores bootstrap visual references copied from the current Go
golden tests.

Current contents:
- `go-golden/primitives/`: snapshot of `tests/visual/reference/primitives/`
  taken on 2026-03-16.

Scope:
- These images are the temporary parity target for visual comparison work.
- They are Go-generated references, not canonical C++ AGG outputs.
- The precompiled C++ examples under `../agg-2.6/build/examples/` remain the
  next source for canonical references when Phase 8.2 moves beyond bootstrap.

Rationale:
- Keep a stable, centralized copy of the golden-test corpus while the visual
  approval workflow is still being formalized.
- Avoid reusing transient files from `tests/visual/output/` as reference data.
