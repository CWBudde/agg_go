//go:build amd64 && !purego

#include "textflag.h"

// rgbMaskPremul: [0xFF,0xFF,0xFF,0x00] × 4 — keeps RGB, zeros alpha.
DATA ·rgbMaskPremul+0(SB)/4,  $0x00FFFFFF
DATA ·rgbMaskPremul+4(SB)/4,  $0x00FFFFFF
DATA ·rgbMaskPremul+8(SB)/4,  $0x00FFFFFF
DATA ·rgbMaskPremul+12(SB)/4, $0x00FFFFFF
GLOBL ·rgbMaskPremul(SB), RODATA|NOPTR, $16

// alphaMaskPremul: [0x00,0x00,0x00,0xFF] × 4 — keeps only alpha.
DATA ·alphaMaskPremul+0(SB)/4,  $0xFF000000
DATA ·alphaMaskPremul+4(SB)/4,  $0xFF000000
DATA ·alphaMaskPremul+8(SB)/4,  $0xFF000000
DATA ·alphaMaskPremul+12(SB)/4, $0xFF000000
GLOBL ·alphaMaskPremul(SB), RODATA|NOPTR, $16

// func premultiplyRGBASSE41Asm(buf []byte, count int)
//
// Premultiplies count tightly-packed RGBA pixels in buf in-place.
// buf is R,G,B,A byte order. For each pixel: R'=R*A/255, G'=G*A/255,
// B'=B*A/255, A'=A (unchanged). Uses the AGG rounding formula:
//   result = (ch*a + 128 + ((ch*a+128)>>8)) >> 8
//
// Processes 4 pixels per SIMD iteration; scalar tail for remainder.
//
// Stack layout (ABI0, amd64):
//   buf_base: 0(FP)   8 bytes
//   buf_len:  8(FP)   8 bytes
//   buf_cap: 16(FP)   8 bytes
//   count:   24(FP)   8 bytes
TEXT ·premultiplyRGBASSE41Asm(SB), NOSPLIT, $0-32
	MOVQ buf_base+0(FP), DI
	MOVQ count+24(FP), CX

	TESTQ CX, CX
	JLE   done

	MOVOU ·bias128W(SB), X14      // [128]*8 rounding bias (16-bit words)
	MOVOU ·rgbMaskPremul(SB), X12 // RGB mask: keeps R,G,B, zeros A
	MOVOU ·alphaMaskPremul(SB), X13 // alpha mask: zeros R,G,B, keeps A

	CMPQ CX, $4
	JB   tail

loop:
	// Load 4 pixels (16 bytes).
	MOVOU (DI), X0
	// Save original pixels for alpha restoration at the end.
	MOVOU X0, X9

	// ── Low 2 pixels (bytes 0–7) ─────────────────────────────────────────
	// PMOVZXBW with XMM src reads the low 8 bytes of X0.
	PMOVZXBW X0, X1 // X1 = R0,G0,B0,A0,R1,G1,B1,A1 as 16-bit words

	// Broadcast alpha for each pixel:
	//   PSHUFLW $0xFF  →  lo 4 words = A0,A0,A0,A0  (word 3 → all)
	//   PSHUFHW $0xFF  →  hi 4 words = A1,A1,A1,A1  (word 7 → all)
	MOVOU      X1, X2
	PSHUFLW    $0xFF, X2, X2
	PSHUFHW    $0xFF, X2, X2

	// Multiply all 8 channels by their respective alpha (16-bit product fits
	// since 255×255 = 65025 < 65536).
	PMULLW X2, X1

	// AGG rounding: (t + 128 + ((t+128)>>8)) >> 8
	PADDW X14, X1
	MOVOU X1, X3
	PSRLW $8, X3
	PADDW X3, X1
	PSRLW $8, X1 // X1 = premultiplied lo 2 pixels (16-bit words, values 0–255)

	// ── High 2 pixels (bytes 8–15) ───────────────────────────────────────
	MOVOU    X0, X4
	PSRLDQ   $8, X4   // shift bytes 8–15 down to positions 0–7
	PMOVZXBW X4, X4   // X4 = R2,G2,B2,A2,R3,G3,B3,A3 as 16-bit words

	MOVOU      X4, X5
	PSHUFLW    $0xFF, X5, X5
	PSHUFHW    $0xFF, X5, X5

	PMULLW X5, X4
	PADDW  X14, X4
	MOVOU  X4, X3
	PSRLW  $8, X3
	PADDW  X3, X4
	PSRLW  $8, X4 // X4 = premultiplied hi 2 pixels (16-bit words)

	// Pack both halves back to bytes (PACKUSWB: X1→lo 8 bytes, X4→hi 8 bytes).
	PACKUSWB X4, X1

	// Restore original alpha channels (alpha was incorrectly set to A*A above).
	PAND X12, X1  // zero alpha byte positions in premultiplied result
	PAND X13, X9  // keep only original alpha bytes
	POR  X9, X1   // combine: premultiplied RGB + original A

	MOVOU X1, (DI) // store 4 pixels
	ADDQ  $16, DI
	SUBQ  $4, CX
	CMPQ  CX, $4
	JGE   loop

tail:
	TESTQ CX, CX
	JLE   done

tail_loop:
	// Scalar tail: process 1 pixel using the same AGG multiply formula.
	MOVBLZX 3(DI), R8  // R8 = alpha
	CMPQ    R8, $255
	JE      tail_next   // fully opaque → skip
	TESTQ   R8, R8
	JZ      tail_zero   // alpha == 0 → zero RGB

	// R' = (R*a + 128 + ((R*a+128)>>8)) >> 8
	MOVBLZX (DI), R9
	IMULQ   R8, R9
	ADDQ    $128, R9
	MOVQ    R9, R10
	SHRQ    $8, R10
	ADDQ    R10, R9
	SHRQ    $8, R9
	MOVB    R9, (DI)

	// G'
	MOVBLZX 1(DI), R9
	IMULQ   R8, R9
	ADDQ    $128, R9
	MOVQ    R9, R10
	SHRQ    $8, R10
	ADDQ    R10, R9
	SHRQ    $8, R9
	MOVB    R9, 1(DI)

	// B'
	MOVBLZX 2(DI), R9
	IMULQ   R8, R9
	ADDQ    $128, R9
	MOVQ    R9, R10
	SHRQ    $8, R10
	ADDQ    R10, R9
	SHRQ    $8, R9
	MOVB    R9, 2(DI)
	JMP     tail_next

tail_zero:
	MOVB $0, (DI)
	MOVB $0, 1(DI)
	MOVB $0, 2(DI)

tail_next:
	ADDQ $4, DI
	DECQ CX
	JNZ  tail_loop

done:
	RET
