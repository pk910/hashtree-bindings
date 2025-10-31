# Static Library Bundle

## Build Information
- **Commit:** 5090ad29e5f2122f100e990db7369d0b4e7094ec
- **Build Date:** 2025-10-31 03:11:26 UTC
- **Build Trigger:** schedule
- **Repository:** pk910/hashtree-bindings
- **Workflow:** Build Static Libraries
- **Run ID:** 18961487052

## Contents
Each platform directory contains:
- `libhashtree.a` - Static library with platform-specific optimizations
- `libhashtree.a.sha256` - SHA256 checksum for verification

### Platform Details:
- **linux_amd64/** - Linux x86_64 with AVX2/AVX512/SHANI optimizations
- **linux_arm64/** - Linux ARM64 with NEON/SHA2 optimizations
- **darwin_arm64/** - macOS Apple Silicon with SHA2 optimizations
- **windows_amd64/** - Windows x86_64 with AVX2/AVX512/SHANI optimizations

**Note:** macOS Intel (darwin_amd64) is not included - uses pure Go fallback implementation

## Reproducible Builds
These libraries are built with reproducible build flags:
- `SOURCE_DATE_EPOCH` set to commit timestamp
- `-ffile-prefix-map` to normalize file paths
- Debug symbols stripped for determinism
- Archive metadata normalized

## Verification
To verify library integrity:
```bash
# Verify SHA256 checksum
sha256sum -c libhashtree.a.sha256

```

## Reproduction
To reproduce these builds:
1. Check out commit: `git checkout 5090ad29e5f2122f100e990db7369d0b4e7094ec`
2. Use the same GitHub Actions environment (see workflow file)
3. The build should produce identical checksums
