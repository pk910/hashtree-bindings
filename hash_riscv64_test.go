//go:build cgo && riscv64

package hashtree

import (
	"syscall"
	"unsafe"
)

// The riscv64 implementation is chosen inside the C library by hashtree_detect,
// which reads the riscv_hwprobe syscall. That function is static (not linkable),
// so activePath replays the exact same hwprobe read and selection logic here.
// This confirms the emulated/native CPU exposes the extensions we expect (i.e.
// that a QEMU -cpu actually took effect) without depending on C internals.
const (
	nrRiscvHwprobe    = 258 // __NR_riscv_hwprobe
	hwprobeKeyImaExt0 = 4   // RISCV_HWPROBE_KEY_IMA_EXT_0
	hwprobeExtZbb     = 1 << 4
	hwprobeExtZbkb    = 1 << 8
	hwprobeExtZknh    = 1 << 13
)

type riscvHwprobePair struct {
	key   int64
	value uint64
}

func activePath() string {
	if !supportedCPU {
		return "generic"
	}
	p := riscvHwprobePair{key: hwprobeKeyImaExt0}
	_, _, errno := syscall.Syscall6(nrRiscvHwprobe,
		uintptr(unsafe.Pointer(&p)), 1, 0, 0, 0, 0)
	if errno != 0 {
		// No hwprobe (kernel < 6.4): the C side falls back to base scalar.
		return "scalar"
	}
	v := p.value
	switch {
	case v&hwprobeExtZknh != 0 && v&hwprobeExtZbkb != 0:
		return "crypto"
	case v&hwprobeExtZbb != 0:
		return "zbb"
	default:
		return "scalar"
	}
}

// The riscv asm paths are selected inside the C library (not via Go vars), so
// they can't be force-overridden here; they are covered by the qemu -cpu matrix.
// Only the pure-Go fallback is forceable from Go.
var forcedCases = []forcedCase{
	{"generic", true, func() func() {
		saved := supportedCPU
		supportedCPU = false
		return func() { supportedCPU = saved }
	}},
}
