package merkle

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestNewTree(t *testing.T) {
	tests := []struct {
		name    string
		chunks  [][]byte
		wantErr bool
	}{
		{
			name: "Valid chunks",
			chunks: [][]byte{
				bytes.Repeat([]byte{1}, 32),
				bytes.Repeat([]byte{2}, 32),
				bytes.Repeat([]byte{3}, 32),
				bytes.Repeat([]byte{4}, 32),
			},
			wantErr: false,
		},
		{
			name:    "Empty chunks",
			chunks:  [][]byte{},
			wantErr: false,
		},
		{
			name: "Invalid chunk size",
			chunks: [][]byte{
				bytes.Repeat([]byte{1}, 32),
				bytes.Repeat([]byte{2}, 16), // Wrong size
				bytes.Repeat([]byte{3}, 32),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree(tt.chunks)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tree == nil {
				t.Errorf("NewTree() returned nil tree, want non-nil")
			}
		})
	}
}

func TestTreeRoot(t *testing.T) {
	// Test with a known tree structure and expected root
	// 4 chunks: [1, 2, 3, 4] with their 32-byte representations
	chunks := [][]byte{
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
		bytes.Repeat([]byte{3}, 32),
		bytes.Repeat([]byte{4}, 32),
	}

	tree, err := NewTree(chunks)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Manually calculate expected root for verification
	// Level 1: Hash(1|2), Hash(3|4)
	h1 := sha256.New()
	h1.Write(chunks[0])
	h1.Write(chunks[1])
	hash12 := h1.Sum(nil)

	h2 := sha256.New()
	h2.Write(chunks[2])
	h2.Write(chunks[3])
	hash34 := h2.Sum(nil)

	// Level 0: Root = Hash(Hash(1|2)|Hash(3|4))
	hRoot := sha256.New()
	hRoot.Write(hash12)
	hRoot.Write(hash34)
	expectedRoot := hRoot.Sum(nil)

	if !bytes.Equal(tree.Root(), expectedRoot) {
		t.Errorf("Tree.Root() = %x, want %x", tree.Root(), expectedRoot)
	}
}

func TestTreeComputeProof(t *testing.T) {
	chunks := [][]byte{
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
		bytes.Repeat([]byte{3}, 32),
		bytes.Repeat([]byte{4}, 32),
	}

	tree, err := NewTree(chunks)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	tests := []struct {
		name      string
		index     int
		wantErr   bool
		proofSize int
	}{
		{
			name:      "Valid index 0",
			index:     0,
			wantErr:   false,
			proofSize: 2, // log2(4) = 2 proof elements
		},
		{
			name:      "Valid index 3",
			index:     3,
			wantErr:   false,
			proofSize: 2,
		},
		{
			name:      "Out of bounds index",
			index:     4,
			wantErr:   true,
			proofSize: 0,
		},
		{
			name:      "Negative index",
			index:     -1,
			wantErr:   true,
			proofSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := tree.ComputeProof(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tree.ComputeProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(proof) != tt.proofSize {
				t.Errorf("Tree.ComputeProof() proof size = %d, want %d", len(proof), tt.proofSize)
			}
		})
	}
}

