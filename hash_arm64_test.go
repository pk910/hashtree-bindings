//go:build cgo && arm64

package hashtree

// activePath reports which implementation the wrapper will dispatch to, mirroring
// wrapper_arm64.s: the SHA-2 crypto path when available, else NEON. NEON is the
// guaranteed baseline on ARMv8 (see bindings_linux_arm64.go), so supportedCPU is
// always true; the generic branch only matters under the forced-fallback test.
func activePath() string {
	if !supportedCPU {
		return "generic"
	}
	if hasShani {
		return "sha_x1"
	}
	return "neon_x4"
}

var forcedCases = []forcedCase{
	{"generic", true, func() func() {
		saved := supportedCPU
		supportedCPU = false
		return func() { supportedCPU = saved }
	}},
	// NEON is mandatory on every arm64 CPU, so this path is always runnable; it
	// covers neon_x4, which natural detection never selects on CPUs with crypto.
	{"neon_x4", true, func() func() {
		savedSupported, savedShani := supportedCPU, hasShani
		supportedCPU, hasShani = true, false
		return func() { supportedCPU, hasShani = savedSupported, savedShani }
	}},
	// sha_x1 requires the SHA-2 crypto extension; gated accordingly.
	{"sha_x1", hasShani, func() func() {
		savedSupported, savedShani := supportedCPU, hasShani
		supportedCPU, hasShani = true, true
		return func() { supportedCPU, hasShani = savedSupported, savedShani }
	}},
}
