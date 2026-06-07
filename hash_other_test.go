//go:build !cgo || (!amd64 && !arm64 && !riscv64)

package hashtree

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
