# hashtree-bindings

Go bindings for the high-performance [hashtree](https://github.com/OffchainLabs/hashtree) SHA-256 library with CPU-optimized implementations.

## Overview

This repository provides Go bindings for the hashtree library, which implements highly optimized SHA-256 hashing with support for:
- **x86-64**: AVX2, AVX512, SHA-NI (Intel SHA Extensions)
- **ARM64**: NEON, ARM SHA2 crypto extensions
- **RISC-V 64**: Zbb (bit-manipulation), Zknh (scalar SHA-2 crypto) extensions
- **Fallback**: Pure Go implementation for unsupported architectures

The library automatically selects the best available implementation at runtime based on CPU capabilities.
















## Hashtree Version

Built from [hashtree](https://github.com/OffchainLabs/hashtree) library:
- **Commit:** [`76c54d6`](https://github.com/OffchainLabs/hashtree/commit/76c54d6dcf3596c03788ea5bc0ef030e9f1b43f4)
- **Version:** `v0.2.6`


## Installation

```bash
go get github.com/pk910/hashtree-bindings
```

## Usage

```go
package main

import (
    "fmt"
    hashtree "github.com/pk910/hashtree-bindings"
)

func main() {
    // Single hash
    data := []byte("hello world")
    hash := hashtree.Hash(data)
    fmt.Printf("SHA256: %x\n", hash)
    
    // Parallel hashing (optimized for multiple inputs)
    inputs := [][]byte{
        []byte("input1"),
        []byte("input2"),
        []byte("input3"),
        []byte("input4"),
    }
    hashes := hashtree.HashParallel(inputs)
    for i, hash := range hashes {
        fmt.Printf("Hash[%d]: %x\n", i, hash)
    }
}
```

## Performance

The hashtree library provides significant performance improvements over standard implementations:
- Up to 8x faster than pure Go crypto/sha256 on AVX512 CPUs
- Up to 4x faster on CPUs with SHA-NI support
- Optimized parallel hashing for batch operations

## Supported Platforms

| Platform | Architecture | Optimizations |
|----------|-------------|---------------|
| Linux | amd64 | AVX2, AVX512, SHA-NI |
| Linux | arm64 | NEON, SHA2 |
| Linux | riscv64 | Zbb, Zknh (scalar SHA-2 crypto) |
| macOS | arm64 (Apple Silicon) | SHA2 |
| Windows | amd64 | AVX2, AVX512, SHA-NI |
| Others | any | Pure Go fallback |

## Build Process

This repository uses GitHub Actions to automatically build and update static libraries from the hashtree submodule. The workflow:
1. Runs daily to check for upstream changes
2. Builds platform-specific static libraries
3. Regenerates the Go and assembly sources from the same upstream commit
4. Automatically commits updates when libraries or sources change

### Relationship to Upstream

The only intended difference from upstream is the packaging: upstream links `.syso` objects, this module links the static libraries in `lib/` through cgo. Everything else is taken from the `hashtree` submodule by `internal/upstreamsync` and must not be edited here:

| Files | Origin |
|-------|--------|
| `bindings.go`, `sha256_1_generic.go`, `bindings_test.go` | Copied from upstream unchanged |
| `bindings_<arch>.go`, `wrapper_*.s` | Copied from upstream; only the build constraint is narrowed to cgo builds on platforms with a static library. |
| `link.go`, `cgo/`, `lib/` | Ours: link the static library |
| `fallback.go`, `fallback.s` | Ours: pure-Go fallback for builds without cgo or without a library for the platform |
| `hash_*_test.go` | Ours: additional tests for every dispatch path |

```bash
go generate ./...                        # regenerate from the submodule
go run ./internal/upstreamsync -check    # verify, as CI does
```

The tool fails if upstream adds or removes a source file, so a change upstream cannot be missed silently.

### Manual Build

To build the libraries manually:

```bash
# Clone with submodules
git clone --recursive https://github.com/pk910/hashtree-bindings
cd hashtree-bindings

# Build hashtree C library
cd hashtree
make libhashtree

# Build Go bindings
cd ..
go build
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

The hashtree library implementation is based on work by:
- Original implementation: [Offchain Labs](https://github.com/OffchainLabs/hashtree)
- Go bindings and modifications: [pk910](https://github.com/pk910/hashtree-bindings)

## Acknowledgments

- Original hashtree implementation by [Offchain Labs](https://github.com/OffchainLabs/hashtree)
