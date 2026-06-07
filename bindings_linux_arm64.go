//go:build cgo && linux && arm64
// +build cgo,linux,arm64

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
	// NEON / Advanced SIMD is mandatory in the ARMv8-A (AArch64) base
	// architecture, so it is present on every GOARCH=arm64 CPU. The wrapper's
	// non-SHA2 path (hashtree_sha256_neon_x4) is therefore always safe to use,
	// and we can mark the CPU as supported unconditionally. hasShani only
	// selects the faster SHA-2 crypto-extension path (sha_x1) over NEON.
	// (Without this, SHA2-less ARM would fall back to the slow pure-Go
	// implementation instead of NEON. Matches bindings_darwin_arm64.go.)
	supportedCPU = true
	hashtreeHash = HashtreeHash
}
