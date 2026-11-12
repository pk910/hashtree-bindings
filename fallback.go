//go:build !(cgo && ((linux && (amd64 || arm64 || riscv64)) || (darwin && arm64) || (windows && amd64)))

package hashtree

import "unsafe"

// Builds without a static library (cgo disabled, or a platform we ship no
// library for) always hash with the pure-Go implementation.
var supportedCPU = false

// hashtreeHashFallback is the body of HashtreeHash in these builds. bindings.go
// comes from upstream unmodified and declares HashtreeHash without a body,
// expecting the assembly wrapper; here it gets a pure-Go one instead. Hash and
// HashByteSlice never reach it (supportedCPU is false), only direct callers do.
//
//go:linkname hashtreeHashFallback github.com/pk910/hashtree-bindings.HashtreeHash
func hashtreeHashFallback(output *byte, input *byte, count uint64) {
	if count == 0 {
		return
	}

	digests := unsafe.Slice((*[32]byte)(unsafe.Pointer(output)), count)
	chunks := unsafe.Slice((*[32]byte)(unsafe.Pointer(input)), 2*count)
	sha256_1_generic(digests, chunks)
}
