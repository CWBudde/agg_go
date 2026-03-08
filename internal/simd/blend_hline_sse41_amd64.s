//go:build amd64 && !purego

#include "textflag.h"

// func blendHlineRGBASSE41Asm(dst []byte, pixelOpaque uint32, alpha uint8, count int)
//
// Blends a solid RGBA color with uniform coverage alpha into packed RGBA pixels.
// pixelOpaque carries src R,G,B,0xFF (A=255 so the lerp formula handles dst_a correctly).
// alpha = rgba8Multiply(src_a, cover) — precomputed by the Go caller.
//
// Stack layout (ABI0, amd64):
//   dst_base:    0(FP)  8 bytes
//   dst_len:     8(FP)  8 bytes
//   dst_cap:    16(FP)  8 bytes
//   pixelOpaque:24(FP)  4 bytes (uint32)
//   alpha:      28(FP)  1 byte  (uint8)
//   [pad:       29-31]
//   count:      32(FP)  8 bytes (int)
//   total args: 40 bytes
TEXT ·blendHlineRGBASSE41Asm(SB), NOSPLIT, $0-40
	MOVQ dst_base+0(FP), DI
	MOVL pixelOpaque+24(FP), AX
	MOVBLZX alpha+28(FP), BX
	MOVQ count+32(FP), CX

	TESTQ CX, CX
	JLE  done

	PXOR X15, X15

	// Build src word vector for 2 pixels: [r,g,b,255, r,g,b,255] in 16-bit words.
	MOVD     AX, X10
	PSHUFD   $0, X10, X10  // broadcast pixelOpaque dword to all 4 dwords
	PMOVZXBW X10, X10      // unpack low 8 bytes → 8 x 16-bit words

	// Build alpha word vector: [alpha]*8 in 16-bit words.
	MOVD       BX, X11
	PSHUFLW    $0, X11, X11
	PUNPCKLQDQ X11, X11

	MOVOU ·bias128W(SB), X14

	CMPQ CX, $2
	JB   tail

loop:
	// Load 2 dst pixels (8 bytes) and unpack to 16-bit words.
	MOVQ     (DI), X0
	PMOVZXBW X0, X0

	// Compute lerp: dst + (src - dst) * alpha / 256.
	// Use PMAXUW/PMINUW to avoid signed-word overflow and track direction.
	MOVOU  X10, X3
	PMAXUW X0, X3   // X3 = max(src, dst)
	MOVOU  X10, X4
	PMINUW X0, X4   // X4 = min(src, dst)
	MOVOU  X3, X9   // X9 = max, for direction test

	PSUBW X4, X3    // X3 = |src - dst|
	PMULLW X11, X3  // X3 *= alpha  (low 16 bits, values fit in 16-bit)
	PADDW  X14, X3  // + 128 (Knuth rounding bias)
	MOVOU  X3, X6
	PSRLW  $8, X6
	PADDW  X6, X3
	PSRLW  $8, X3   // X3 = (t + (t>>8)) >> 8

	// Reconstruct: add diff when src >= dst, subtract when src < dst.
	MOVOU X0, X7
	PADDW X3, X7    // X7 = dst + diff
	MOVOU X0, X8
	PSUBW X3, X8    // X8 = dst - diff

	PCMPEQW X10, X9  // X9 = 0xFFFF where max == src (i.e., src >= dst)
	MOVOU   X9, X12
	PAND    X7, X12  // select (dst+diff) where src >= dst
	PANDN   X8, X9   // select (dst-diff) where src < dst
	POR     X12, X9

	PACKUSWB X9, X9   // pack to 8 bytes (uses both halves, low half = our result)
	MOVQ     X9, (DI) // store 2 pixels

	ADDQ $8, DI
	SUBQ $2, CX
	CMPQ CX, $2
	JGE  loop

tail:
	TESTQ CX, CX
	JLE   done

tail_loop:
	// Scalar lerp for 1 pixel: lerp(dst_ch, src_ch, alpha).
	// alpha is in BX; pixelOpaque bytes: [R,G,B,0xFF] in AX.

	// --- R channel (byte 0): lerp(dst[0], src_r, alpha) ---
	MOVBLZX AX, R9     // R9 = src_r
	MOVBLZX (DI), R10  // R10 = dst_r
	CMPQ    R9, R10
	JAE     tail_r_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX (DI), R10
	SUBQ    R11, R10
	MOVB    R10, (DI)
	JMP     tail_g

tail_r_add:
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX (DI), R10
	ADDQ    R11, R10
	MOVB    R10, (DI)

tail_g:
	// --- G channel (byte 1): lerp(dst[1], src_g, alpha) ---
	MOVQ    AX, R9
	SHRQ    $8, R9
	ANDQ    $0xFF, R9  // R9 = src_g
	MOVBLZX 1(DI), R10
	CMPQ    R9, R10
	JAE     tail_g_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 1(DI), R10
	SUBQ    R11, R10
	MOVB    R10, 1(DI)
	JMP     tail_b

tail_g_add:
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 1(DI), R10
	ADDQ    R11, R10
	MOVB    R10, 1(DI)

tail_b:
	// --- B channel (byte 2): lerp(dst[2], src_b, alpha) ---
	MOVQ    AX, R9
	SHRQ    $16, R9
	ANDQ    $0xFF, R9  // R9 = src_b
	MOVBLZX 2(DI), R10
	CMPQ    R9, R10
	JAE     tail_b_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 2(DI), R10
	SUBQ    R11, R10
	MOVB    R10, 2(DI)
	JMP     tail_a

tail_b_add:
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 2(DI), R10
	ADDQ    R11, R10
	MOVB    R10, 2(DI)

tail_a:
	// --- A channel (byte 3): lerp(dst[3], 255, alpha) ---
	// src_a = 255 (from pixelOpaque high byte), so 255 >= dst_a always → add path.
	MOVQ    $255, R9
	MOVBLZX 3(DI), R10
	SUBQ R10, R9   // 255 - dst_a
	MOVQ R9, R11
	IMULQ BX, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 3(DI), R10
	ADDQ    R11, R10
	MOVB    R10, 3(DI)

tail_next:
	ADDQ $4, DI
	DECQ CX
	JNZ  tail_loop

done:
	RET
