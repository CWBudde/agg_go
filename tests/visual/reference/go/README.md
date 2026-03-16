This directory stores screenshots rendered by the Go demo ports.

These PNGs are generated artifacts, not golden references. The demo-parity test
regenerates them on each run, and they are intentionally ignored by Git.

Capture details:

- Captured on 2026-03-16.
- Demos were rendered through the default headless `demorunner` path, which
  renders one frame to PNG and exits.
- Most demos were run from their own example directories.
- `image1` and `image_alpha` were run from the repository root because their
  current asset-loading path expects `examples/shared/art/...` relative to the
  process working directory.

Current scope:

- 40 example screenshots imported successfully.
- The set mirrors the current initial C++ screenshot corpus under
  `tests/visual/reference/cpp/examples/`.
- These files are the current Go-port outputs for direct visual comparison
  against the C++ captures.

Next work:

- Compare these images directly against the C++ corpus and record scenario
  mappings.
- Normalize demo asset loading so examples like `image1` and `image_alpha` do
  not depend on repository-root execution.
