//go:build cgo && linux && riscv64
// +build cgo,linux,riscv64

package hashtree

import (
	// link to the static library via cgo
	_ "github.com/pk910/hashtree-bindings/cgo"
)

//go:noescape
func HashtreeHash(output *byte, input *byte, count uint64)

func init() {
	supportedCPU = true
	hashtreeHash = HashtreeHash
}
