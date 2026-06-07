//go:build linux && riscv64
// +build linux,riscv64

package cgo

/*
#cgo LDFLAGS: -L../lib/linux_riscv64 -lhashtree
*/
import "C"
