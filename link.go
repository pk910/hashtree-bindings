//go:build cgo && ((linux && (amd64 || arm64 || riscv64)) || (darwin && arm64) || (windows && amd64))

package hashtree

// Link the prebuilt static library for this platform. The cgo directives live
// in their own package because a package cannot mix cgo with Go assembly.
import _ "github.com/pk910/hashtree-bindings/cgo"
