package beacon

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

// Constants from Ethereum spec
const (
	BytesPerChunk  = 32
	SecondsPerSlot = 12 // Ethereum consensus layer slot duration
)

// BlockHeader represents a simplified beacon block header
type BlockHeader struct {
	Slot          uint64
	ProposerIndex uint64
	ParentRoot    []byte
	StateRoot     []byte
	BodyRoot      []byte
}

// HeaderData represents the raw data received from the API
type HeaderData struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
	BlockRoot     string `json:"block_root"`
	Timestamp     int64  `json:"timestamp"`
}

// FromAPIResponse creates a BlockHeader from an API response data
func (b *BlockHeader) FromAPIResponse(data HeaderData) error {
	var err error

	// Convert slot and proposer_index to uint64
	if data.Slot != "" {
		b.Slot, err = strconv.ParseUint(data.Slot, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing slot: %w", err)
		}
	}

	if data.ProposerIndex != "" {
		b.ProposerIndex, err = strconv.ParseUint(data.ProposerIndex, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing proposer_index: %w", err)
		}
	}

	// Convert hex strings to bytes
	if data.ParentRoot != "" {
		b.ParentRoot, err = hex.DecodeString(trimHexPrefix(data.ParentRoot))
		if err != nil {
			return fmt.Errorf("decoding parent_root: %w", err)
		}
	} else {
		b.ParentRoot = make([]byte, 32)
	}

	if data.StateRoot != "" {
		b.StateRoot, err = hex.DecodeString(trimHexPrefix(data.StateRoot))
		if err != nil {
			return fmt.Errorf("decoding state_root: %w", err)
		}
	} else {
		b.StateRoot = make([]byte, 32)
	}

	if data.BodyRoot != "" {
		b.BodyRoot, err = hex.DecodeString(trimHexPrefix(data.BodyRoot))
		if err != nil {
			return fmt.Errorf("decoding body_root: %w", err)
		}
	} else {
		b.BodyRoot = make([]byte, 32)
	}

	return nil
}

// SerializeForMerkleization serializes the header fields for SSZ merkleization
func (b *BlockHeader) SerializeForMerkleization() [][]byte {
	// Create slice to hold serialized fields
	serialized := make([][]byte, 5)

	// Serialize uint64 fields in little-endian format and pad to 32 bytes
	slotBytes := make([]byte, 32)
	proposerBytes := make([]byte, 32)

	// Convert uint64 to little-endian bytes
	writeUint64LittleEndian(slotBytes, b.Slot)
	writeUint64LittleEndian(proposerBytes, b.ProposerIndex)

	// Assign serialized fields in correct order
	serialized[0] = slotBytes
	serialized[1] = proposerBytes
	serialized[2] = b.ParentRoot
	serialized[3] = b.StateRoot
	serialized[4] = b.BodyRoot

	return serialized
}

// Utility functions

// Helper function to write uint64 in little-endian format
func writeUint64LittleEndian(buf []byte, val uint64) {
	for i := 0; i < 8; i++ {
		buf[i] = byte(val >> (i * 8))
	}
}

// Helper function to trim "0x" prefix from hex strings
func trimHexPrefix(hexStr string) string {
	if len(hexStr) >= 2 && hexStr[0:2] == "0x" {
		return hexStr[2:]
	}
	return hexStr
}
