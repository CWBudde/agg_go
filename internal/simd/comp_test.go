package simd

import "testing"

// ─────────────────────────────────────────────────────────────────────────────
// Phase 8.2 — Composite blend mode tests
// ─────────────────────────────────────────────────────────────────────────────

// refCompSrcOver is a float64 reference for Porter-Duff SrcOver on premultiplied dst.
// (r,g,b,a) is straight-alpha src; cv is per-pixel coverage.
func refCompSrcOver(dca [4]byte, r, g, b, a, cv byte) [4]byte {
	sa := float64(rgba8Multiply(a, cv)) / 255.0
	scar := (float64(r) / 255.0) * sa
	scag := (float64(g) / 255.0) * sa
	scab := (float64(b) / 255.0) * sa
	dc := [4]float64{float64(dca[0]) / 255, float64(dca[1]) / 255, float64(dca[2]) / 255, float64(dca[3]) / 255}
	out := [4]float64{
		scar + dc[0]*(1-sa),
		scag + dc[1]*(1-sa),
		scab + dc[2]*(1-sa),
		sa + dc[3]*(1-sa),
	}
	var res [4]byte
	for i, v := range out {
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		res[i] = byte(v*255 + 0.5)
	}
	return res
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestCompSrcOverHspanRGBAComprehensive verifies CompSrcOverHspanRGBA across all
// SIMD implementation tiers for a range of pixel counts and coverage patterns.
func TestCompSrcOverHspanRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	type scenario struct {
		name       string
		dst        []byte
		r, g, b, a byte
		covers     []byte
	}

	mkDst := func(pixels [][4]byte) []byte {
		buf := make([]byte, len(pixels)*4)
		for i, p := range pixels {
			copy(buf[i*4:], p[:])
		}
		return buf
	}

	scenarios := []scenario{
		{
			name: "transparent_src",
			dst:  mkDst([][4]byte{{128, 64, 32, 200}}),
			r:    255, g: 0, b: 0, a: 0,
		},
		{
			name: "opaque_src_over_transparent_dst",
			dst:  mkDst([][4]byte{{0, 0, 0, 0}}),
			r:    200, g: 100, b: 50, a: 255,
		},
		{
			name: "opaque_src_over_opaque_dst",
			dst:  mkDst([][4]byte{{80, 80, 80, 255}}),
			r:    200, g: 100, b: 50, a: 255,
		},
		{
			name: "half_alpha_2px",
			dst:  mkDst([][4]byte{{100, 150, 200, 200}, {50, 50, 50, 128}}),
			r:    200, g: 50, b: 100, a: 128,
		},
		{
			name: "4_pixels_simd_boundary",
			dst:  mkDst([][4]byte{{10, 20, 30, 40}, {50, 60, 70, 80}, {90, 100, 110, 120}, {130, 140, 150, 160}}),
			r:    200, g: 100, b: 50, a: 180,
		},
		{
			name: "5_pixels_tail",
			dst:  mkDst([][4]byte{{10, 20, 30, 40}, {50, 60, 70, 80}, {90, 100, 110, 120}, {130, 140, 150, 160}, {170, 180, 190, 200}}),
			r:    100, g: 150, b: 200, a: 200,
		},
		{
			name: "variable_covers",
			dst:  mkDst([][4]byte{{100, 100, 100, 255}, {100, 100, 100, 255}, {100, 100, 100, 255}}),
			r:    200, g: 50, b: 50, a: 200,
			covers: []byte{0, 128, 255},
		},
		{
			name: "zero_cover_skips_pixel",
			dst:  mkDst([][4]byte{{200, 200, 200, 255}}),
			r:    0, g: 0, b: 0, a: 255,
			covers: []byte{0},
		},
		{
			name: "large_span_64px",
			dst: func() []byte {
				b := make([]byte, 64*4)
				for i := range 64 {
					b[i*4+0] = byte(i * 3)
					b[i*4+1] = byte(i * 7)
					b[i*4+2] = byte(i * 11)
					b[i*4+3] = byte(i * 4)
				}
				return b
			}(),
			r: 180, g: 90, b: 45, a: 200,
		},
		{
			name: "odd_count_13px",
			dst: func() []byte {
				b := make([]byte, 13*4)
				for i := range 13 {
					b[i*4+0] = byte(i * 17 % 256)
					b[i*4+1] = byte(i * 31 % 256)
					b[i*4+2] = byte(i * 53 % 256)
					b[i*4+3] = byte(i * 7 % 256)
				}
				return b
			}(),
			r: 200, g: 100, b: 150, a: 220,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			n := len(sc.dst) / 4
			// Build float64 reference.
			ref := append([]byte(nil), sc.dst...)
			for i := range n {
				cv := byte(255)
				if sc.covers != nil && i < len(sc.covers) {
					cv = sc.covers[i]
				}
				d := [4]byte{ref[i*4], ref[i*4+1], ref[i*4+2], ref[i*4+3]}
				out := refCompSrcOver(d, sc.r, sc.g, sc.b, sc.a, cv)
				copy(ref[i*4:], out[:])
			}

			for _, impl := range allImplCases() {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					dst := append([]byte(nil), sc.dst...)
					CompSrcOverHspanRGBA(dst, sc.covers, sc.r, sc.g, sc.b, sc.a, n)

					for i := range n {
						p := i * 4
						for ch := range 4 {
							got := int(dst[p+ch])
							want := int(ref[p+ch])
							if absInt(got-want) > 1 {
								t.Fatalf("pixel %d ch %d: got %d, want %d (±1 allowed)", i, ch, got, want)
							}
						}
					}
				})
			}
		})
	}
}

