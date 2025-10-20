//go:build cgo && darwin && arm64
// +build cgo,darwin,arm64

package hashtree

import (
	"github.com/klauspost/cpuid/v2"

	// link to the static library via cgo
	_ "github.com/pk910/hashtree-bindings/cgo"
)

//go:noescape
func HashtreeHash(output *byte, input *byte, count uint64)

var hasShani = cpuid.CPU.Supports(cpuid.SHA2)

func init() {
	supportedCPU = true
	hashtreeHash = HashtreeHash
}
