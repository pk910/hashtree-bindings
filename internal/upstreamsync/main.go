// Command upstreamsync regenerates the Go and assembly sources this module
// takes from the upstream hashtree checkout (the `hashtree` submodule).
//
// Upstream ships its native code as .syso objects next to the Go files. This
// module links prebuilt static libraries through cgo instead, which is the only
// intended difference. To keep it that way, no upstream-derived file is edited
// by hand: this tool copies them out of the submodule and rewrites nothing but
// the build constraint of the platform-specific files (they must only build
// where a static library is linked).
//
// Usage:
//
//	go run ./internal/upstreamsync          # regenerate the files
//	go run ./internal/upstreamsync -check   # fail if they are out of date
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/build/constraint"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// nativeConstraint selects the builds that link a static library from lib/.
// It must list exactly the platforms that have an internal/cgo/cgo_<os>_<arch>.go file.
// Every other build uses the pure-Go fallback (fallback.go).
const nativeConstraint = "cgo && ((linux && (amd64 || arm64 || riscv64)) || " +
	"(darwin && arm64) || (windows && amd64))"

// platform is a GOOS/GOARCH pair.
type platform struct {
	os   string
	arch string
}

// nativePlatforms are the platforms nativeConstraint selects (with cgo).
var nativePlatforms = []platform{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"linux", "riscv64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
}

// mode describes what happens to an upstream source file.
type mode int

const (
	// modeVerbatim copies the file byte for byte.
	modeVerbatim mode = iota
	// modeGated copies the file and narrows its build constraint to the
	// platforms in nativePlatforms. The rest of the file is left untouched.
	modeGated
	// modeIgnored skips the file. See the comment at each entry for why.
	modeIgnored
)

// upstreamFiles lists every .go and .s file in the upstream repository root.
// The tool fails on a file that is not listed, so a file added upstream cannot
// be missed silently.
var upstreamFiles = map[string]mode{
	"bindings.go":         modeVerbatim,
	"bindings_test.go":    modeVerbatim,
	"sha256_1_generic.go": modeVerbatim,

	"bindings_amd64.go":       modeGated,
	"bindings_arm64.go":       modeGated,
	"bindings_riscv64.go":     modeGated,
	"wrapper_arm64.s":         modeGated,
	"wrapper_linux_amd64.s":   modeGated,
	"wrapper_linux_riscv64.s": modeGated,
	"wrapper_windows_amd64.s": modeGated,

	// Upstream's pure-Go fallback for darwin/amd64 (an amd64-only stub).
	// fallback.go covers that platform together with all other builds
	// without a static library.
	"bindings_darwin_amd64.go": modeIgnored,
	"wrapper_darwin_amd64.s":   modeIgnored,

	// Upstream CI helper that forces the NEON path with `-tags neonforce`.
	// TestForcedPaths in hash_arm64_test.go already runs that path.
	"force_neon_arm64_test.go": modeIgnored,
}

// packagingConstraints are the hand-written files whose build constraint has to
// stay consistent with nativeConstraint, so that exactly one definition of
// HashtreeHash and supportedCPU exists in every build.
var packagingConstraints = map[string]string{
	"link.go":     nativeConstraint,
	"fallback.go": "!(" + nativeConstraint + ")",
}

func main() {
	upstreamDir := flag.String("upstream", "hashtree", "path to the upstream hashtree checkout")
	outDir := flag.String("out", ".", "directory of the bindings package")
	check := flag.Bool("check", false, "do not write; exit 1 if any file is out of date")

	flag.Parse()

	if err := run(*upstreamDir, *outDir, *check); err != nil {
		fmt.Fprintln(os.Stderr, "upstreamsync:", err)
		os.Exit(1)
	}
}

