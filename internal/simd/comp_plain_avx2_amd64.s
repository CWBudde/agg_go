//go:build amd64 && !purego

#include "textflag.h"

// func compSrcOverPlainStraightHspanRGBAAVX2Asm(dst []byte, src, is1 *[4]float64, count int)
//
// SIMD (AVX2, float64) tier for the straight-alpha composite blender's
// uniform-coverage Porter-Duff SrcOver span. It is BIT-FOR-BIT identical to the
// scalar premultiply-on-read / op / demultiply-on-write bridge in
// CompositeBlenderPlain.BlendSolidSpanStraight: it uses float64 throughout (no
// FMA contraction — separate VMULPD/VADDPD), VDIVPD for the same divisions the
// scalar path performs, and VCVTTPD2DQ (truncate toward zero) to match Go's
// uint8(v*255.0+0.5) rounding after clamping the normalized value to [0,1].
//
// Per pixel (one pixel per 256-bit register, lane k <-> dst byte k):
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
// the alpha lane is fixed at lane 3 by the blend masks below.
//
// Stack layout (ABI0, amd64):
//   dst_base:  0(FP)   8 bytes
//   dst_len:   8(FP)   8 bytes
//   dst_cap:  16(FP)   8 bytes
//   src:      24(FP)   8 bytes (*[4]float64)
//   is1:      32(FP)   8 bytes (*[4]float64)
//   count:    40(FP)   8 bytes (int)
//   total:    48 bytes
TEXT ·compSrcOverPlainStraightHspanRGBAAVX2Asm(SB), NOSPLIT, $0-48
	MOVQ dst_base+0(FP), DI
	MOVQ src+24(FP), SI
	MOVQ is1+32(FP), DX
	MOVQ count+40(FP), CX

	TESTQ CX, CX
	JLE   done

	VMOVUPD (SI), Y1            // Y1 = premultiplied source {s0,s1,s2,sa}
	VMOVUPD (DX), Y2            // Y2 = {is1,is1,is1,is1} = 1-Sa broadcast
	VMOVUPD ·compF255(SB), Y3   // Y3 = {255,255,255,255}
	VMOVUPD ·compF1(SB), Y4     // Y4 = {1,1,1,1}
	VMOVUPD ·compFHalf(SB), Y5  // Y5 = {0.5,0.5,0.5,0.5}
	VXORPD  Y6, Y6, Y6          // Y6 = {0,0,0,0}

loop:
	// Load 4 straight dst bytes -> 4 int32 -> 4 float64.
	VMOVD     (DI), X10
	VPMOVZXBD X10, X10         // X10 = {dr,dg,db,da} (int32)
	VCVTDQ2PD X10, Y0          // Y0 = {dr,dg,db,da} (float64)

	// Premultiply: vn = bytes/255 (divide #1).
	VDIVPD Y3, Y0, Y0          // Y0 = Y0 / 255

	// d = vn * {da,da,da,1}, where da = vn lane 3.
	VPERMPD $0xFF, Y0, Y7      // Y7 = {da,da,da,da}
	VBLENDPD $0x8, Y4, Y7, Y7  // Y7 lane3 <- 1.0, lanes0-2 <- da
	VMULPD  Y7, Y0, Y0         // Y0 = d (premultiplied dst)

	// SrcOver: res = src + d*is1  (no FMA: separate multiply then add).
	VMULPD Y2, Y0, Y8          // Y8 = d * is1
	VADDPD Y1, Y8, Y0          // Y0 = src + d*is1 = res

	// Demultiply: out = res / {resa,resa,resa,1} (divide #2). resa = res lane 3,
	// always > 0 here (resa = Sa + Da*(1-Sa) >= Sa > 0), so no zero guard needed.
	VPERMPD  $0xFF, Y0, Y9     // Y9 = {resa,resa,resa,resa}
	VBLENDPD $0x8, Y4, Y9, Y9  // Y9 lane3 <- 1.0
	VDIVPD   Y9, Y0, Y0        // Y0 = out (normalized straight result)

	// to8: clamp to [0,1], *255, +0.5, truncate toward zero.
	VMAXPD Y6, Y0, Y0          // max(out, 0)
	VMINPD Y4, Y0, Y0          // min(out, 1)
	VMULPD Y3, Y0, Y0          // *255
	VADDPD Y5, Y0, Y0          // +0.5

	VCVTTPD2DQY Y0, X0         // 4 x float64 -> 4 x int32 (truncate; ymm source)
	VPACKUSDW  X0, X0, X0      // int32 -> uint16 (saturate); low 4 words valid
	VPACKUSWB  X0, X0, X0      // uint16 -> uint8 (saturate); low 4 bytes valid
	VMOVD      X0, (DI)        // store 4 straight bytes

	ADDQ $4, DI
	DECQ CX
	JNZ  loop

	VZEROUPPER

done:
	RET

// float64[4] constant vectors (16-byte alignment via GLOBL flags).
GLOBL ·compF255(SB), RODATA|NOPTR, $32
DATA  ·compF255+0(SB)/8, $0x406FE00000000000  // 255.0
DATA  ·compF255+8(SB)/8, $0x406FE00000000000
DATA  ·compF255+16(SB)/8, $0x406FE00000000000
DATA  ·compF255+24(SB)/8, $0x406FE00000000000

GLOBL ·compF1(SB), RODATA|NOPTR, $32
DATA  ·compF1+0(SB)/8, $0x3FF0000000000000  // 1.0
DATA  ·compF1+8(SB)/8, $0x3FF0000000000000
DATA  ·compF1+16(SB)/8, $0x3FF0000000000000
DATA  ·compF1+24(SB)/8, $0x3FF0000000000000

GLOBL ·compFHalf(SB), RODATA|NOPTR, $32
DATA  ·compFHalf+0(SB)/8, $0x3FE0000000000000  // 0.5
DATA  ·compFHalf+8(SB)/8, $0x3FE0000000000000
DATA  ·compFHalf+16(SB)/8, $0x3FE0000000000000
DATA  ·compFHalf+24(SB)/8, $0x3FE0000000000000
