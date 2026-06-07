//go:build cgo && linux && riscv64
// +build cgo,linux,riscv64

TEXT ·HashtreeHash(SB), 0, $0-24
	MOV	output+0(FP), A0
	MOV	input+8(FP), A1
	MOV	count+16(FP), A2

	CALL	hashtree_hash(SB)
	RET
