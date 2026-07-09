# Static Library Bundle

## Hashtree Version
- **Commit:** 30497cff98a06362eadde897202634f91d504fd8
- **Tag:** v0.2.5

## Build Information
- **Build Date:** 2026-07-09 06:12:17 UTC
- **Build Trigger:** schedule
- **Repository:** pk910/hashtree-bindings
- **Workflow:** Update Static Libraries
- **Run ID:** 28997726101

## Contents
Each platform directory contains:
- `libhashtree.a` - Static library with platform-specific optimizations
- `libhashtree.a.sha256` - SHA256 checksum for verification

### Platform Details:
- **linux_amd64/** - Linux x86_64 with AVX2/AVX512/SHANI optimizations
- **linux_arm64/** - Linux ARM64 with NEON/SHA2 optimizations
- **linux_riscv64/** - Linux RISC-V64 with Zbb / Zknh scalar-crypto (runtime-detected)
- **darwin_arm64/** - macOS Apple Silicon with SHA2 optimizations
- **windows_amd64/** - Windows x86_64 with AVX2/AVX512/SHANI optimizations

**Note:** macOS Intel (darwin_amd64) is not included - uses pure Go fallback implementation

## Build Environment

Every library is built from the hashtree commit above. `SOURCE_DATE_EPOCH`
is exported from that commit's timestamp before each build:

```bash
export SOURCE_DATE_EPOCH=$(git -C hashtree log -1 --format=%ct)
```

| Platform | Host / toolchain | Compiler | `make` invocation |
|----------|------------------|----------|-------------------|
| linux_amd64 | ubuntu-22.04, native | `gcc (Ubuntu 11.4.0-1ubuntu1~22.04.3) 11.4.0` | `make CFLAGS="-O3 -Wall -Werror -static -fno-plt -ffile-prefix-map=$(pwd)=." ASFLAGS="-g -fpic"` |
| linux_arm64 | ubuntu-22.04, gcc-aarch64-linux-gnu | `aarch64-linux-gnu-gcc (Ubuntu 11.4.0-1ubuntu1~22.04.3) 11.4.0` | `make CC=aarch64-linux-gnu-gcc ARM=1 CFLAGS="-O3 -Wall -Werror -static -fno-plt -ffile-prefix-map=$(pwd)=." ASFLAGS="-g -fpic"` |
| linux_riscv64 | ubuntu-22.04, gcc-riscv64-linux-gnu | `riscv64-linux-gnu-gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0` | `make CC=riscv64-linux-gnu-gcc RISCV=1 CFLAGS="-O3 -Wall -Werror -static -fno-plt -ffile-prefix-map=$(pwd)=. -march=rv64gc_zbb_zk" ASFLAGS="-g -fpic -march=rv64gc_zbb_zk"` |
| darwin_arm64 | macos-14, Apple clang | `Apple clang version 15.0.0 (clang-1500.3.9.4)` | `make CFLAGS="-O3 -Wall -Werror -mmacosx-version-min=11.0 -ffile-prefix-map=$(pwd)=." ASFLAGS="-g -fpic"` |
| windows_amd64 | windows-latest, MSYS2 MINGW64 | `gcc.exe (Rev5, Built by MSYS2 project) 16.1.0` | `make CFLAGS="-O3 -Wall -Werror -static -ffile-prefix-map=$(pwd)=." ASFLAGS="-g -fpic"` |

After `make`, the archive is copied out of `build/lib/`, stripped, and re-indexed:

| Platform | Archive normalization |
|----------|-----------------------|
| linux_amd64 | `strip --strip-debug` then `ranlib` |
| linux_arm64 | `aarch64-linux-gnu-strip --strip-debug` then `aarch64-linux-gnu-ranlib` |
| linux_riscv64 | `riscv64-linux-gnu-strip --strip-debug` then `riscv64-linux-gnu-ranlib` |
| darwin_arm64 | `strip -S` then `ranlib`, then zero `ar` member mtime/uid/gid |
| windows_amd64 | `strip --strip-debug` then `ranlib` |

## Reproducible Builds

Given the same compiler version, the build is deterministic:
- `SOURCE_DATE_EPOCH` pins embedded timestamps to the commit time.
- `-ffile-prefix-map=$(pwd)=.` strips absolute build paths from the objects.
- Debug symbols are stripped and the archive index is rebuilt with `ranlib`.
- GNU `ar`/`ranlib` (Linux and Windows/MinGW) write deterministic archives by
  default. macOS BSD tools do not, so member `mtime`/`uid`/`gid` are zeroed
  afterwards (see `.github/scripts/zero_ar_metadata.py`).

Because the object code depends on the exact compiler, a byte-for-byte match
requires the **same compiler version** shown in the table above.

## Verification

### 1. Checksum
```bash
cd lib/<platform>
sha256sum -c libhashtree.a.sha256   # macOS: shasum -a 256 -c libhashtree.a.sha256
```

### 2. Rebuild & compare (Docker)
Rebuild from upstream in a clean container and compare the digest with the
published one. Each command is self-contained: copy it, run it, and check
the printed hash against the matching `libhashtree.a.sha256`. These cover
the Linux targets; macOS and Windows use native toolchains that cannot run
in a Linux container.

**linux_amd64** - expect the hash in `lib/linux_amd64/libhashtree.a.sha256`:
```bash
docker run --rm -e COMMIT=30497cff98a06362eadde897202634f91d504fd8 ubuntu:22.04 bash -c '
  set -eu
  apt-get update -qq
  apt-get install -y -qq git ca-certificates make build-essential >/dev/null
  git clone -q https://github.com/OffchainLabs/hashtree /h && git -C /h checkout -q "$COMMIT"
  cd /h/src
  export SOURCE_DATE_EPOCH=$(git -C /h log -1 --format=%ct)
  make CFLAGS="-O3 -Wall -Werror -static -fno-plt -ffile-prefix-map=$(pwd)=." ASFLAGS="-g -fpic" >/dev/null
  cp ../build/lib/libhashtree.a . && strip --strip-debug libhashtree.a && ranlib libhashtree.a
  sha256sum libhashtree.a
'
```

**linux_arm64** - expect the hash in `lib/linux_arm64/libhashtree.a.sha256`:
```bash
docker run --rm -e COMMIT=30497cff98a06362eadde897202634f91d504fd8 ubuntu:22.04 bash -c '
  set -eu
  apt-get update -qq
  apt-get install -y -qq git ca-certificates make binutils gcc-aarch64-linux-gnu binutils-aarch64-linux-gnu >/dev/null
  git clone -q https://github.com/OffchainLabs/hashtree /h && git -C /h checkout -q "$COMMIT"
  cd /h/src
  export SOURCE_DATE_EPOCH=$(git -C /h log -1 --format=%ct)
  make CC=aarch64-linux-gnu-gcc ARM=1 CFLAGS="-O3 -Wall -Werror -static -fno-plt -ffile-prefix-map=$(pwd)=." ASFLAGS="-g -fpic" >/dev/null
  cp ../build/lib/libhashtree.a . && aarch64-linux-gnu-strip --strip-debug libhashtree.a && aarch64-linux-gnu-ranlib libhashtree.a
  sha256sum libhashtree.a
'
```

**linux_riscv64** - expect the hash in `lib/linux_riscv64/libhashtree.a.sha256`:
```bash
docker run --rm -e COMMIT=30497cff98a06362eadde897202634f91d504fd8 ubuntu:22.04 bash -c '
  set -eu
  apt-get update -qq
  apt-get install -y -qq git ca-certificates make binutils gcc-riscv64-linux-gnu binutils-riscv64-linux-gnu >/dev/null
  git clone -q https://github.com/OffchainLabs/hashtree /h && git -C /h checkout -q "$COMMIT"
  cd /h/src
  export SOURCE_DATE_EPOCH=$(git -C /h log -1 --format=%ct)
  make CC=riscv64-linux-gnu-gcc RISCV=1 CFLAGS="-O3 -Wall -Werror -static -fno-plt -ffile-prefix-map=$(pwd)=. -march=rv64gc_zbb_zk" ASFLAGS="-g -fpic -march=rv64gc_zbb_zk" >/dev/null
  cp ../build/lib/libhashtree.a . && riscv64-linux-gnu-strip --strip-debug libhashtree.a && riscv64-linux-gnu-ranlib libhashtree.a
  sha256sum libhashtree.a
'
```

> Verified reproducible with gcc 11.4.0 (Ubuntu 22.04). If a digest differs,
> confirm your `gcc --version` matches the **Compiler** column above; a
> different compiler patch release can change the output.