// TestCompOtherOpsGeneric verifies DstOver, SrcIn, DstIn, SrcOut, DstOut, Xor
// using float64 reference values.
func TestCompOtherOpsGeneric(t *testing.T) {
	t.Parallel()

	toF := func(v byte) float64 { return float64(v) / 255 }
	from8 := func(v float64) byte {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 255
		}
		return byte(v*255 + 0.5)
	}

	type testOp struct {
		name string
		fn   func(dst, covers []byte, r, g, b, a uint8, count int)
		ref  func(d [4]byte, r, g, b, a byte) [4]byte
	}

	ops := []testOp{
		{
			name: "DstOver",
			fn:   CompDstOverHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				return [4]byte{
					from8(toF(d[0]) + toF(r)*sa*(1-da)),
					from8(toF(d[1]) + toF(g)*sa*(1-da)),
					from8(toF(d[2]) + toF(b)*sa*(1-da)),
					from8(da + sa*(1-da)),
				}
			},
		},
		{
			name: "SrcIn",
			fn:   CompSrcInHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				return [4]byte{
					from8(toF(r) * sa * da),
					from8(toF(g) * sa * da),
					from8(toF(b) * sa * da),
					from8(sa * da),
				}
			},
		},
		{
			name: "DstIn",
			fn:   CompDstInHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				sa := toF(a)
				return [4]byte{
					from8(toF(d[0]) * sa),
					from8(toF(d[1]) * sa),
					from8(toF(d[2]) * sa),
					from8(toF(d[3]) * sa),
				}
			},
		},
		{
			name: "SrcOut",
			fn:   CompSrcOutHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				return [4]byte{
					from8(toF(r) * sa * (1 - da)),
					from8(toF(g) * sa * (1 - da)),
					from8(toF(b) * sa * (1 - da)),
					from8(sa * (1 - da)),
				}
			},
		},
		{
			name: "DstOut",
			fn:   CompDstOutHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				sa := toF(a)
				return [4]byte{
					from8(toF(d[0]) * (1 - sa)),
					from8(toF(d[1]) * (1 - sa)),
					from8(toF(d[2]) * (1 - sa)),
					from8(toF(d[3]) * (1 - sa)),
				}
			},
		},
		{
			name: "Xor",
			fn:   CompXorHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				scar, scag, scab := toF(r)*sa, toF(g)*sa, toF(b)*sa
				return [4]byte{
					from8(scar*(1-da) + toF(d[0])*(1-sa)),
					from8(scag*(1-da) + toF(d[1])*(1-sa)),
					from8(scab*(1-da) + toF(d[2])*(1-sa)),
					from8(sa + da - 2*sa*da),
				}
			},
		},
	}

	dstPixels := [][4]byte{
		{0, 0, 0, 0},
		{255, 255, 255, 255},
		{100, 150, 200, 200},
		{50, 50, 50, 128},
	}
	srcs := [][4]byte{
		{200, 100, 50, 255},
		{200, 100, 50, 128},
		{200, 100, 50, 0},
	}

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			for _, dp := range dstPixels {
				for _, sp := range srcs {
					dst := []byte{dp[0], dp[1], dp[2], dp[3]}
					want := op.ref(dp, sp[0], sp[1], sp[2], sp[3])
					op.fn(dst, nil, sp[0], sp[1], sp[2], sp[3], 1)
					for ch := range 4 {
						got := int(dst[ch])
						wantV := int(want[ch])
						if absInt(got-wantV) > 1 {
							t.Errorf("dst=%v src=%v ch=%d: got %d, want %d",
								dp, sp, ch, got, wantV)
						}
					}
				}
			}
		})
	}
}

// TestCompClearHspanRGBA verifies that Clear zeroes all pixels in the span.
func TestCompClearHspanRGBA(t *testing.T) {
	t.Parallel()

	dst := []byte{255, 128, 64, 200, 100, 50, 25, 180}
	CompClearHspanRGBA(dst, 2)
	for i, v := range dst {
		if v != 0 {
			t.Fatalf("byte %d not zeroed: got %d", i, v)
		}
	}

	// Partial clear — second pixel must be untouched.
	dst2 := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	CompClearHspanRGBA(dst2, 1)
	for i := range 4 {
		if dst2[i] != 0 {
			t.Fatalf("pixel 0 byte %d not cleared: got %d", i, dst2[i])
		}
	}
	if dst2[4] != 5 {
		t.Fatalf("second pixel modified: dst2[4]=%d want 5", dst2[4])
	}
}
