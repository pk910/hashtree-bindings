//go:build cgo && amd64

package hashtree

import "github.com/klauspost/cpuid/v2"

// activePath reports which implementation the wrapper will dispatch to, mirroring
// the precedence in wrapper_*_amd64.s (shani > avx512 > avx2) and the
// supportedCPU gate. avx_x1 is unreachable in practice (supportedCPU requires one
// of the three) and only listed for completeness.
func activePath() string {
	if !supportedCPU {
		return "generic"
	}
	switch {
	case hasShani:
		return "shani"
	case hasAVX512:
		return "avx512"
	case hasAVX2:
		return "avx2"
	default:
		return "avx_x1"
	}
}

// forceAMD64 returns an apply() that overrides the detection vars to
// (supported, shani, avx512, avx2) and restores them afterwards. The wrapper
// reads these at call time, so this redirects dispatch to a chosen branch.
func forceAMD64(supported, shani, avx512, avx2 bool) func() func() {
	return func() func() {
		s, sh, a5, a2 := supportedCPU, hasShani, hasAVX512, hasAVX2
		supportedCPU, hasShani, hasAVX512, hasAVX2 = supported, shani, avx512, avx2
		return func() { supportedCPU, hasShani, hasAVX512, hasAVX2 = s, sh, a5, a2 }
	}
}

// Every dispatch branch, each gated on whether this CPU can actually run that
// path's instructions (captured from real detection before any forcing). On a
// runner this exercises all locally-supported optimized paths, not just the one
// natural detection would pick.
var forcedCases = []forcedCase{
	{"generic", true, forceAMD64(false, false, false, false)},
	{"avx2", hasAVX2, forceAMD64(true, false, false, true)},
	{"shani", hasShani, forceAMD64(true, true, false, false)},
	{"avx512", hasAVX512, forceAMD64(true, false, true, false)},
	{"avx_x1", cpuid.CPU.Supports(cpuid.AVX), forceAMD64(true, false, false, false)},
}
