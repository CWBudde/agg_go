//go:build amd64 && !purego

#include "textflag.h"

DATA ·alphaDupMask2+0(SB)/1, $0
DATA ·alphaDupMask2+1(SB)/1, $0
DATA ·alphaDupMask2+2(SB)/1, $0
DATA ·alphaDupMask2+3(SB)/1, $0
DATA ·alphaDupMask2+4(SB)/1, $1
DATA ·alphaDupMask2+5(SB)/1, $1
DATA ·alphaDupMask2+6(SB)/1, $1
DATA ·alphaDupMask2+7(SB)/1, $1
DATA ·alphaDupMask2+8(SB)/1, $0x80
DATA ·alphaDupMask2+9(SB)/1, $0x80
DATA ·alphaDupMask2+10(SB)/1, $0x80
DATA ·alphaDupMask2+11(SB)/1, $0x80
DATA ·alphaDupMask2+12(SB)/1, $0x80
DATA ·alphaDupMask2+13(SB)/1, $0x80
DATA ·alphaDupMask2+14(SB)/1, $0x80
DATA ·alphaDupMask2+15(SB)/1, $0x80
GLOBL ·alphaDupMask2(SB), RODATA|NOPTR, $16

// func blendSolidHspanRGBASSE41Asm(dst []byte, covers []byte, pixelOpaque uint32, srcA uint8, count int)
TEXT ·blendSolidHspanRGBASSE41Asm(SB), NOSPLIT, $0-64
	MOVQ dst_base+0(FP), DI
	MOVQ covers_base+24(FP), SI
	MOVL pixelOpaque+48(FP), AX
	MOVBLZX srcA+52(FP), BX
	MOVQ count+56(FP), CX

	TESTQ CX, CX
	JLE done

	PXOR X15, X15
	MOVD AX, X10
	PSHUFD $0, X10, X10
	PMOVZXBW X10, X10
	MOVOU ·bias128W(SB), X14

	MOVL BX, DX
	CMPQ CX, $2
	JB tail

loop_dispatch:
	CMPQ DX, $255
	JE loop_opaque

	MOVD DX, X11
	PSHUFLW $0, X11, X11
	PUNPCKLQDQ X11, X11

loop_alpha:
	MOVQ (DI), X0
	MOVW (SI), R13
	MOVD R13, X1
	PSHUFB ·alphaDupMask2(SB), X1
	PMOVZXBW X0, X0
	PMOVZXBW X1, X1
	PMULLW X11, X1
	PADDW X14, X1
	MOVOU X1, X2
	PSRLW $8, X2
	PADDW X2, X1
	PSRLW $8, X1
	JMP blend_words

loop_opaque:
	MOVQ (DI), X0
	MOVW (SI), R13
	MOVD R13, X1
	PSHUFB ·alphaDupMask2(SB), X1
	PMOVZXBW X0, X0
	PMOVZXBW X1, X1

blend_words:
	MOVOU X10, X3
	PMAXUW X0, X3
	MOVOU X10, X4
	PMINUW X0, X4
	MOVOU X3, X9
	PSUBW X4, X3
	PMULLW X1, X3
	PADDW X14, X3
	MOVOU X3, X6
	PSRLW $8, X6
	PADDW X6, X3
	PSRLW $8, X3
	MOVOU X0, X7
	PADDW X3, X7
	MOVOU X0, X8
	PSUBW X3, X8
	PCMPEQW X10, X9
	MOVOU X9, X12
	PAND X7, X12
	PANDN X8, X9
	POR X12, X9
	PACKUSWB X9, X9
	MOVQ X9, (DI)
	ADDQ $8, DI
	ADDQ $2, SI
	SUBQ $2, CX
	CMPQ CX, $2
	JGE loop_dispatch

tail:
	TESTQ CX, CX
	JLE done

tail_loop:
	MOVBLZX (SI), R8
	TESTQ R8, R8
	JZ tail_next

	CMPQ BX, $255
	JNE tail_alpha
	CMPQ R8, $255
	JNE tail_use_cover
	MOVL AX, (DI)
	JMP tail_next

tail_alpha:
	IMULQ BX, R8
	ADDQ $128, R8
	MOVQ R8, R9
	SHRQ $8, R9
	ADDQ R9, R8
	SHRQ $8, R8
	JMP tail_have_alpha

tail_use_cover:
tail_have_alpha:
	MOVBLZX AX, R9
	MOVBLZX (DI), R10
	CMPQ R9, R10
	JAE tail_r_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX (DI), R10
	SUBQ R11, R10
	MOVB R10, (DI)
	JMP tail_g

tail_r_add:
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX (DI), R10
	ADDQ R11, R10
	MOVB R10, (DI)

tail_g:
	MOVQ AX, R9
	SHRQ $8, R9
	ANDQ $0xFF, R9
	MOVBLZX 1(DI), R10
	CMPQ R9, R10
	JAE tail_g_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX 1(DI), R10
	SUBQ R11, R10
	MOVB R10, 1(DI)
	JMP tail_b

tail_g_add:
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX 1(DI), R10
	ADDQ R11, R10
	MOVB R10, 1(DI)

tail_b:
	MOVQ AX, R9
	SHRQ $16, R9
	ANDQ $0xFF, R9
	MOVBLZX 2(DI), R10
	CMPQ R9, R10
	JAE tail_b_add
	SUBQ R9, R10
	MOVQ R10, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX 2(DI), R10
	SUBQ R11, R10
	MOVB R10, 2(DI)
	JMP tail_a

tail_b_add:
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX 2(DI), R10
	ADDQ R11, R10
	MOVB R10, 2(DI)

tail_a:
	MOVQ $255, R9
	MOVBLZX 3(DI), R10
	SUBQ R10, R9
	MOVQ R9, R11
	IMULQ R8, R11
	ADDQ $128, R11
	MOVQ R11, R12
	SHRQ $8, R12
	ADDQ R12, R11
	SHRQ $8, R11
	MOVBLZX 3(DI), R10
	ADDQ R11, R10
	MOVB R10, 3(DI)

tail_next:
	ADDQ $4, DI
	INCQ SI
	DECQ CX
	JNZ tail_loop

done:
	RET
