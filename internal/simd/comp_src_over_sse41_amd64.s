//go:build amd64 && !purego

#include "textflag.h"

// func compSrcOverHspanRGBASSE41Asm(dst []byte, sca uint32, sa uint8, count int)
//
// Blends a premultiplied solid color (sca = [scar,scag,scab,sa] as RGBA bytes)
// over count premultiplied dst pixels using Porter-Duff SrcOver:
//
//   Dca' = rgba8Prelerp(Dca, Sca, Sa) = Dca + Sca - mul(Dca, Sa)
//   Da'  = rgba8Prelerp(Da,  Sa,  Sa) = Da  + Sa  - mul(Da,  Sa)
//
// This covers the uniform-coverage case (covers == nil). Variable coverage
// is handled by the Go scalar path.
//
// Processes 2 pixels (8 bytes) per SIMD iteration using SSE4.1.
//
// Stack layout (ABI0, amd64):
//   dst_base: 0(FP)   8 bytes
//   dst_len:  8(FP)   8 bytes
//   dst_cap: 16(FP)   8 bytes
//   sca:     24(FP)   4 bytes (uint32, R|G<<8|B<<16|A<<24)
//   sa:      28(FP)   1 byte  (uint8)
//   [pad:   29-31]
//   count:   32(FP)   8 bytes (int)
//   total:   40 bytes
TEXT ·compSrcOverHspanRGBASSE41Asm(SB), NOSPLIT, $0-40
	MOVQ    dst_base+0(FP), DI
	MOVL    sca+24(FP), AX        // AX = sca packed [scar,scag,scab,sa]
	MOVBLZX sa+28(FP), BX         // BX = sa (source alpha)
	MOVQ    count+32(FP), CX

	TESTQ CX, CX
	JLE   done

	MOVOU ·bias128W(SB), X14

	// Build Sca word vector for 2 pixels:
	//   X10 = [scar,scag,scab,sa, scar,scag,scab,sa] as 8 × 16-bit words
	MOVD     AX, X10
	PSHUFD   $0, X10, X10   // broadcast the 32-bit sca to all four dwords
	PMOVZXBW X10, X10       // unpack low 8 bytes to 8 × 16-bit words

	// Build Sa word vector: [sa]*8 as 16-bit words
	MOVD       BX, X11
	PSHUFLW    $0, X11, X11
	PUNPCKLQDQ X11, X11

	CMPQ CX, $2
	JB   tail

loop:
	// Load 2 dst pixels (8 bytes) and zero-extend to 16-bit words.
	MOVQ     (DI), X0
	PMOVZXBW X0, X0          // X0 = Dca (8 × 16-bit)

	// Compute mul(Dca, Sa): (Dca*Sa + 128 + ((Dca*Sa+128)>>8)) >> 8
	MOVOU  X0, X1
	PMULLW X11, X1           // X1 = Dca * Sa (fits in 16-bit: max 255*255=65025)
	PADDW  X14, X1           // + 128
	MOVOU  X1, X2
	PSRLW  $8, X2
	PADDW  X2, X1
	PSRLW  $8, X1            // X1 = mul(Dca, Sa)

	// Dca - mul(Dca, Sa)  [result ≥ 0 since mul(Dca,Sa) ≤ Dca]
	PSUBW X1, X0             // X0 = Dca - mul(Dca, Sa)

	// + Sca
	PADDW X10, X0            // X0 = Dca' = Dca + Sca - mul(Dca, Sa) ∈ [0,255]

	// Pack 8 × 16-bit words back to 8 bytes (saturates at 255).
	PACKUSWB X0, X0          // low 8 bytes = our 2-pixel result
	MOVQ     X0, (DI)        // store 2 pixels

	ADDQ $8, DI
	SUBQ $2, CX
	CMPQ CX, $2
	JGE  loop

tail:
	TESTQ CX, CX
	JLE   done

tail_loop:
	// Scalar: Dca'[ch] = Dca[ch] + Sca[ch] - mul(Dca[ch], sa)
	// AX = sca bytes; BX = sa

	// R: dst[0] + sca_r - mul(dst[0], sa)
	MOVBLZX (DI), R9          // Dca_r
	MOVQ    R9, R10
	IMULQ   BX, R10           // Dca_r * sa
	ADDQ    $128, R10
	MOVQ    R10, R11
	SHRQ    $8, R11
	ADDQ    R11, R10
	SHRQ    $8, R10            // R10 = mul(Dca_r, sa)
	SUBQ    R10, R9            // Dca_r - mul
	MOVBLZX AX, R10            // sca_r (byte 0)
	ADDQ    R10, R9
	MOVB    R9, (DI)

	// G: dst[1] + sca_g - mul(dst[1], sa)
	MOVBLZX 1(DI), R9
	MOVQ    R9, R10
	IMULQ   BX, R10
	ADDQ    $128, R10
	MOVQ    R10, R11
	SHRQ    $8, R11
	ADDQ    R11, R10
	SHRQ    $8, R10
	SUBQ    R10, R9
	MOVQ    AX, R10
	SHRQ    $8, R10
	ANDQ    $0xFF, R10         // sca_g (byte 1)
	ADDQ    R10, R9
	MOVB    R9, 1(DI)

	// B: dst[2] + sca_b - mul(dst[2], sa)
	MOVBLZX 2(DI), R9
	MOVQ    R9, R10
	IMULQ   BX, R10
	ADDQ    $128, R10
	MOVQ    R10, R11
	SHRQ    $8, R11
	ADDQ    R11, R10
	SHRQ    $8, R10
	SUBQ    R10, R9
	MOVQ    AX, R10
	SHRQ    $16, R10
	ANDQ    $0xFF, R10         // sca_b (byte 2)
	ADDQ    R10, R9
	MOVB    R9, 2(DI)

	// A: dst[3] + sa - mul(dst[3], sa)
	MOVBLZX 3(DI), R9
	MOVQ    R9, R10
	IMULQ   BX, R10
	ADDQ    $128, R10
	MOVQ    R10, R11
	SHRQ    $8, R11
	ADDQ    R11, R10
	SHRQ    $8, R10
	SUBQ    R10, R9
	ADDQ    BX, R9             // + sa
	MOVB    R9, 3(DI)

	ADDQ $4, DI
	DECQ CX
	JNZ  tail_loop

done:
	RET
