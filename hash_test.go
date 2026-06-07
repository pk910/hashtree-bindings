package hashtree

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

// referenceParent returns the expected Merkle parent digest for a 64-byte pair
// (two 32-byte chunks), computed with the standard library. crypto/sha256 is an
// independent implementation, so this is the authoritative known-answer oracle.
func referenceParent(left, right [32]byte) [32]byte {
	var buf [64]byte
	copy(buf[:32], left[:])
	copy(buf[32:], right[:])
	return sha256.Sum256(buf[:])
}

// genChunks deterministically generates 2*pairs chunks from a seed.
func genChunks(pairs int, seed int64) [][32]byte {
	r := rand.New(rand.NewSource(seed))
	chunks := make([][32]byte, 2*pairs)
	for i := range chunks {
		r.Read(chunks[i][:])
	}
	return chunks
}

// verifyHash is the shared core used by every size/path test. For the given
// chunks it checks Hash against the crypto/sha256 oracle, against the pure-Go
// reference, for determinism, and for HashByteSlice parity.
func verifyHash(t *testing.T, chunks [][32]byte) {
	t.Helper()
	pairs := len(chunks) / 2

	digests := make([][32]byte, pairs)
	if err := Hash(digests, chunks); err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	// 1. Independent oracle: crypto/sha256 over each 64-byte pair.
	for i := range pairs {
		want := referenceParent(chunks[2*i], chunks[2*i+1])
		if digests[i] != want {
			t.Fatalf("pair %d/%d: digest mismatch vs crypto/sha256\n got  %x\n want %x",
				i, pairs, digests[i], want)
		}
	}

	// 2. Cross-check against the pure-Go reference implementation.
	ref := make([][32]byte, pairs)
	sha256_1_generic(ref, chunks)
	for i := range pairs {
		if digests[i] != ref[i] {
			t.Fatalf("pair %d/%d: optimized path != pure-Go generic\n opt %x\n gen %x",
				i, pairs, digests[i], ref[i])
		}
	}

	// 3. Determinism: hashing the same input again yields the same output.
	again := make([][32]byte, pairs)
	if err := Hash(again, chunks); err != nil {
		t.Fatalf("Hash (2nd call) error: %v", err)
	}
	for i := range pairs {
		if again[i] != digests[i] {
			t.Fatalf("pair %d: non-deterministic output\n 1st %x\n 2nd %x", i, digests[i], again[i])
		}
	}

	// 4. HashByteSlice parity with Hash on the same data.
	if pairs > 0 {
		flatChunks := make([]byte, 0, pairs*64)
		for _, c := range chunks {
			flatChunks = append(flatChunks, c[:]...)
		}
		flatDigests := make([]byte, pairs*32)
		if err := HashByteSlice(flatDigests, flatChunks); err != nil {
			t.Fatalf("HashByteSlice error: %v", err)
		}
		for i := range pairs {
			if !bytes.Equal(flatDigests[32*i:32*i+32], digests[i][:]) {
				t.Fatalf("pair %d: HashByteSlice != Hash\n bs  %x\n h   %x",
					i, flatDigests[32*i:32*i+32], digests[i][:])
			}
		}
	}
}

// sizeSet covers boundaries around every SIMD lane width (x1/x2/x4/x8/x16) and
// their remainders, so the batch loop and its tail are both exercised.
var sizeSet = []int{
	0, 1, 2, 3, 4, 7, 8, 9, 15, 16, 17, 31, 32, 33,
	63, 64, 65, 127, 128, 129, 255, 256, 257, 1000,
}

// TestHashSizes hashes pseudo-random input at every boundary size and verifies
// each result against the oracle and the reference implementation.
func TestHashSizes(t *testing.T) {
	for _, n := range sizeSet {
		t.Run(fmt.Sprintf("pairs=%d", n), func(t *testing.T) {
			verifyHash(t, genChunks(n, int64(n)+1))
		})
	}
}

// TestHashLarge is the large stress test: 10000 pairs (20000 input chunks).
func TestHashLarge(t *testing.T) {
	verifyHash(t, genChunks(10000, 0x5eed))
}

// TestHashStructured exercises non-random inputs that tend to expose carry /
// state-reset bugs in SIMD implementations.
func TestHashStructured(t *testing.T) {
	mk := func(pairs int, fill func(i int) [32]byte) [][32]byte {
		c := make([][32]byte, 2*pairs)
		for i := range c {
			c[i] = fill(i)
		}
		return c
	}
	ones := func(int) [32]byte {
		var b [32]byte
		for j := range b {
			b[j] = 0xff
		}
		return b
	}
	seq := func(i int) [32]byte {
		var b [32]byte
		for j := range b {
			b[j] = byte(i + j)
		}
		return b
	}
	cases := map[string][][32]byte{
		"zeros": mk(64, func(int) [32]byte { return [32]byte{} }),
		"ones":  mk(64, ones),
		"seq":   mk(64, seq),
	}
	for name, chunks := range cases {
		t.Run(name, func(t *testing.T) { verifyHash(t, chunks) })
	}
}

