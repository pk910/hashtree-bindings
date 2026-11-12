//go:build cgo && (linux || darwin) && arm64
// +build cgo
// +build linux darwin
// +build arm64

package hashtree

import (
	"github.com/klauspost/cpuid/v2"
)

var hasShani = cpuid.CPU.Supports(cpuid.SHA2)
var supportedCPU = true
