#!/usr/bin/env bash
#
# Cross-architecture / cross-CPU-feature test matrix for the hashtree bindings.
#
# For each Linux target it cross-compiles a statically linked cgo test binary
# (linking the per-platform lib/<platform>/libhashtree.a) and runs it under
# qemu-user with several -cpu models, so every runtime-detected implementation
# path (and the pure-Go fallback) is exercised on one host. HASHTREE_EXPECT_PATH
# makes the binary assert which path it selected, so a -cpu that silently fails
# to apply is caught instead of giving false coverage.
#
# Requirements (provided by the CI container, see build-static-libs.yml):
#   go, gcc, gcc-aarch64-linux-gnu, gcc-riscv64-linux-gnu, qemu-user-static,
#   and the static libs present in lib/{linux_amd64,linux_arm64,linux_riscv64}/.
#
# NOTE: qemu-user (TCG) cannot emulate AVX-512, so that amd64 path is covered
# natively elsewhere, not here.

set -u
cd "$(cd "$(dirname "$0")/../.." && pwd)"

LDFLAGS='-linkmode external -extldflags "-static"'
BIN_DIR="$(mktemp -d)"
fail=0

build() { # arch cc
	local arch="$1" cc="$2"
	echo ">> building $arch test binary"
	if ! CGO_ENABLED=1 GOOS=linux GOARCH="$arch" CC="$cc" \
		go test -c -ldflags "$LDFLAGS" -o "$BIN_DIR/$arch.test" .; then
		echo "BUILD FAIL $arch"
		fail=1
	fi
}

run() { # arch qemu cpu expect [xfail]
	local arch="$1" qemu="$2" cpu="$3" expect="$4" xfail="${5:-0}"
	local log="$BIN_DIR/$arch.log"
	HASHTREE_EXPECT_PATH="$expect" "$qemu" -cpu "$cpu" "$BIN_DIR/$arch.test" \
		-test.run 'TestHash|TestZeroHashChain|TestErrors|TestForcedPaths|TestActivePathMatchesEnv' \
		>"$log" 2>&1
	local rc=$?
	if [ "$xfail" = "1" ]; then
		if [ $rc -ne 0 ]; then
			printf '  XFAIL %-8s %-42s -> %s (known upstream Zbb-detect bug)\n' "$arch" "$cpu" "$expect"
		else
			printf '  XPASS %-8s %-42s -> %s  (upstream fixed? remove the xfail)\n' "$arch" "$cpu" "$expect"
			fail=1
		fi
		return
	fi
	if [ $rc -eq 0 ]; then
		printf '  PASS  %-8s %-42s -> %s\n' "$arch" "$cpu" "$expect"
	else
		printf '  FAIL  %-8s %-42s -> %s\n' "$arch" "$cpu" "$expect"
		grep -E '^(---|    |FAIL|panic|SIGILL|SIGSEGV)' "$log" | head -10 | sed 's/^/        /'
		fail=1
	fi
}

go version
build amd64   gcc
build arm64   aarch64-linux-gnu-gcc
build riscv64 riscv64-linux-gnu-gcc

echo "=== matrix ==="
# amd64: qemu64 has no AVX state -> pure-Go fallback; named models carry the
# OSXSAVE/XCR0 state that github.com/klauspost/cpuid requires.
run amd64   qemu-x86_64  'qemu64'                                  generic
run amd64   qemu-x86_64  'Haswell'                                 avx2
run amd64   qemu-x86_64  'EPYC'                                    shani
# arm64: only the SHA-2 path is reachable via a -cpu (every qemu model reports
# SHA2); neon_x4 and the fallback are exercised by the inline forced tests.
run arm64   qemu-aarch64 'max'                                     sha_x1
# riscv64: scalar is xfail until upstream fixes its riscv_hwprobe bit constants
# (it mis-selects the Zbb impl on a non-Zbb CPU -> SIGILL). See upstream report.
run riscv64 qemu-riscv64 'rv64,zbb=false'                          scalar  1
run riscv64 qemu-riscv64 'rv64,zbb=true'                           zbb
run riscv64 qemu-riscv64 'rv64,zbb=true,zknh=true,zbkb=true'       crypto

echo "=== result: $([ $fail -eq 0 ] && echo ALL GREEN || echo FAILURES) ==="
exit $fail
