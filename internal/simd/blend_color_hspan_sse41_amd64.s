//go:build amd64 && !purego

#include "textflag.h"

// aByteFF16: 16-byte mask [0,0,0,0xFF] repeated 4× — used to OR src A bytes → 255
// so that lerp(dst_a, 255, alpha) is computed via the standard PMAXUW/PMINUW path.
DATA ·aByteFF16+0(SB)/4,  $0xFF000000
DATA ·aByteFF16+4(SB)/4,  $0xFF000000
DATA ·aByteFF16+8(SB)/4,  $0xFF000000
DATA ·aByteFF16+12(SB)/4, $0xFF000000
GLOBL ·aByteFF16(SB), RODATA|NOPTR, $16

// func blendColorHspanRGBASSE41Asm(dst []byte, srcColors []byte, covers []byte, count int)
//
// Blends per-pixel RGBA source colors with per-pixel coverage into packed RGBA dst.
// srcColors: flat RGBA bytes, 4 per pixel.  covers: one cover byte per pixel (never nil).
// premulSrc is NOT handled here — the Go caller routes premul to the generic.
//
// Stack layout (ABI0, amd64):
//   dst_base:       0(FP)  8 bytes
//   dst_len:        8(FP)  8 bytes
//   dst_cap:       16(FP)  8 bytes
//   srcColors_base:24(FP)  8 bytes
//   srcColors_len: 32(FP)  8 bytes
//   srcColors_cap: 40(FP)  8 bytes
//   covers_base:   48(FP)  8 bytes
//   covers_len:    56(FP)  8 bytes
//   covers_cap:    64(FP)  8 bytes
//   count:         72(FP)  8 bytes
//   total args:    80 bytes
TEXT ·blendColorHspanRGBASSE41Asm(SB), NOSPLIT, $0-80
	MOVQ dst_base+0(FP),       DI  // dst ptr
	MOVQ srcColors_base+24(FP), SI  // src ptr
	MOVQ covers_base+48(FP),   DX  // covers ptr
	MOVQ count+72(FP),         CX  // count

	TESTQ CX, CX
	JLE   done

	// Load constants into registers once.
	MOVOU ·aByteFF16(SB), X10  // [0,0,0,0xFF]*4 — to OR src A bytes → 255
	MOVOU ·bias128W(SB), X14   // [128]*8 16-bit — rounding bias

	CMPQ CX, $2
	JB   tail

loop:
	// ── scalar: compute alpha0 = multiply(sa0, cover0) ──────────────────
	MOVBLZX 3(SI), R8    // R8 = sa0
	MOVBLZX (DX),  R10   // R10 = cover0
	IMULQ   R10, R8
	ADDQ    $128, R8
	MOVQ    R8, R12
	SHRQ    $8, R12
	ADDQ    R12, R8
	SHRQ    $8, R8      // R8 = alpha0

	// ── scalar: compute alpha1 = multiply(sa1, cover1) ──────────────────
	MOVBLZX 7(SI), R9    // R9 = sa1
	MOVBLZX 1(DX), R11   // R11 = cover1
	IMULQ   R11, R9
	ADDQ    $128, R9
	MOVQ    R9, R12
	SHRQ    $8, R12
	ADDQ    R12, R9
	SHRQ    $8, R9      // R9 = alpha1

	// ── build alpha word vector: [alpha0]*4, [alpha1]*4 ─────────────────
	MOVD       R8, X2
	PSHUFLW    $0, X2, X2      // X2 lo quad = [alpha0,alpha0,alpha0,alpha0]
	MOVD       R9, X3
	PSHUFLW    $0, X3, X3      // X3 lo quad = [alpha1,alpha1,alpha1,alpha1]
	PUNPCKLQDQ X3, X2          // X2 = [alpha0]*4 | [alpha1]*4  (16-bit words)

	// ── build src_opaque: [sr,sg,sb,255, sr1,sg1,sb1,255] in 16-bit ────
	MOVQ     (SI), X1           // X1 lo = [sr0,sg0,sb0,sa0, sr1,sg1,sb1,sa1] bytes
	POR      X10, X1            // bytes 3,7 → 0xFF
	PMOVZXBW X1, X1             // X1 = 8 × 16-bit words (src_opaque)

	// ── load dst ────────────────────────────────────────────────────────
	MOVQ     (DI), X0
	PMOVZXBW X0, X0             // X0 = 8 × 16-bit words (dst)

	// ── lerp: dst + (src − dst) * alpha / 256 via PMAXUW/PMINUW ────────
	MOVOU  X1, X3
	PMAXUW X0, X3   // X3 = max(src, dst)
	MOVOU  X1, X4
	PMINUW X0, X4   // X4 = min(src, dst)
	MOVOU  X3, X9   // save max for direction

	PSUBW  X4, X3   // X3 = |src − dst|
	PMULLW X2, X3   // X3 *= alpha (fits 16-bit)
	PADDW  X14, X3  // + 128
	MOVOU  X3, X6
	PSRLW  $8, X6
	PADDW  X6, X3
	PSRLW  $8, X3   // Knuth rounding: (t + (t>>8)) >> 8

	MOVOU X0, X7
	PADDW X3, X7    // dst + diff  (where src ≥ dst)
	MOVOU X0, X8
	PSUBW X3, X8    // dst − diff  (where src < dst)

	PCMPEQW X1, X9  // X9 = 0xFFFF where max==src, i.e., src ≥ dst
	MOVOU   X9, X12
	PAND    X7, X12  // select (dst+diff) where src ≥ dst
	PANDN   X8, X9   // select (dst-diff) where src < dst
	POR     X12, X9

	PACKUSWB X9, X9   // pack 16-bit→8-bit (low 8 bytes = result)
	MOVQ     X9, (DI) // store 2 pixels

	ADDQ $8, DI
	ADDQ $8, SI
	ADDQ $2, DX
	SUBQ $2, CX
	CMPQ CX, $2
	JGE  loop

