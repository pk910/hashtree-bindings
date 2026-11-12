// Package hashtree provides Go bindings for the hashtree SHA-256 library,
// linked from prebuilt static libraries.
//
// The Go and assembly sources of the bindings are taken from the upstream
// repository (the hashtree submodule) by internal/upstreamsync. Do not edit
// them here; change them upstream and regenerate.
package hashtree

//go:generate go run ./internal/upstreamsync
