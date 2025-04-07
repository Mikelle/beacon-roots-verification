// Package merkle provides functions for creating and verifying Merkle trees
package merkle

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/bits"
)

// Tree represents a Merkle tree with methods for generating proofs
type Tree struct {
	chunks [][]byte
	root   []byte
}

// NewTree creates a new Merkle tree from a list of 32-byte chunks
func NewTree(chunks [][]byte) (*Tree, error) {
	// Ensure chunks are all 32 bytes
	for i, chunk := range chunks {
		if len(chunk) != 32 {
			return nil, fmt.Errorf("chunk %d has length %d, expected 32", i, len(chunk))
		}
	}

	tree := &Tree{
		chunks: make([][]byte, len(chunks)),
	}

	// Make a copy of the chunks to avoid modifying the original
	copy(tree.chunks, chunks)

	// Compute the root
	root, err := tree.merkleize()
	if err != nil {
		return nil, err
	}
	tree.root = root

	return tree, nil
}

// Root returns the Merkle root of the tree
func (t *Tree) Root() []byte {
	return t.root
}

// Chunks returns the original chunks used to create the tree
func (t *Tree) Chunks() [][]byte {
	return t.chunks
}

// ComputeProof generates a Merkle proof for a specific chunk index
func (t *Tree) ComputeProof(index int) ([][]byte, error) {
	if index < 0 || index >= len(t.chunks) {
		return nil, fmt.Errorf("index %d is out of range for chunks of length %d", index, len(t.chunks))
	}

	// Create a working copy of chunks
	chunks := make([][]byte, len(t.chunks))
	copy(chunks, t.chunks)

	// Ensure the number of chunks is a power of 2
	nextPow2 := nextPowerOfTwo(len(chunks))
	if len(chunks) < nextPow2 {
		// Pad with zero chunks
		zeroChunk := make([]byte, 32)
		for i := len(chunks); i < nextPow2; i++ {
			chunks = append(chunks, zeroChunk)
		}
	}

	proof := make([][]byte, 0, bits.Len(uint(nextPow2)))

	// Bottom layer consists of the chunks
	treeIndex := index
	layer := chunks

	// Build the proof by collecting siblings at each layer
	for len(layer) > 1 {
		siblingIndex := treeIndex ^ 1 // XOR with 1 to get the sibling index
		if siblingIndex < len(layer) {
			proof = append(proof, layer[siblingIndex])
		} else {
			proof = append(proof, make([]byte, 32)) // Zero chunk if no sibling
		}

		// Compute the next layer
		newLayer := make([][]byte, 0, (len(layer)+1)/2)
		for i := 0; i < len(layer); i += 2 {
			left := layer[i]
			right := make([]byte, 32)
			if i+1 < len(layer) {
				right = layer[i+1]
			}

			h := sha256.New()
			h.Write(left)
			h.Write(right)
			newLayer = append(newLayer, h.Sum(nil))
		}

		treeIndex = treeIndex / 2
		layer = newLayer
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof against the tree's root
func (t *Tree) VerifyProof(index int, value []byte, proof [][]byte) bool {
	current := value
	for i, sibling := range proof {
		h := sha256.New()
		if (index>>uint(i))&1 == 1 {
			h.Write(sibling)
			h.Write(current)
		} else {
			h.Write(current)
			h.Write(sibling)
		}
		current = h.Sum(nil)
	}
	return bytes.Equal(current, t.root)
}

// merkleize computes a merkle tree root from chunks
func (t *Tree) merkleize() ([]byte, error) {
	if len(t.chunks) == 0 {
		return make([]byte, 32), nil
	}

	// Ensure the number of chunks is a power of 2
	chunks := make([][]byte, len(t.chunks))
	copy(chunks, t.chunks)

	nextPow2 := nextPowerOfTwo(len(chunks))
	if len(chunks) < nextPow2 {
		// Pad with zero chunks
		zeroChunk := make([]byte, 32)
		for i := len(chunks); i < nextPow2; i++ {
			chunks = append(chunks, zeroChunk)
		}
	}

	// Bottom layer of the tree consists of the chunks
	tree := make([][]byte, len(chunks))
	copy(tree, chunks)

	// Compute parent nodes
	layerSize := len(tree)
	for layerSize > 1 {
		newLayer := make([][]byte, 0, (layerSize+1)/2)
		for i := 0; i < layerSize; i += 2 {
			left := tree[i]
			right := make([]byte, 32)
			if i+1 < layerSize {
				right = tree[i+1]
			}

			// Hash the concatenation of left and right
			h := sha256.New()
			h.Write(left)
			h.Write(right)
			newLayer = append(newLayer, h.Sum(nil))
		}
		tree = newLayer
		layerSize = len(tree)
	}

	return tree[0], nil
}

// nextPowerOfTwo returns the next power of 2 >= n
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	return 1 << uint(bits.Len(uint(n-1)))
}