func run(upstreamDir, outDir string, check bool) error {
	names, err := listSources(upstreamDir)
	if err != nil {
		return err
	}

	if err := verifyPackaging(outDir); err != nil {
		return err
	}

	stale := make([]string, 0, len(names))

	for _, name := range names {
		if upstreamFiles[name] == modeIgnored {
			continue
		}

		want, err := generate(upstreamDir, name)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		target := filepath.Join(outDir, name)

		have, err := os.ReadFile(target)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read %s: %w", target, err)
		}

		if bytes.Equal(normalizeEOL(have), want) {
			continue
		}

		stale = append(stale, name)

		if check {
			continue
		}

		if err := os.WriteFile(target, want, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}

		fmt.Println("updated", name)
	}

	if check && len(stale) > 0 {
		return fmt.Errorf("out of date with %s: %s (run `go generate ./...`)",
			upstreamDir, strings.Join(stale, ", "))
	}

	return nil
}

// listSources returns the .go and .s files in the upstream repository root and
// verifies that they match upstreamFiles exactly.
func listSources(upstreamDir string) ([]string, error) {
	entries, err := os.ReadDir(upstreamDir)
	if err != nil {
		return nil, fmt.Errorf("read upstream checkout (is the submodule initialised?): %w", err)
	}

	names := make([]string, 0, len(upstreamFiles))
	seen := make(map[string]bool, len(upstreamFiles))

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || (!strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".s")) {
			continue
		}

		if _, known := upstreamFiles[name]; !known {
			return nil, fmt.Errorf("upstream has a source file this tool does not know: %s "+
				"(add it to upstreamFiles in internal/upstreamsync)", name)
		}

		names = append(names, name)
		seen[name] = true
	}

	for name, m := range upstreamFiles {
		// An ignored file that disappears upstream needs no attention.
		if !seen[name] && m != modeIgnored {
			return nil, fmt.Errorf("upstream no longer has %s "+
				"(update upstreamFiles in internal/upstreamsync)", name)
		}
	}

	sort.Strings(names)

	return names, nil
}

// generate returns the content this module should hold for an upstream file.
func generate(upstreamDir, name string) ([]byte, error) {
	src, err := os.ReadFile(filepath.Join(upstreamDir, name))
	if err != nil {
		return nil, fmt.Errorf("read upstream file: %w", err)
	}

	src = normalizeEOL(src)

	switch upstreamFiles[name] {
	case modeVerbatim:
		return src, nil
	case modeGated:
		return gate(name, src)
	default:
		return nil, errors.New("file is not generated")
	}
}