func TestTreeVerifyProof(t *testing.T) {
	// Create a tree with 8 chunks for more interesting verification tests
	chunks := make([][]byte, 8)
	for i := 0; i < 8; i++ {
		chunk := make([]byte, 32)
		chunk[0] = byte(i + 1) // Make each chunk unique but still 32 bytes
		chunks[i] = chunk
	}

	tree, err := NewTree(chunks)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	tests := []struct {
		name       string
		index      int
		value      []byte
		tamperWith string // describes what to tamper with for negative test cases
		want       bool
	}{
		{
			name:  "Valid proof for index 0",
			index: 0,
			value: chunks[0],
			want:  true,
		},
		{
			name:  "Valid proof for index 3",
			index: 3,
			value: chunks[3],
			want:  true,
		},
		{
			name:  "Valid proof for index 7",
			index: 7,
			value: chunks[7],
			want:  true,
		},
		{
			name:  "Invalid proof - wrong value",
			index: 2,
			value: bytes.Repeat([]byte{99}, 32), // Wrong value
			want:  false,
		},
		{
			name:  "Invalid proof - wrong index",
			index: 5,
			value: chunks[4], // Value doesn't match index
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute the proof for the test case
			var proof [][]byte
			var err error

			if tt.name == "Invalid proof - wrong index" {
				// Get proof for a different index to test index matching
				proof, err = tree.ComputeProof(4)
			} else {
				proof, err = tree.ComputeProof(tt.index)
			}

			if err != nil {
				t.Fatalf("Failed to compute proof: %v", err)
			}

			// Verify the proof
			got := tree.VerifyProof(tt.index, tt.value, proof)
			if got != tt.want {
				t.Errorf("Tree.VerifyProof() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{15, 16},
		{16, 16},
		{17, 32},
		{31, 32},
		{32, 32},
		{33, 64},
		{63, 64},
		{64, 64},
		{65, 128},
		{127, 128},
		{128, 128},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
			if got := nextPowerOfTwo(tt.n); got != tt.want {
				t.Errorf("nextPowerOfTwo(%d) = %d, want %d", tt.n, got, tt.want)
			}
		})
	}
}

func TestNonPowerOfTwoChunks(t *testing.T) {
	// Test with non-power-of-2 number of chunks
	chunks := [][]byte{
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
		bytes.Repeat([]byte{3}, 32),
		bytes.Repeat([]byte{4}, 32),
		bytes.Repeat([]byte{5}, 32),
	}

	tree, err := NewTree(chunks)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Verify that all indices have valid proofs
	for i := 0; i < len(chunks); i++ {
		proof, err := tree.ComputeProof(i)
		if err != nil {
			t.Fatalf("Failed to compute proof for index %d: %v", i, err)
		}

		if !tree.VerifyProof(i, chunks[i], proof) {
			t.Errorf("Proof verification failed for index %d", i)
		}
	}
}

func TestEmptyTree(t *testing.T) {
	// Test with empty chunk set
	tree, err := NewTree([][]byte{})
	if err != nil {
		t.Fatalf("Failed to create empty tree: %v", err)
	}

	// Empty tree should have a 32-byte zero value root
	expectedRoot := make([]byte, 32)
	if !bytes.Equal(tree.Root(), expectedRoot) {
		t.Errorf("Empty tree root = %x, want %x", tree.Root(), expectedRoot)
	}
}

func TestLargeTree(t *testing.T) {
	// Test with a larger tree (32 chunks)
	numChunks := 32
	chunks := make([][]byte, numChunks)

	for i := 0; i < numChunks; i++ {
		chunk := make([]byte, 32)
		chunk[0] = byte(i & 0xFF)
		chunk[1] = byte((i >> 8) & 0xFF)
		chunks[i] = chunk
	}

	tree, err := NewTree(chunks)
	if err != nil {
		t.Fatalf("Failed to create large tree: %v", err)
	}

	// Test a few random indices
	testIndices := []int{0, 5, 15, 27, 31}
	for _, idx := range testIndices {
		proof, err := tree.ComputeProof(idx)
		if err != nil {
			t.Fatalf("Failed to compute proof for index %d: %v", idx, err)
		}

		// Verify the proof
		if !tree.VerifyProof(idx, chunks[idx], proof) {
			t.Errorf("Proof verification failed for index %d in large tree", idx)
		}

		// Check that the proof has the expected size
		expectedProofSize := 5 // log2(32) = 5
		if len(proof) != expectedProofSize {
			t.Errorf("Proof size for index %d = %d, want %d", idx, len(proof), expectedProofSize)
		}
	}
}
