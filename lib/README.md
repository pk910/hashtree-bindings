# Static Library Bundle

## Build Information
- **Commit:** 6d7afea5dbda0dcfc1f4e7b1a67bf4fc6dc6b161
- **Build Date:** 2025-11-08 03:02:58 UTC
- **Build Trigger:** schedule
- **Repository:** pk910/hashtree-bindings
- **Workflow:** Build Static Libraries
- **Run ID:** 19186992224

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
1. Check out commit: `git checkout 6d7afea5dbda0dcfc1f4e7b1a67bf4fc6dc6b161`
2. Use the same GitHub Actions environment (see workflow file)
3. The build should produce identical checksums
