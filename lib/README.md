# Static Library Bundle

## Hashtree Version
- **Commit:** 8e32ae9bda35ee388d74546b7d4f62e8b50c55f9
- **Tag:** untagged

## Build Information
- **Build Date:** 2026-01-20 03:45:06 UTC
- **Build Trigger:** schedule
- **Repository:** pk910/hashtree-bindings
- **Workflow:** Build Static Libraries
- **Run ID:** 21158687823

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
sha256sum -c libhashtree.a.sha256
```
