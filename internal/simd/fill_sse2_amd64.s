//go:build amd64 && !purego

#include "textflag.h"

// func fillRGBASSE2Asm(dst []byte, pixel uint32, count int)
TEXT ·fillRGBASSE2Asm(SB), NOSPLIT, $0-40
	MOVQ dst_base+0(FP), DI
	MOVL pixel+24(FP), AX
	MOVQ count+32(FP), CX

	TESTQ CX, CX
	JLE done

	MOVD AX, X0
	PSHUFD $0, X0, X0

	CMPQ CX, $4
	JB tail

loop:
	MOVOU X0, (DI)
	ADDQ $16, DI
	SUBQ $4, CX
	CMPQ CX, $4
	JGE loop

tail:
	TESTQ CX, CX
	JLE done

tail_loop:
	MOVL AX, (DI)
	ADDQ $4, DI
	DECQ CX
	JNZ tail_loop

done:
	RET