tail:
	TESTQ CX, CX
	JLE   done

tail_loop:
	// ── scalar tail: blend 1 pixel ───────────────────────────────────────
	// alpha = multiply(sa, cover) — then lerp(dst_ch, src_ch, alpha) per channel.
	MOVBLZX 3(SI), R8   // R8 = sa
	MOVBLZX (DX),  R10  // R10 = cover

	IMULQ R10, R8
	ADDQ  $128, R8
	MOVQ  R8, R12
	SHRQ  $8, R12
	ADDQ  R12, R8
	SHRQ  $8, R8       // R8 = alpha

	// Skip pixel if alpha == 0.
	TESTQ R8, R8
	JZ    tail_next

	// --- R channel: lerp(dst[0], src_r, alpha) ---
	MOVBLZX (SI),  R9   // R9 = src_r
	MOVBLZX (DI),  R10  // R10 = dst_r
	CMPQ    R9, R10
	JAE     tail_r_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ R8, R11
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
	IMULQ R8, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX (DI), R10
	ADDQ    R11, R10
	MOVB    R10, (DI)

tail_g:
	// --- G channel: lerp(dst[1], src_g, alpha) ---
	MOVBLZX 1(SI), R9
	MOVBLZX 1(DI), R10
	CMPQ    R9, R10
	JAE     tail_g_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ R8, R11
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
	IMULQ R8, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 1(DI), R10
	ADDQ    R11, R10
	MOVB    R10, 1(DI)

tail_b:
	// --- B channel: lerp(dst[2], src_b, alpha) ---
	MOVBLZX 2(SI), R9
	MOVBLZX 2(DI), R10
	CMPQ    R9, R10
	JAE     tail_b_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ R8, R11
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
	IMULQ R8, R11
	ADDQ  $128, R11
	MOVQ  R11, R12
	SHRQ  $8, R12
	ADDQ  R12, R11
	SHRQ  $8, R11
	MOVBLZX 2(DI), R10
	ADDQ    R11, R10
	MOVB    R10, 2(DI)

tail_a:
	// --- A channel: lerp(dst[3], 255, alpha)  (src A treated as 255) ---
	MOVQ    $255, R9
	MOVBLZX 3(DI), R10
	SUBQ    R10, R9    // 255 − dst_a (always ≥ 0)
	MOVQ    R9, R11
	IMULQ   R8, R11
	ADDQ    $128, R11
	MOVQ    R11, R12
	SHRQ    $8, R12
	ADDQ    R12, R11
	SHRQ    $8, R11
	MOVBLZX 3(DI), R10
	ADDQ    R11, R10
	MOVB    R10, 3(DI)

tail_next:
	ADDQ $4, DI
	ADDQ $4, SI
	ADDQ $1, DX
	DECQ CX
	JNZ  tail_loop

done:
	RET
