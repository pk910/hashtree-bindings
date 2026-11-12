//go:build !cgo || (!amd64 && !arm64 && !riscv64)

package hashtree

import "testing"

// Fallback for builds without cgo (pure-Go fallback only) or architectures
// without a dedicated optimized path. supportedCPU is false in these builds, so
// the library always uses the pure-Go implementation.
func activePath() string {
	if supportedCPU {
		return "unknown"
	}
	return "generic"
}

var forcedCases = []forcedCase{
	{"generic", true, func() func() {
		saved := supportedCPU
		supportedCPU = false
		return func() { supportedCPU = saved }
	}},
}

// TestHashtreeHashFallback calls HashtreeHash directly. In these builds its body
// is the pure-Go one that fallback.go attaches to the upstream declaration.
func TestHashtreeHashFallback(t *testing.T) {
	for _, pairs := range []int{1, 2, 17} {
		chunks := genChunks(pairs, int64(pairs))
		digests := make([][32]byte, pairs)

		HashtreeHash(&digests[0][0], &chunks[0][0], uint64(pairs))

		for i := range pairs {
			if want := referenceParent(chunks[2*i], chunks[2*i+1]); digests[i] != want {
				t.Fatalf("pairs=%d: digest %d = %x, want %x", pairs, i, digests[i], want)
			}
		}
	}
}
