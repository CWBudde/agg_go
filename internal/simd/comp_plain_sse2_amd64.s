//go:build amd64 && !purego

#include "textflag.h"

// func compSrcOverPlainStraightHspanRGBASSE2Asm(dst []byte, src, is1 *[4]float64, count int)
//
// SSE2 fallback tier for the straight-alpha composite blender's
// uniform-coverage Porter-Duff SrcOver span, for pre-AVX2 amd64 machines. It is
// the exact SSE2 mirror of compSrcOverPlainStraightHspanRGBAAVX2Asm and is
// BIT-FOR-BIT identical to it (and to the scalar bridge in
// CompositeBlenderPlain.BlendSolidSpanStraight): float64 throughout, the same
// DIVPD divisions, and CVTTPD2PL (truncate toward zero) to match Go's
// uint8(v*255.0+0.5) after clamping to [0,1].
//
// AVX2 packs all four channels into one 256-bit register; SSE2 has only 128-bit
// registers (2 float64), so each pixel is carried in TWO xmm registers:
//   lo = {r, g}   hi = {b, a}
// Every elementwise op runs on both halves, so the per-lane arithmetic — and
// therefore the IEEE-754 result — is identical to the AVX2 single-register form.
//
// Per pixel (lane k <-> dst byte k):
//   vn  = {dr,dg,db,da} / 255                       (premultiply, divide #1)
//   d   = vn * {da,da,da,1}                          (straight -> premultiplied dst)
//   res = src + d*is1                                (SrcOver in premult space)
//   out = res / {resa,resa,resa,1}                   (demultiply, divide #2)
//   store to8(clamp(out,0,1)*255 + 0.5)
//
// src holds the premultiplied source {Sca,Sca,Sca,Sa} in DST BYTE ORDER and is1
// holds {1-Sa,...}; both are caller-built float64[4]. Alpha must be at byte 3
// (RGBA or BGRA): SrcOver is symmetric across the three colour lanes, so any
// colour permutation is handled by the caller placing src in byte order, while
// the alpha lane is fixed at lane 3 (the high lane of the hi register).
//
// This kernel is purely legacy-SSE encoded (no VEX), so no VZEROUPPER is
// required. The 4-byte pixel load/store uses MOVL with an xmm operand, which is
// Go amd64's spelling for the 32-bit MOVD (66 0F 6E / 66 0F 7E); plain MOVD here
// is Go's 64-bit move and would over-read/over-write 8 bytes per pixel.
//
// Stack layout (ABI0, amd64), identical to the AVX2 entry point:
//   dst_base:  0(FP)   8 bytes
//   dst_len:   8(FP)   8 bytes
//   dst_cap:  16(FP)   8 bytes
//   src:      24(FP)   8 bytes (*[4]float64)
//   is1:      32(FP)   8 bytes (*[4]float64)
//   count:    40(FP)   8 bytes (int)
//   total:    48 bytes
TEXT ·compSrcOverPlainStraightHspanRGBASSE2Asm(SB), NOSPLIT, $0-48
	MOVQ dst_base+0(FP), DI
	MOVQ src+24(FP), SI
	MOVQ is1+32(FP), DX
	MOVQ count+40(FP), CX

	TESTQ CX, CX
	JLE   done

	MOVUPD (SI), X1             // X1 = src lo {s0,s1}
	MOVUPD 16(SI), X2           // X2 = src hi {s2,sa}
	MOVUPD (DX), X7             // X7 = is1 {1-Sa,1-Sa} (uniform; same for both halves)
	MOVUPD ·compF255(SB), X3    // X3 = {255,255}
	MOVUPD ·compF1(SB), X4      // X4 = {1,1}
	MOVUPD ·compFHalf(SB), X5   // X5 = {0.5,0.5}
	PXOR   X6, X6               // X6 = {0,0}

loop:
	// Load 4 straight dst bytes -> 4 int32 -> {dr,dg} lo + {db,da} hi (float64).
	MOVL      (DI), X10
	PXOR      X11, X11
	PUNPCKLBW X11, X10          // zero-extend bytes -> 4 words {dr,dg,db,da}
	PUNPCKLWL X11, X10          // zero-extend words -> 4 dwords {dr,dg,db,da}
	CVTPL2PD  X10, X0           // X0 = {dr,dg} (float64)
	PSHUFD    $0xEE, X10, X9    // X9 dwords = {db,da,db,da}
	CVTPL2PD  X9, X8            // X8 = {db,da} (float64)

	// Premultiply: vn = bytes/255 (divide #1).
	DIVPD X3, X0               // X0 = {dr,dg}/255
	DIVPD X3, X8               // X8 = {db,da}/255

	// d = vn * {da,da,da,1}. da = vn lane 3 = high double of X8.
	MOVAPD X8, X9
	SHUFPD $3, X9, X9          // X9 = {da,da}      (lo multiplier)
	MOVAPD X4, X11             // X11 = {1,1}
	MOVSD  X9, X11             // X11 = {da,1}      (hi multiplier; high lane kept = 1)
	MULPD  X9, X0              // X0 = {dr,dg}/255 * da          = d lo
	MULPD  X11, X8             // X8 = {db/255*da, da}           = d hi

	// SrcOver: res = src + d*is1 (no FMA: separate multiply then add).
	MULPD X7, X0               // d lo * is1
	MULPD X7, X8               // d hi * is1
	ADDPD X1, X0               // X0 = src lo + d lo*is1 = res lo {res_r,res_g}
	ADDPD X2, X8               // X8 = src hi + d hi*is1 = res hi {res_b,res_a}

	// Demultiply: out = res / {resa,resa,resa,1}. resa = res lane 3 = high of X8,
	// always > 0 (resa = Sa + Da*(1-Sa) >= Sa > 0), so no zero guard needed.
	MOVAPD X8, X9
	SHUFPD $3, X9, X9          // X9 = {resa,resa}  (lo divisor)
	MOVAPD X4, X11             // X11 = {1,1}
	MOVSD  X9, X11             // X11 = {resa,1}    (hi divisor)
	DIVPD  X9, X0              // X0 = res lo / {resa,resa}
	DIVPD  X11, X8             // X8 = res hi / {resa,1} = {res_b/resa, res_a}

	// to8: clamp to [0,1], *255, +0.5, truncate toward zero.
	MAXPD X6, X0
	MAXPD X6, X8
	MINPD X4, X0
	MINPD X4, X8
	MULPD X3, X0
	MULPD X3, X8
	ADDPD X5, X0
	ADDPD X5, X8

	CVTTPD2PL X0, X0           // X0 low dwords = {ir,ig} (int32, high quad zeroed)
	CVTTPD2PL X8, X8           // X8 low dwords = {ib,ia}

	PUNPCKLQDQ X8, X0          // X0 = {ir,ig,ib,ia} (4 x int32)
	PACKSSLW   X0, X0          // int32 -> int16; values in [0,255] so signed pack is exact
	PACKUSWB   X0, X0          // int16 -> uint8 (unsigned saturate); low 4 bytes valid
	MOVL       X0, (DI)        // store 4 straight bytes

	ADDQ $4, DI
	DECQ CX
	JNZ  loop

done:
	RET
