//go:build cgo && amd64

package hashtree

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

var forcedCases = []forcedCase{
	{"generic", func() func() {
		saved := supportedCPU
		supportedCPU = false
		return func() { supportedCPU = saved }
	}},
}