// gate replaces the build constraint of src with one that selects the native
// platforms the upstream constraint allows. Everything after the constraint
// lines is kept as is.
func gate(name string, src []byte) ([]byte, error) {
	upstream, body, err := splitConstraint(src)
	if err != nil {
		return nil, err
	}

	// A file name suffix like _linux_amd64 constrains the file as well.
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	matched := make([]platform, 0, len(nativePlatforms))

	for _, p := range nativePlatforms {
		if !nameAllows(stem, p) {
			continue
		}

		if upstream == nil || upstream.Eval(func(tag string) bool {
			return tag == p.os || tag == p.arch
		}) {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		return nil, errors.New("upstream constraint selects no platform we ship a library for")
	}

	expr, err := parseExpr(platformConstraint(matched))
	if err != nil {
		return nil, fmt.Errorf("build constraint: %w", err)
	}

	// Both constraint styles are written, as upstream does in its Go files.
	plusBuild, err := constraint.PlusBuildLines(expr)
	if err != nil {
		return nil, fmt.Errorf("build +build lines: %w", err)
	}

	var out bytes.Buffer

	fmt.Fprintf(&out, "//go:build %s\n", expr)

	for _, line := range plusBuild {
		out.WriteString(line + "\n")
	}

	out.WriteString("\n")
	out.Write(body)

	return out.Bytes(), nil
}

// nameAllows reports whether a file with the given name (without extension)
// builds on p according to Go's _GOOS, _GOARCH and _GOOS_GOARCH suffix rules,
// as far as the native platforms are concerned.
func nameAllows(stem string, p platform) bool {
	parts := strings.Split(stem, "_")
	last := parts[len(parts)-1]

	for _, other := range nativePlatforms {
		if last == other.arch {
			if last != p.arch {
				return false
			}

			if len(parts) < 2 {
				return true
			}

			prev := parts[len(parts)-2]

			for _, o := range nativePlatforms {
				if prev == o.os {
					return prev == p.os
				}
			}

			return true
		}

		if last == other.os {
			return last == p.os
		}
	}

	return true
}

// platformConstraint renders the shortest readable constraint selecting the
// given platforms with cgo enabled.
func platformConstraint(platforms []platform) string {
	sameOS, sameArch := true, true
	oses := make([]string, 0, len(platforms))
	arches := make([]string, 0, len(platforms))
	pairs := make([]string, 0, len(platforms))

	for _, p := range platforms {
		sameOS = sameOS && p.os == platforms[0].os
		sameArch = sameArch && p.arch == platforms[0].arch
		oses = append(oses, p.os)
		arches = append(arches, p.arch)
		pairs = append(pairs, "("+p.os+" && "+p.arch+")")
	}

	switch {
	case len(platforms) == 1:
		return "cgo && " + oses[0] + " && " + arches[0]
	case sameArch:
		return "cgo && (" + strings.Join(oses, " || ") + ") && " + arches[0]
	case sameOS:
		return "cgo && " + oses[0] + " && (" + strings.Join(arches, " || ") + ")"
	default:
		return "cgo && (" + strings.Join(pairs, " || ") + ")"
	}
}

// splitConstraint extracts the build constraint from the leading comment block
// of src. It returns the constraint (nil if the file has none) and the file
// without its constraint lines.
func splitConstraint(src []byte) (constraint.Expr, []byte, error) {
	var (
		goBuild   constraint.Expr
		plusBuild constraint.Expr
		kept      = make([]string, 0, 8)
		lines     = strings.SplitAfter(string(src), "\n")
		rest      = len(lines)
	)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Constraints are only valid before the first non-comment line.
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			rest = i

			break
		}

		switch {
		case constraint.IsGoBuild(trimmed):
			expr, err := constraint.Parse(trimmed)
			if err != nil {
				return nil, nil, fmt.Errorf("parse %q: %w", trimmed, err)
			}

			goBuild = expr
		case constraint.IsPlusBuild(trimmed):
			expr, err := constraint.Parse(trimmed)
			if err != nil {
				return nil, nil, fmt.Errorf("parse %q: %w", trimmed, err)
			}

			// Multiple "+build" lines are ANDed together.
			if plusBuild != nil {
				expr = &constraint.AndExpr{X: plusBuild, Y: expr}
			}

			plusBuild = expr
		default:
			kept = append(kept, line)
		}
	}

	// A "//go:build" line takes precedence over "+build" lines.
	upstream := goBuild
	if upstream == nil {
		upstream = plusBuild
	}

	header := strings.TrimLeft(strings.Join(kept, ""), "\n")
	body := header + strings.Join(lines[rest:], "")

	return upstream, []byte(body), nil
}

// verifyPackaging checks that the hand-written packaging files carry the
// constraint this tool gates the upstream files with.
func verifyPackaging(outDir string) error {
	for name, want := range packagingConstraints {
		wantExpr, err := parseExpr(want)
		if err != nil {
			return fmt.Errorf("parse constraint for %s: %w", name, err)
		}

		src, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			return fmt.Errorf("read packaging file: %w", err)
		}

		have, _, err := splitConstraint(src)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		if have == nil || have.String() != wantExpr.String() {
			return fmt.Errorf("%s must have the build constraint %q", name, wantExpr)
		}
	}

	return nil
}

// normalizeEOL undoes the CRLF conversion git applies on Windows checkouts, so
// the comparison does not depend on the platform the tool runs on.
func normalizeEOL(src []byte) []byte {
	return bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
}

func parseExpr(expr string) (constraint.Expr, error) {
	return constraint.Parse("//go:build " + expr)
}