// TestZeroHashChain validates iterative/chained use (Merkle zero-subtree hashes)
// against both the crypto/sha256 oracle and a hardcoded canonical anchor.
func TestZeroHashChain(t *testing.T) {
	// Canonical SSZ zerohash[1] = sha256(0^64); a well-known constant.
	const z1Hex = "f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b"

	var z [32]byte // zerohash[0] = 0
	for k := 1; k <= 40; k++ {
		d := make([][32]byte, 1)
		if err := Hash(d, [][32]byte{z, z}); err != nil {
			t.Fatalf("level %d: %v", k, err)
		}
		if want := referenceParent(z, z); d[0] != want {
			t.Fatalf("level %d: %x != oracle %x", k, d[0], want)
		}
		if k == 1 {
			if got := hex.EncodeToString(d[0][:]); got != z1Hex {
				t.Fatalf("zerohash[1] = %s, want %s", got, z1Hex)
			}
		}
		z = d[0]
	}
}

// TestErrors covers the validation/error paths of Hash and HashByteSlice.
func TestErrors(t *testing.T) {
	if err := Hash(make([][32]byte, 1), make([][32]byte, 3)); !errors.Is(err, ErrOddChunks) {
		t.Errorf("odd chunks: got %v, want ErrOddChunks", err)
	}
	if err := Hash(make([][32]byte, 1), make([][32]byte, 4)); !errors.Is(err, ErrNotEnoughDigests) {
		t.Errorf("short digests: got %v, want ErrNotEnoughDigests", err)
	}
	if err := Hash(nil, nil); err != nil {
		t.Errorf("empty: got %v, want nil", err)
	}
	if err := HashByteSlice(make([]byte, 32), make([]byte, 63)); !errors.Is(err, ErrChunksNotMultipleOf64) {
		t.Errorf("chunks not mult 64: got %v, want ErrChunksNotMultipleOf64", err)
	}
	if err := HashByteSlice(make([]byte, 31), make([]byte, 64)); !errors.Is(err, ErrDigestsNotMultipleOf32) {
		t.Errorf("digests not mult 32: got %v, want ErrDigestsNotMultipleOf32", err)
	}
	// 2 pairs (128 bytes) need 64 digest bytes; 32 is a valid multiple but too short.
	if err := HashByteSlice(make([]byte, 32), make([]byte, 128)); !errors.Is(err, ErrNotEnoughDigests) {
		t.Errorf("byteslice short digests: got %v, want ErrNotEnoughDigests", err)
	}
}

// forcedCase forces a specific implementation by temporarily overriding the
// detection state. apply() mutates the package vars and returns a restore func.
// supported reports whether the host CPU can actually execute this path's
// instructions (forcing an unsupported path would fault), so on a given runner
// only the paths it can run are exercised. The concrete cases per architecture
// live in the hash_<arch>_test.go files.
type forcedCase struct {
	name      string
	supported bool
	apply     func() (restore func())
}

// TestForcedPaths runs the functional core against every dispatch branch the
// host CPU can execute, by overriding detection regardless of what it would pick
// naturally. This is how the non-Linux runners (macOS, Windows) — which have no
// qemu CPU matrix — still get every supported optimized path exercised; under
// the Linux qemu matrix the unsupported branches simply skip per emulated -cpu.
func TestForcedPaths(t *testing.T) {
	for _, fc := range forcedCases {
		t.Run(fc.name, func(t *testing.T) {
			if !fc.supported {
				t.Skipf("host CPU does not support the %q path", fc.name)
			}
			restore := fc.apply()
			defer restore()
			for _, n := range []int{1, 2, 8, 17, 100, 1000} {
				verifyHash(t, genChunks(n, int64(n)*7+1))
			}
		})
	}
}

// TestActivePathMatchesEnv asserts that the implementation the library naturally
// selected matches HASHTREE_EXPECT_PATH. The CI matrix sets this per emulated
// -cpu so a -cpu that silently fails to apply (and thus exercises the wrong
// path) is caught instead of giving false coverage. Skipped when unset.
func TestActivePathMatchesEnv(t *testing.T) {
	want := os.Getenv("HASHTREE_EXPECT_PATH")
	if want == "" {
		t.Skip("HASHTREE_EXPECT_PATH not set")
	}
	if got := activePath(); got != want {
		t.Fatalf("active path = %q, want %q (CPU detection / -cpu mismatch)", got, want)
	}
	t.Logf("active path = %q (supportedCPU=%v)", want, supportedCPU)
}
