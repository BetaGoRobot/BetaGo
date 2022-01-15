#include "textflag.h"

#ifdef GOARCH_amd64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ 152(R14), R13
	MOVQ R13, ret+0(FP)
	RET
#endif

#ifdef GOARCH_arm64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVD g, R14
	MOVD 152(R14), R13
	MOVD R13, ret+0(FP)
	RET
#endif

#ifdef GOARCH_arm
TEXT ·goid(SB),NOSPLIT,$0-4
	MOVW g, R8
	MOVW 80(R8), R7
	MOVW R7, ret+0(FP)
	RET
#endif

#ifdef GOARCH_386
TEXT ·goid(SB),NOSPLIT,$0-4
	MOVL (TLS), AX
	MOVL 80(AX), BX
	MOVL BX, ret+0(FP)
	RET
#endif
